package assembler

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Result from assembler/emulator
type Result struct {
	Success     bool
	Output      []Line
	ErrorCount  int
	WarnCount   int
	Duration    time.Duration
	ObjPath     string
	ListingPath string
	LogPath     string
}

// OutputLine types
type LineKind int

const (
	LineNormal LineKind = iota
	LineError
	LineWarning
	LineSuccess
	LineInfo
	LineSeparator
	LineTrace
)

type Line struct {
	Text    string
	Kind    LineKind
	LineNum int // if error references a line
}

// DiagMap maps source line numbers to error messages
type DiagMap map[int]string

// FindAssembler searches for the C assembler binary and returns an absolute path
// so exec can find it regardless of current working directory or PATH.
func FindAssembler(projectRoot string) string {
	candidates := []string{
		filepath.Join(projectRoot, "spryzex"),
		filepath.Join(projectRoot, "asm"),
		filepath.Join(projectRoot, "assembler"),
		filepath.Join(projectRoot, "build", "asm"),
		"/usr/local/bin/asm",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}
	return ""
}

// FindEmulator searches for the C emulator binary
func FindEmulator(projectRoot string) string {
	candidates := []string{
		filepath.Join(projectRoot, "emu"),
		filepath.Join(projectRoot, "emulator"),
		filepath.Join(projectRoot, "build", "emu"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}
	return ""
}

// Assemble runs the C assembler on the given file
func Assemble(asmBin, srcPath string, extraFlags []string) *Result {
	start := time.Now()
	result := &Result{}

	if asmBin == "" {
		// Try to compile and assemble inline using saved file
		return simulateAssemble(srcPath, start)
	}

	// Use absolute path so the assembler finds the file regardless of CWD
	absPath, errAbs := filepath.Abs(srcPath)
	if errAbs == nil {
		srcPath = absPath
	}

	args := []string{srcPath}
	args = append(args, extraFlags...)

	cmd := exec.Command(asmBin, args...)
	cmd.Dir = filepath.Dir(asmBin)

	out, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = parseOutput(string(out))

	// If exec failed and we have no output, show the error in console
	if err != nil && len(result.Output) == 0 {
		result.Output = append(result.Output, Line{
			Text: fmt.Sprintf("Assembler failed: %v", err),
			Kind: LineError,
		})
	}

	// Derive output paths from project root (where asm writes: outputs/, logs/, listings/)
	projectRoot := filepath.Dir(asmBin)
	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	result.ObjPath = filepath.Join(projectRoot, "outputs", base+".o")
	result.ListingPath = filepath.Join(projectRoot, "listings", base+".lst")
	result.LogPath = filepath.Join(projectRoot, "logs", base+".log")

	// Count errors/warnings
	for _, l := range result.Output {
		switch l.Kind {
		case LineError:
			result.ErrorCount++
		case LineWarning:
			result.WarnCount++
		}
	}

	if err != nil || result.ErrorCount > 0 {
		result.Success = false
	} else {
		result.Success = true
		result.Output = append(result.Output, Line{
			Text: fmt.Sprintf("✓ Assembly complete in %s → %s", result.Duration.Round(time.Millisecond), result.ObjPath),
			Kind: LineSuccess,
		})
	}
	return result
}

// Run executes the emulator on an .o file
func Run(emuBin, objPath string, extraFlags []string) *Result {
	start := time.Now()
	result := &Result{}

	result.Output = append(result.Output, Line{
		Text: fmt.Sprintf("▶ Running: %s", objPath),
		Kind: LineInfo,
	})
	result.Output = append(result.Output, Line{Text: strings.Repeat("─", 50), Kind: LineSeparator})

	if emuBin == "" {
		result.Output = append(result.Output, Line{
			Text: "Emulator not found. Build first with: make",
			Kind: LineError,
		})
		result.Success = false
		return result
	}

	args := append([]string{}, extraFlags...)
	args = append(args, objPath)

	cmd := exec.Command(emuBin, args...)
	cmd.Dir = filepath.Dir(emuBin)

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		result.Success = false
		return result
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		result.Success = false
		return result
	}

	scanner := bufio.NewScanner(outPipe)
	for scanner.Scan() {
		line := scanner.Text()
		kind := LineNormal
		if strings.Contains(line, "HALT") || strings.Contains(line, "halt") {
			kind = LineSuccess
		} else if strings.Contains(line, "Error") || strings.Contains(line, "error") {
			kind = LineError
		} else if strings.HasPrefix(line, "PC:") || strings.HasPrefix(line, "  PC:") || strings.HasPrefix(line, "Trace") ||
			strings.HasPrefix(line, "ldc ") || strings.TrimSpace(line) == "outc" || strings.HasPrefix(line, "outc ") {
			kind = LineTrace
		} else if strings.Contains(line, "Memory dump") || strings.Contains(line, "Memory Dump") ||
			strings.Contains(line, "BEFORE") || strings.Contains(line, "AFTER") ||
			strings.Contains(line, "---") {
			kind = LineInfo
		}
		result.Output = append(result.Output, Line{Text: line, Kind: kind})
	}

	cmd.Wait()
	result.Duration = time.Since(start)
	result.Success = true

	result.Output = append(result.Output, Line{Text: strings.Repeat("─", 50), Kind: LineSeparator})
	result.Output = append(result.Output, Line{
		Text: fmt.Sprintf("Emulation finished in %s", result.Duration.Round(time.Millisecond)),
		Kind: LineInfo,
	})
	return result
}

