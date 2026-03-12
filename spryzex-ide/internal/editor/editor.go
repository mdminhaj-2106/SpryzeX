package editor

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"spryzex-ide/internal/theme"
)

// Mode represents vim-like editor mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeCommand
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeCommand:
		return "COMMAND"
	}
	return "NORMAL"
}

// Editor holds the full editor state
type Editor struct {
	Lines     []string
	CursorRow int
	CursorCol int
	Mode      Mode
	FilePath  string
	Dirty     bool
	ScrollRow int // top visible line
	ScrollCol int // left visible col

	// Visual mode
	VisualStartRow int
	VisualStartCol int

	// Command mode
	CommandBuf string
	CmdMsg     string

	// Search
	SearchQuery   string
	SearchMatches [][2]int // [row, col] pairs
	SearchIdx     int

	// Undo stack (simple: store full line snapshots)
	UndoStack []UndoState
	UndoIdx   int

	// Diagnostics
	ErrorLines map[int]string // line -> error message

	width  int
	height int
}

type UndoState struct {
	Lines     []string
	CursorRow int
	CursorCol int
}

// New creates a new editor
func New(w, h int) *Editor {
	return &Editor{
		Lines:      []string{""},
		width:      w,
		height:     h,
		ErrorLines: make(map[int]string),
	}
}

// SetSize updates the editor viewport size
func (e *Editor) SetSize(w, h int) {
	e.width = w
	e.height = h
	e.clampScroll()
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			e.Lines = []string{""}
			e.FilePath = path
			e.Dirty = false
			return nil
		}
		return err
	}
	e.FilePath = path
	content := string(data)
	if content == "" {
		e.Lines = []string{""}
	} else {
		lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
		// Remove trailing empty line if file ends in newline
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		if len(lines) == 0 {
			lines = []string{""}
		}
		e.Lines = lines
	}
	e.CursorRow = 0
	e.CursorCol = 0
	e.Dirty = false
	return nil
}

// SaveFile saves editor content to disk
func (e *Editor) SaveFile() error {
	if e.FilePath == "" {
		return fmt.Errorf("no file path set")
	}
	content := strings.Join(e.Lines, "\n") + "\n"
	err := os.WriteFile(e.FilePath, []byte(content), 0644)
	if err == nil {
		e.Dirty = false
	}
	return err
}

// HandleKey processes a keypress in the appropriate mode
func (e *Editor) HandleKey(key string) (cmd string) {
	switch e.Mode {
	case ModeNormal:
		return e.handleNormal(key)
	case ModeInsert:
		return e.handleInsert(key)
	case ModeVisual:
		return e.handleVisual(key)
	case ModeCommand:
		return e.handleCommand(key)
	}
	return ""
}

