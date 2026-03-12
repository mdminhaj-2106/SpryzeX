package assembler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Preprocessor handles macros, .include, and .equ before C assembler sees the file
type Preprocessor struct {
	Defines map[string]string   // .equ / SET constants
	Macros  map[string][]string // .macro body lines
}

func NewPreprocessor() *Preprocessor {
	return &Preprocessor{
		Defines: make(map[string]string),
		Macros:  make(map[string][]string),
	}
}

var reEqu = regexp.MustCompile(`(?i)^\s*\.?(?:equ|set)\s+(\w+)\s*[,=]\s*(.+)$`)
var reMacroDef = regexp.MustCompile(`(?i)^\s*\.macro\s+(\w+)(?:\s+(.+))?$`)
var reMacroEnd = regexp.MustCompile(`(?i)^\s*\.endmacro\s*$`)
var reInclude = regexp.MustCompile(`(?i)^\s*\.include\s+"([^"]+)"`)
var reComment = regexp.MustCompile(`;.*$`)

// Process expands macros, includes, and constants in the source
// Returns the expanded source or an error
func (p *Preprocessor) Process(srcPath string) (string, []string, error) {
	return p.processFile(srcPath, 0)
}

func (p *Preprocessor) processFile(path string, depth int) (string, []string, error) {
	if depth > 16 {
		return "", nil, fmt.Errorf("include depth exceeded at %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var warnings []string
	var out []string
	var inMacro string
	var macroParams []string
	var macroBody []string

	for i, rawLine := range lines {
		line := rawLine
		_ = i

		// .include
		if m := reInclude.FindStringSubmatch(line); m != nil {
			incPath := m[1]
			if !filepath.IsAbs(incPath) {
				incPath = filepath.Join(filepath.Dir(path), incPath)
			}
			expanded, warns, err2 := p.processFile(incPath, depth+1)
			if err2 != nil {
				warnings = append(warnings, fmt.Sprintf("Warning: cannot include %s: %v", incPath, err2))
			} else {
				warnings = append(warnings, warns...)
				out = append(out, expanded)
			}
			continue
		}

		// .equ / SET
		if m := reEqu.FindStringSubmatch(line); m != nil {
			name := m[1]
			val := strings.TrimSpace(reComment.ReplaceAllString(m[2], ""))
			p.Defines[name] = val
			out = append(out, "; [equ] "+name+" = "+val)
			continue
		}

		// .macro start
		if m := reMacroDef.FindStringSubmatch(line); m != nil {
			inMacro = m[1]
			if m[2] != "" {
				macroParams = strings.Split(m[2], ",")
				for i := range macroParams {
					macroParams[i] = strings.TrimSpace(macroParams[i])
				}
			} else {
				macroParams = nil
			}
			macroBody = nil
			continue
		}

		// .endmacro
		if reMacroEnd.MatchString(line) {
			if inMacro != "" {
				p.Macros[inMacro] = macroBody
			}
			inMacro = ""
			macroBody = nil
			macroParams = nil
			continue
		}

		// Inside macro definition — collect body
		if inMacro != "" {
			macroBody = append(macroBody, line)
			continue
		}

		// Macro invocation?
		expanded := p.tryExpandMacro(line, macroParams)
		if expanded != "" {
			out = append(out, expanded)
			continue
		}

		// Constant substitution in line
		line = p.substituteConstants(line)

		out = append(out, line)
	}

	return strings.Join(out, "\n"), warnings, nil
}

func (p *Preprocessor) tryExpandMacro(line string, _ []string) string {
	trimmed := strings.TrimSpace(line)
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return ""
	}

	name := parts[0]
	body, ok := p.Macros[name]
	if !ok {
		return ""
	}

	// Substitute parameters
	callArgs := parts[1:]
	var result []string
	for _, bodyLine := range body {
		expanded := bodyLine
		for i, arg := range callArgs {
			paramName := fmt.Sprintf("\\%d", i+1)
			expanded = strings.ReplaceAll(expanded, paramName, arg)
		}
		result = append(result, expanded)
	}
	return strings.Join(result, "\n")
}

func (p *Preprocessor) substituteConstants(line string) string {
	// Find comment position — don't substitute inside comments
	commentIdx := strings.Index(line, ";")
	codePart := line
	commentPart := ""
	if commentIdx >= 0 {
		codePart = line[:commentIdx]
		commentPart = line[commentIdx:]
	}

	for name, val := range p.Defines {
		// Word boundary substitution
		codePart = replaceWholeWord(codePart, name, val)
	}
	return codePart + commentPart
}

func replaceWholeWord(s, old, new string) string {
	// Simple whole-word replacement
	result := s
	for {
		idx := strings.Index(result, old)
		if idx < 0 {
			break
		}
		before := idx > 0 && isWordChar(rune(result[idx-1]))
		after := idx+len(old) < len(result) && isWordChar(rune(result[idx+len(old)]))
		if before || after {
			// Skip this occurrence, advance past it
			break
		}
		result = result[:idx] + new + result[idx+len(old):]
	}
	return result
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// SymbolTable parses the listing file and extracts the symbol table
type Symbol struct {
	Name    string
	Address int
	Value   int
	Kind    string // "label", "equ", "data"
}

func ParseSymbolTable(listingPath string) []Symbol {
	data, err := os.ReadFile(listingPath)
	if err != nil {
		return nil
	}

	var symbols []Symbol
	inSymTable := false

	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "Symbol Table") || strings.Contains(line, "SYMBOL TABLE") {
			inSymTable = true
			continue
		}
		if inSymTable && strings.TrimSpace(line) == "" {
			continue
		}
		if inSymTable {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				val, err := strconv.ParseInt(parts[1], 0, 64)
				if err == nil {
					symbols = append(symbols, Symbol{
						Name:    name,
						Address: int(val),
						Kind:    "label",
					})
				}
			}
		}
	}
	return symbols
}