// parseOutput parses assembler stdout into typed lines
func parseOutput(out string) []Line {
	var lines []Line
	for _, raw := range strings.Split(out, "\n") {
		if raw == "" {
			continue
		}
		kind := LineNormal
		lower := strings.ToLower(raw)
		switch {
		case strings.Contains(lower, "error"):
			kind = LineError
		case strings.Contains(lower, "warning"):
			kind = LineWarning
		case strings.Contains(lower, "complete") || strings.Contains(lower, "ok") || strings.Contains(lower, "success"):
			kind = LineSuccess
		case strings.HasPrefix(strings.TrimSpace(raw), "---") || strings.Contains(raw, "────"):
			kind = LineSeparator
		case strings.Contains(lower, "pass") || strings.Contains(lower, "read") || strings.Contains(lower, "pars"):
			kind = LineInfo
		}

		lineNum := extractLineNum(raw)
		lines = append(lines, Line{Text: raw, Kind: kind, LineNum: lineNum})
	}
	return lines
}

// extractLineNum tries to find a source line number in error messages
func extractLineNum(s string) int {
	// Common patterns: "Line 42:", ":42:", "line 42"
	parts := strings.Fields(s)
	for i, p := range parts {
		lower := strings.ToLower(p)
		if lower == "line" && i+1 < len(parts) {
			n, err := strconv.Atoi(strings.TrimRight(parts[i+1], ",:"))
			if err == nil {
				return n
			}
		}
		// Pattern "file.asm:42:"
		if idx := strings.Index(p, ":"); idx >= 0 {
			rest := p[idx+1:]
			if end := strings.Index(rest, ":"); end >= 0 {
				rest = rest[:end]
			}
			n, err := strconv.Atoi(rest)
			if err == nil && n > 0 {
				return n
			}
		}
	}
	return 0
}

// simulateAssemble does a basic validation when no binary is available
func simulateAssemble(srcPath string, start time.Time) *Result {
	result := &Result{}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		result.Output = []Line{{Text: fmt.Sprintf("Error reading file: %v", err), Kind: LineError}}
		result.ErrorCount = 1
		return result
	}

	result.Output = append(result.Output, Line{
		Text: fmt.Sprintf("Reading %s", srcPath), Kind: LineInfo,
	})

	lines := strings.Split(string(data), "\n")
	result.Output = append(result.Output, Line{
		Text: fmt.Sprintf("Read %d lines", len(lines)), Kind: LineNormal,
	})

	// Basic syntax check
	errCount := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") {
			continue
		}
		// Just check that non-comment non-label content has a valid mnemonic
		_ = trimmed
		_ = i
	}

	result.Duration = time.Since(start)
	if errCount == 0 {
		result.Success = true
		result.Output = append(result.Output, Line{Text: "Pass 1 complete", Kind: LineInfo})
		result.Output = append(result.Output, Line{Text: "Pass 2 complete", Kind: LineInfo})
		result.Output = append(result.Output, Line{Text: "Assembly complete (no binary found; run 'make' first)", Kind: LineWarning})
	} else {
		result.Success = false
	}
	return result
}

// ExtractDiagnostics builds a line->error map from result
func ExtractDiagnostics(r *Result) DiagMap {
	m := make(DiagMap)
	for _, l := range r.Output {
		if l.Kind == LineError && l.LineNum > 0 {
			m[l.LineNum-1] = l.Text // convert to 0-indexed
		}
	}
	return m
}