func (e *Editor) handleNormal(key string) string {
	switch key {
	// Mode switches
	case "i":
		e.Mode = ModeInsert
	case "I":
		e.Mode = ModeInsert
		e.CursorCol = 0
	case "a":
		e.Mode = ModeInsert
		e.CursorCol = min(e.CursorCol+1, len(e.currentLine()))
	case "A":
		e.Mode = ModeInsert
		e.CursorCol = len(e.currentLine())
	case "o":
		e.pushUndo()
		e.CursorRow++
		e.Lines = append(e.Lines[:e.CursorRow], append([]string{""}, e.Lines[e.CursorRow:]...)...)
		e.CursorCol = 0
		e.Mode = ModeInsert
		e.Dirty = true
	case "O":
		e.pushUndo()
		e.Lines = append(e.Lines[:e.CursorRow], append([]string{""}, e.Lines[e.CursorRow:]...)...)
		e.CursorCol = 0
		e.Mode = ModeInsert
		e.Dirty = true
	case "v":
		e.Mode = ModeVisual
		e.VisualStartRow = e.CursorRow
		e.VisualStartCol = e.CursorCol
	case ":":
		e.Mode = ModeCommand
		e.CommandBuf = ""

	// Movement
	case "h", "left":
		e.moveLeft()
	case "l", "right":
		e.moveRight()
	case "j", "down":
		e.moveDown()
	case "k", "up":
		e.moveUp()
	case "0":
		e.CursorCol = 0
	case "$":
		e.CursorCol = max(0, len(e.currentLine())-1)
	case "^":
		e.CursorCol = e.firstNonSpace()
	case "g":
		// gg handled as two keypresses — simplified: just go to line 1
		e.CursorRow = 0
		e.CursorCol = 0
	case "G":
		e.CursorRow = len(e.Lines) - 1
		e.CursorCol = 0
	case "ctrl+f", "pgdown":
		e.CursorRow = min(e.CursorRow+e.height/2, len(e.Lines)-1)
	case "ctrl+b", "pgup":
		e.CursorRow = max(e.CursorRow-e.height/2, 0)
	case "ctrl+d":
		e.CursorRow = min(e.CursorRow+e.height/4, len(e.Lines)-1)
	case "ctrl+u":
		e.CursorRow = max(e.CursorRow-e.height/4, 0)
	case "w":
		e.wordForward()
	case "b":
		e.wordBackward()
	case "e":
		e.wordEnd()

	// Editing
	case "x":
		e.deleteChar()
	case "d":
		// dd — delete line
		e.pushUndo()
		e.deleteLine()
	case "D":
		e.pushUndo()
		line := e.currentLine()
		e.Lines[e.CursorRow] = line[:e.CursorCol]
		e.Dirty = true
	case "C":
		e.pushUndo()
		line := e.currentLine()
		e.Lines[e.CursorRow] = line[:e.CursorCol]
		e.Mode = ModeInsert
		e.Dirty = true
	case "cc":
		e.pushUndo()
		e.Lines[e.CursorRow] = ""
		e.CursorCol = 0
		e.Mode = ModeInsert
		e.Dirty = true
	case "u":
		e.undo()
	case "ctrl+r":
		e.redo()
	case "p":
		// paste from kill buffer (simplified: insert blank line)

	// Search
	case "/":
		e.Mode = ModeCommand
		e.CommandBuf = "/"

	// Actions
	case "B":
		return "build"
	case "R":
		return "run"
	case "S":
		_ = e.SaveFile()
		return "saved"
	}

	e.clampCursor()
	e.clampScroll()
	return ""
}

func (e *Editor) handleInsert(key string) string {
	switch key {
	case "esc":
		e.Mode = ModeNormal
		e.CursorCol = max(0, e.CursorCol-1)
		e.clampCursor()
	case "ctrl+c":
		e.Mode = ModeNormal
	case "backspace":
		e.backspace()
	case "enter":
		e.newline()
	case "tab":
		e.insertAt("    ") // 4 spaces
	default:
		if len(key) == 1 && key[0] >= 32 {
			e.insertAt(key)
		}
	}
	e.clampScroll()
	return ""
}

func (e *Editor) handleVisual(key string) string {
	switch key {
	case "esc":
		e.Mode = ModeNormal
	case "h", "left":
		e.moveLeft()
	case "l", "right":
		e.moveRight()
	case "j", "down":
		e.moveDown()
	case "k", "up":
		e.moveUp()
	case "d", "x":
		e.deleteVisualSelection()
		e.Mode = ModeNormal
	case "y":
		// yank selection — simplified
		e.Mode = ModeNormal
	}
	e.clampCursor()
	e.clampScroll()
	return ""
}