// HexDump creates a formatted hex dump of an object file
func HexDump(objPath string) []Line {
	data, err := os.ReadFile(objPath)
	if err != nil {
		return []Line{{Text: fmt.Sprintf("Cannot read %s", objPath), Kind: LineError}}
	}

	var lines []Line
	lines = append(lines, Line{
		Text: fmt.Sprintf("Hex dump: %s (%d bytes, %d words)", objPath, len(data), len(data)/4),
		Kind: LineInfo,
	})
	lines = append(lines, Line{Text: strings.Repeat("─", 60), Kind: LineSeparator})
	lines = append(lines, Line{
		Text: fmt.Sprintf("%-6s  %-8s  %-8s  %-8s  %-8s  ASCII", "ADDR", "+0", "+4", "+8", "+C"),
		Kind: LineInfo,
	})

	for i := 0; i < len(data); i += 16 {
		addr := i / 4
		var hexParts []string
		var ascii strings.Builder

		for j := 0; j < 16 && i+j < len(data); j += 4 {
			if i+j+3 < len(data) {
				word := uint32(data[i+j])<<24 | uint32(data[i+j+1])<<16 |
					uint32(data[i+j+2])<<8 | uint32(data[i+j+3])
				hexParts = append(hexParts, fmt.Sprintf("%08X", word))
			}
		}
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b < 127 {
				ascii.WriteByte(b)
			} else {
				ascii.WriteByte('.')
			}
		}

		line := fmt.Sprintf("%-6d  %-8s  %-8s  %-8s  %-8s  %s",
			addr,
			getOrEmpty(hexParts, 0),
			getOrEmpty(hexParts, 1),
			getOrEmpty(hexParts, 2),
			getOrEmpty(hexParts, 3),
			ascii.String(),
		)
		lines = append(lines, Line{Text: line, Kind: LineNormal})
	}
	return lines
}

func getOrEmpty(parts []string, i int) string {
	if i < len(parts) {
		return parts[i]
	}
	return "        "
}