func (e *Editor) handleCommand(key string) string {
	if e.CommandBuf == "/" {
		// Search mode
		switch key {
		case "esc":
			e.Mode = ModeNormal
			e.CommandBuf = ""
		case "enter":
			e.SearchQuery = e.CommandBuf[1:]
			e.doSearch()
			e.Mode = ModeNormal
			e.CommandBuf = ""
		case "backspace":
			if len(e.CommandBuf) > 1 {
				e.CommandBuf = e.CommandBuf[:len(e.CommandBuf)-1]
			}
		default:
			if len(key) == 1 {
				e.CommandBuf += key
			}
		}
		return ""
	}

	switch key {
	case "esc":
		e.Mode = ModeNormal
		e.CommandBuf = ""
	case "enter":
		cmd := strings.TrimSpace(e.CommandBuf)
		e.CommandBuf = ""
		e.Mode = ModeNormal
		return e.execCommand(cmd)
	case "backspace":
		if len(e.CommandBuf) > 0 {
			e.CommandBuf = e.CommandBuf[:len(e.CommandBuf)-1]
		}
	default:
		if len(key) == 1 || key == " " {
			e.CommandBuf += key
		}
	}
	return ""
}

func (e *Editor) execCommand(cmd string) string {
	switch {
	case cmd == "w":
		_ = e.SaveFile()
		e.CmdMsg = "File saved"
	case cmd == "q":
		return "quit"
	case cmd == "wq", cmd == "x":
		_ = e.SaveFile()
		return "quit"
	case cmd == "q!":
		return "quit!"
	case strings.HasPrefix(cmd, "w "):
		e.FilePath = strings.TrimPrefix(cmd, "w ")
		_ = e.SaveFile()
		e.CmdMsg = fmt.Sprintf("Saved to %s", e.FilePath)
	case strings.HasPrefix(cmd, "e "):
		path := strings.TrimPrefix(cmd, "e ")
		if err := e.LoadFile(path); err != nil {
			e.CmdMsg = fmt.Sprintf("Error: %v", err)
		} else {
			e.CmdMsg = fmt.Sprintf("Opened %s", path)
		}
	case cmd == "build", cmd == "b":
		return "build"
	case cmd == "run", cmd == "r":
		return "run"
	case strings.HasPrefix(cmd, "set "):
		// :set number, :set nonumber etc (placeholder)
	default:
		if cmd != "" {
			e.CmdMsg = fmt.Sprintf("Unknown command: %s", cmd)
		}
	}
	return ""
}

// View renders the editor content
func (e *Editor) View(focused bool) string {
	if e.height <= 0 || e.width <= 0 {
		return ""
	}

	gutterW := e.gutterWidth()
	contentW := e.contentWidth()

	var sb strings.Builder

	for row := e.ScrollRow; row < e.ScrollRow+e.height && row < len(e.Lines); row++ {
		lineNum := row + 1
		isCursorRow := row == e.CursorRow
		_, hasDiag := e.ErrorLines[row]

		// Gutter
		gutterStr := fmt.Sprintf("%*d ", gutterW-2, lineNum)
		if hasDiag {
			gutterStr += "!"
		} else {
			gutterStr += " "
		}
		switch {
		case hasDiag:
			sb.WriteString(lipgloss.NewStyle().
				Foreground(theme.ColorError).
				Background(theme.BgSurface).
				Bold(true).
				Render(gutterStr))
		case isCursorRow:
			sb.WriteString(theme.LineNumActiveStyle.Render(gutterStr))
		default:
			sb.WriteString(theme.LineNumStyle.Render(gutterStr))
		}

		// Separator
		sepStyle := lipgloss.NewStyle().Foreground(theme.BorderSubtle).Background(theme.BgSurface)
		if hasDiag {
			sepStyle = lipgloss.NewStyle().Foreground(theme.ColorError).Background(theme.BgSurface)
		} else if focused && isCursorRow {
			sepStyle = lipgloss.NewStyle().Foreground(theme.SpryzexRed).Background(theme.BgSurface)
		}
		sb.WriteString(sepStyle.Render("│"))

		// Line content with syntax highlighting
		rendered := e.renderLine(e.Lines[row], contentW, isCursorRow && focused)

		if isCursorRow && focused {
			sb.WriteString(theme.CursorLineStyle.Width(contentW).Render(rendered))
		} else {
			sb.WriteString(rendered)
		}

		sb.WriteRune('\n')
	}

	// Fill remaining lines with tilde
	rendered := sb.String()
	renderedLines := strings.Count(rendered, "\n")
	for i := renderedLines; i < e.height; i++ {
		sb.WriteString(theme.LineNumStyle.Render(strings.Repeat(" ", gutterW)))
		sb.WriteString(lipgloss.NewStyle().
			Foreground(theme.BorderSubtle).
			Background(theme.BgSurface).
			Render("│"))
		tildeStyle := lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.BgSurface)
		sb.WriteString(tildeStyle.Render("~"))
		if contentW > 1 {
			sb.WriteString(strings.Repeat(" ", contentW-1))
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}

// renderLine applies syntax highlighting to a single line
func (e *Editor) renderLine(line string, width int, isCursor bool) string {
	if len(line) == 0 {
		// Even empty cursor line needs cursor char
		if isCursor {
			cursorStyle := lipgloss.NewStyle().Background(theme.SpryzexGlow).Foreground(theme.BgDeep)
			if e.CursorCol-e.ScrollCol == 0 {
				return cursorStyle.Render(" ") + strings.Repeat(" ", width-1)
			}
		}
		return strings.Repeat(" ", width)
	}

	tokens := tokenize(line)
	var sb strings.Builder
	absCol := 0
	visibleCol := 0

	for _, tok := range tokens {
		var style lipgloss.Style
		switch tok.kind {
		case tokMnemonic:
			style = lipgloss.NewStyle().Foreground(theme.SynKeyword).Bold(true)
		case tokLabel:
			style = lipgloss.NewStyle().Foreground(theme.SynLabel).Bold(true)
		case tokNumber:
			style = lipgloss.NewStyle().Foreground(theme.SynNumber)
		case tokComment:
			style = lipgloss.NewStyle().Foreground(theme.SynComment).Italic(true)
		case tokDirective:
			style = lipgloss.NewStyle().Foreground(theme.SynDirectiv).Bold(true)
		case tokRegister:
			style = lipgloss.NewStyle().Foreground(theme.SynReg)
		default:
			style = lipgloss.NewStyle().Foreground(theme.TextPrimary)
		}

		// Apply cursor within token
		for _, ch := range []rune(tok.text) {
			if absCol < e.ScrollCol {
				absCol++
				continue
			}
			if visibleCol >= width {
				break
			}

			chStr := string(ch)

			if isCursor && absCol == e.CursorCol {
				// Draw cursor char
				cursorStyle := lipgloss.NewStyle().
					Background(theme.SpryzexGlow).
					Foreground(theme.BgDeep).
					Bold(true)
				sb.WriteString(cursorStyle.Render(chStr))
			} else {
				sb.WriteString(style.Render(chStr))
			}
			absCol++
			visibleCol++
		}
	}

	// Cursor at end of line
	lineLen := len([]rune(line))
	if isCursor && e.CursorCol >= e.ScrollCol && e.CursorCol == lineLen && visibleCol < width {
		cursorStyle := lipgloss.NewStyle().Background(theme.SpryzexGlow).Foreground(theme.BgDeep)
		sb.WriteString(cursorStyle.Render(" "))
		visibleCol++
	}

	result := sb.String()
	// Pad to width
	if visibleCol < width {
		result += strings.Repeat(" ", width-visibleCol)
	}
	return result
}

// ---- Tokenizer ----

type tokenKind int

const (
	tokNormal tokenKind = iota
	tokMnemonic
	tokLabel
	tokNumber
	tokComment
	tokDirective
	tokRegister
)

type token struct {
	text string
	kind tokenKind
}

var mnemonics = map[string]bool{
	"ldc": true, "adc": true, "ldl": true, "stl": true,
	"ldnl": true, "stnl": true, "add": true, "sub": true,
	"shl": true, "shr": true, "adj": true, "a2sp": true,
	"sp2a": true, "call": true, "return": true, "brz": true,
	"brlz": true, "br": true, "HALT": true, "halt": true,
	"out": true, "outc": true, "outnl": true, "data": true, "SET": true, "set": true,
	// bonus aliases
	"nop": true, "and": true, "or": true, "xor": true, "not": true,
	"push": true, "pop": true, "ld": true, "st": true,
}

func tokenize(line string) []token {
	var tokens []token

	// Comment: everything after ';'
	if idx := strings.Index(line, ";"); idx >= 0 {
		if idx > 0 {
			tokens = append(tokens, tokenize(line[:idx])...)
		}
		tokens = append(tokens, token{text: line[idx:], kind: tokComment})
		return tokens
	}

	// Directives: lines starting with '.' or '#'
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, ".") || strings.HasPrefix(trimmed, "#") {
		return []token{{text: line, kind: tokDirective}}
	}

	// Tokenize word by word
	words := tokenizeWords(line)
	for i, w := range words {
		kind := tokNormal
		text := w.text

		// Skip whitespace tokens
		if strings.TrimSpace(text) == "" {
			tokens = append(tokens, token{text: text, kind: tokNormal})
			continue
		}

		t := strings.TrimSpace(text)

		// Label: ends with ':'
		if strings.HasSuffix(t, ":") {
			kind = tokLabel
		} else if mnemonics[strings.ToLower(t)] {
			// Mnemonic (first meaningful token or after label)
			if i == 0 || (i > 0 && isOnlyLabels(words[:i])) {
				kind = tokMnemonic
			}
		} else if isNumber(t) {
			kind = tokNumber
		} else if strings.HasPrefix(t, "r") && len(t) > 1 && isDigits(t[1:]) {
			kind = tokRegister
		}

		tokens = append(tokens, token{text: text, kind: kind})
	}
	return tokens
}

type wordToken struct {
	text string
	isWS bool
}

func tokenizeWords(line string) []wordToken {
	var result []wordToken
	var cur strings.Builder
	wasSpace := false

	for _, ch := range line {
		isSpace := unicode.IsSpace(ch)
		if isSpace != wasSpace && cur.Len() > 0 {
			result = append(result, wordToken{text: cur.String(), isWS: wasSpace})
			cur.Reset()
		}
		cur.WriteRune(ch)
		wasSpace = isSpace
	}
	if cur.Len() > 0 {
		result = append(result, wordToken{text: cur.String(), isWS: wasSpace})
	}
	return result
}

func isOnlyLabels(words []wordToken) bool {
	for _, w := range words {
		if w.isWS {
			continue
		}
		if !strings.HasSuffix(strings.TrimSpace(w.text), ":") {
			return false
		}
	}
	return true
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		for _, c := range s[2:] {
			if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
				return false
			}
		}
		return len(s) > 2
	}
	return isDigits(s)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ---- Movement ----

func (e *Editor) moveUp() {
	if e.CursorRow > 0 {
		e.CursorRow--
	}
}

func (e *Editor) moveDown() {
	if e.CursorRow < len(e.Lines)-1 {
		e.CursorRow++
	}
}

func (e *Editor) moveLeft() {
	if e.CursorCol > 0 {
		e.CursorCol--
	} else if e.CursorRow > 0 {
		e.CursorRow--
		e.CursorCol = len(e.currentLine())
	}
}

func (e *Editor) moveRight() {
	line := e.currentLine()
	if e.CursorCol < len(line) {
		e.CursorCol++
	} else if e.CursorRow < len(e.Lines)-1 {
		e.CursorRow++
		e.CursorCol = 0
	}
}

func (e *Editor) wordForward() {
	line := e.currentLine()
	col := e.CursorCol
	// skip current word
	for col < len(line) && !unicode.IsSpace(rune(line[col])) {
		col++
	}
	// skip whitespace
	for col < len(line) && unicode.IsSpace(rune(line[col])) {
		col++
	}
	if col >= len(line) && e.CursorRow < len(e.Lines)-1 {
		e.CursorRow++
		e.CursorCol = 0
		return
	}
	e.CursorCol = col
}

func (e *Editor) wordBackward() {
	col := e.CursorCol
	line := e.currentLine()
	if col > 0 {
		col--
	}
	for col > 0 && unicode.IsSpace(rune(line[col])) {
		col--
	}
	for col > 0 && !unicode.IsSpace(rune(line[col-1])) {
		col--
	}
	e.CursorCol = col
}

func (e *Editor) wordEnd() {
	line := e.currentLine()
	col := e.CursorCol
	if col < len(line)-1 {
		col++
	}
	for col < len(line)-1 && unicode.IsSpace(rune(line[col])) {
		col++
	}
	for col < len(line)-1 && !unicode.IsSpace(rune(line[col+1])) {
		col++
	}
	e.CursorCol = col
}

func (e *Editor) firstNonSpace() int {
	line := e.currentLine()
	for i, ch := range line {
		if !unicode.IsSpace(ch) {
			return i
		}
	}
	return 0
}

// ---- Editing ----

func (e *Editor) insertAt(s string) {
	e.pushUndo()
	line := e.currentLine()
	col := e.CursorCol
	if col > len(line) {
		col = len(line)
	}
	e.Lines[e.CursorRow] = line[:col] + s + line[col:]
	e.CursorCol = col + len(s)
	e.Dirty = true
}

func (e *Editor) backspace() {
	e.pushUndo()
	line := e.currentLine()
	if e.CursorCol > 0 {
		e.Lines[e.CursorRow] = line[:e.CursorCol-1] + line[e.CursorCol:]
		e.CursorCol--
		e.Dirty = true
	} else if e.CursorRow > 0 {
		prevLine := e.Lines[e.CursorRow-1]
		e.CursorCol = len(prevLine)
		e.Lines[e.CursorRow-1] = prevLine + line
		e.Lines = append(e.Lines[:e.CursorRow], e.Lines[e.CursorRow+1:]...)
		e.CursorRow--
		e.Dirty = true
	}
}

func (e *Editor) newline() {
	e.pushUndo()
	line := e.currentLine()
	col := e.CursorCol
	if col > len(line) {
		col = len(line)
	}
	// Auto-indent: count leading spaces of current line
	indent := ""
	for _, ch := range line {
		if ch == ' ' {
			indent += " "
		} else if ch == '\t' {
			indent += "\t"
		} else {
			break
		}
	}
	before := line[:col]
	after := indent + line[col:]
	e.Lines[e.CursorRow] = before
	e.CursorRow++
	e.Lines = append(e.Lines[:e.CursorRow], append([]string{after}, e.Lines[e.CursorRow:]...)...)
	e.CursorCol = len(indent)
	e.Dirty = true
}

func (e *Editor) deleteChar() {
	e.pushUndo()
	line := e.currentLine()
	if e.CursorCol < len(line) {
		e.Lines[e.CursorRow] = line[:e.CursorCol] + line[e.CursorCol+1:]
		e.Dirty = true
	}
}

func (e *Editor) deleteLine() {
	if len(e.Lines) == 1 {
		e.Lines[0] = ""
		e.CursorCol = 0
		return
	}
	e.Lines = append(e.Lines[:e.CursorRow], e.Lines[e.CursorRow+1:]...)
	if e.CursorRow >= len(e.Lines) {
		e.CursorRow = len(e.Lines) - 1
	}
	e.CursorCol = 0
	e.Dirty = true
}

func (e *Editor) deleteVisualSelection() {
	e.pushUndo()
	r1, c1 := e.VisualStartRow, e.VisualStartCol
	r2, c2 := e.CursorRow, e.CursorCol
	if r1 > r2 || (r1 == r2 && c1 > c2) {
		r1, r2 = r2, r1
		c1, c2 = c2, c1
	}
	if r1 == r2 {
		line := e.Lines[r1]
		if c2 > len(line) {
			c2 = len(line)
		}
		e.Lines[r1] = line[:c1] + line[c2:]
	} else {
		// Multi-line delete
		first := e.Lines[r1][:c1]
		last := ""
		if r2 < len(e.Lines) {
			if c2 < len(e.Lines[r2]) {
				last = e.Lines[r2][c2:]
			}
		}
		e.Lines[r1] = first + last
		e.Lines = append(e.Lines[:r1+1], e.Lines[r2+1:]...)
	}
	e.CursorRow = r1
	e.CursorCol = c1
	e.Dirty = true
}

// ---- Search ----

func (e *Editor) doSearch() {
	e.SearchMatches = nil
	if e.SearchQuery == "" {
		return
	}
	q := strings.ToLower(e.SearchQuery)
	for row, line := range e.Lines {
		lower := strings.ToLower(line)
		col := 0
		for {
			idx := strings.Index(lower[col:], q)
			if idx < 0 {
				break
			}
			e.SearchMatches = append(e.SearchMatches, [2]int{row, col + idx})
			col += idx + len(q)
		}
	}
	if len(e.SearchMatches) > 0 {
		e.SearchIdx = 0
		e.CursorRow = e.SearchMatches[0][0]
		e.CursorCol = e.SearchMatches[0][1]
		e.clampScroll()
	}
}

func (e *Editor) SearchNext() {
	if len(e.SearchMatches) == 0 {
		return
	}
	e.SearchIdx = (e.SearchIdx + 1) % len(e.SearchMatches)
	e.CursorRow = e.SearchMatches[e.SearchIdx][0]
	e.CursorCol = e.SearchMatches[e.SearchIdx][1]
	e.clampScroll()
}

func (e *Editor) SearchPrev() {
	if len(e.SearchMatches) == 0 {
		return
	}
	e.SearchIdx = (e.SearchIdx - 1 + len(e.SearchMatches)) % len(e.SearchMatches)
	e.CursorRow = e.SearchMatches[e.SearchIdx][0]
	e.CursorCol = e.SearchMatches[e.SearchIdx][1]
	e.clampScroll()
}

// ---- Undo ----

func (e *Editor) pushUndo() {
	state := UndoState{
		Lines:     make([]string, len(e.Lines)),
		CursorRow: e.CursorRow,
		CursorCol: e.CursorCol,
	}
	copy(state.Lines, e.Lines)
	e.UndoStack = append(e.UndoStack[:e.UndoIdx], state)
	e.UndoIdx = len(e.UndoStack)
}

func (e *Editor) undo() {
	if e.UndoIdx <= 0 || len(e.UndoStack) == 0 {
		e.CmdMsg = "Already at oldest change"
		return
	}
	e.UndoIdx--
	state := e.UndoStack[e.UndoIdx]
	e.Lines = make([]string, len(state.Lines))
	copy(e.Lines, state.Lines)
	e.CursorRow = state.CursorRow
	e.CursorCol = state.CursorCol
	e.Dirty = true
}

func (e *Editor) redo() {
	if e.UndoIdx >= len(e.UndoStack) {
		e.CmdMsg = "Already at newest change"
		return
	}
	state := e.UndoStack[e.UndoIdx]
	e.UndoIdx++
	e.Lines = make([]string, len(state.Lines))
	copy(e.Lines, state.Lines)
	e.CursorRow = state.CursorRow
	e.CursorCol = state.CursorCol
	e.Dirty = true
}

// ---- Helpers ----

func (e *Editor) currentLine() string {
	if e.CursorRow >= len(e.Lines) {
		return ""
	}
	return e.Lines[e.CursorRow]
}

func (e *Editor) clampCursor() {
	if e.CursorRow >= len(e.Lines) {
		e.CursorRow = len(e.Lines) - 1
	}
	if e.CursorRow < 0 {
		e.CursorRow = 0
	}
	lineLen := len([]rune(e.currentLine()))
	if e.Mode == ModeNormal && lineLen > 0 {
		if e.CursorCol >= lineLen {
			e.CursorCol = lineLen - 1
		}
	} else if e.Mode == ModeInsert {
		if e.CursorCol > lineLen {
			e.CursorCol = lineLen
		}
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
}

func (e *Editor) clampScroll() {
	if e.height <= 0 {
		e.height = 1
	}
	if e.width <= 0 {
		e.width = 1
	}

	// Vertical
	if e.CursorRow < e.ScrollRow {
		e.ScrollRow = e.CursorRow
	}
	if e.CursorRow >= e.ScrollRow+e.height {
		e.ScrollRow = e.CursorRow - e.height + 1
	}
	if e.ScrollRow < 0 {
		e.ScrollRow = 0
	}
	maxScrollRow := max(len(e.Lines)-e.height, 0)
	if e.ScrollRow > maxScrollRow {
		e.ScrollRow = maxScrollRow
	}

	// Horizontal
	if e.CursorCol < e.ScrollCol {
		e.ScrollCol = e.CursorCol
	}
	contentW := e.contentWidth()
	if e.CursorCol >= e.ScrollCol+contentW {
		e.ScrollCol = e.CursorCol - contentW + 1
	}
	maxScrollCol := max(len([]rune(e.currentLine()))-contentW+1, 0)
	if e.ScrollCol > maxScrollCol {
		e.ScrollCol = maxScrollCol
	}
	if e.ScrollCol < 0 {
		e.ScrollCol = 0
	}
}

func (e *Editor) LineCount() int {
	return len(e.Lines)
}

func (e *Editor) Position() (int, int) {
	return e.CursorRow + 1, e.CursorCol + 1
}

// SetDiagnostics updates error lines from assembler output
func (e *Editor) SetDiagnostics(errs map[int]string) {
	e.ErrorLines = errs
}

func visibleLen(s string) int {
	return len([]rune(s))
}

func (e *Editor) DiagnosticAt(row int) string {
	return e.ErrorLines[row]
}

func (e *Editor) ScrollBy(delta int) {
	if delta == 0 {
		return
	}
	maxScroll := max(len(e.Lines)-e.height, 0)
	e.ScrollRow += delta
	if e.ScrollRow < 0 {
		e.ScrollRow = 0
	}
	if e.ScrollRow > maxScroll {
		e.ScrollRow = maxScroll
	}
	if e.CursorRow < e.ScrollRow {
		e.CursorRow = e.ScrollRow
	}
	if e.CursorRow >= e.ScrollRow+e.height {
		e.CursorRow = e.ScrollRow + e.height - 1
	}
	e.clampCursor()
	e.clampScroll()
}

func (e *Editor) MoveToViewPosition(x, y int) {
	if len(e.Lines) == 0 {
		return
	}
	if y < 0 {
		y = 0
	}
	if y >= e.height {
		y = e.height - 1
	}
	row := e.ScrollRow + y
	if row >= len(e.Lines) {
		row = len(e.Lines) - 1
	}

	col := e.ScrollCol + max(0, x-e.gutterWidth()-1)
	lineLen := len([]rune(e.Lines[row]))
	if e.Mode == ModeNormal && lineLen > 0 && col >= lineLen {
		col = lineLen - 1
	}
	if col > lineLen {
		col = lineLen
	}

	e.CursorRow = row
	e.CursorCol = max(0, col)
	e.clampCursor()
	e.clampScroll()
}

func (e *Editor) gutterWidth() int {
	gutterW := len(fmt.Sprintf("%d", len(e.Lines))) + 2
	if gutterW < 5 {
		gutterW = 5
	}
	return gutterW
}

func (e *Editor) contentWidth() int {
	contentW := e.width - e.gutterWidth() - 1
	if contentW < 1 {
		contentW = 1
	}
	return contentW
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
