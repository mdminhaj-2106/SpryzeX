package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"spryzex-ide/internal/assembler"
	"spryzex-ide/internal/editor"
	"spryzex-ide/internal/spryzex"
	"spryzex-ide/internal/theme"
)

// ---- Message types ----

type tickMsg struct{ t time.Time }
type buildDoneMsg struct{ result *assembler.Result }
type runDoneMsg struct{ result *assembler.Result }

// ---- Panel focus ----

type Focus int

const (
	FocusEditor Focus = iota
	FocusConsole
	FocusSpryzex
)

// ---- Output tabs ----

type OutputTab int

const (
	TabLive OutputTab = iota
	TabLog
	TabListing
	TabObj
)

func (t OutputTab) String() string {
	switch t {
	case TabLive:
		return "LIVE"
	case TabLog:
		return "LOG"
	case TabListing:
		return "LISTING"
	case TabObj:
		return "OBJ"
	}
	return "LIVE"
}

// ---- App model ----

type model struct {
	// Layout
	width  int
	height int

	// Components
	ed    *editor.Editor
	anim  *spryzex.Animator
	focus Focus

	// Build state
	building    bool
	running     bool
	lastResult  *assembler.Result
	activeTab   OutputTab
	consoleScrl int // scroll offset in console

	// Paths
	projectRoot string
	asmBin      string
	emuBin      string

	// Notifications
	notification string
	notifyUntil  time.Time

	// Status
	statusMsg string

	// Build stats
	buildCount int
	errCount   int
}

func initialModel(filePath string) model {
	m := model{
		anim:        spryzex.NewAnimator(30, 20),
		projectRoot: findProjectRoot(filePath),
	}

	// Create editor
	m.ed = editor.New(80, 30)

	// Load file
	if filePath != "" {
		if err := m.ed.LoadFile(filePath); err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		}
	} else {
		m.ed.FilePath = "untitled.asm"
		m.ed.Lines = defaultASM()
	}

	// Find binaries
	m.asmBin = assembler.FindAssembler(m.projectRoot)
	m.emuBin = assembler.FindEmulator(m.projectRoot)

	return m
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateEditorSize()

	case tickMsg:
		m.anim.Tick()
		return m, tickCmd()

	case buildDoneMsg:
		m.building = false
		m.lastResult = msg.result
		m.buildCount++
		if msg.result.Success {
			m.anim.SetState(spryzex.StateSuccess)
			m.errCount = 0
			m.notify("✓ Build succeeded!")
			m.statusMsg = fmt.Sprintf("Build OK in %s", msg.result.Duration.Round(time.Millisecond))
		} else {
			m.anim.SetState(spryzex.StateError)
			m.errCount = msg.result.ErrorCount
			m.notify(fmt.Sprintf("✗ %d error(s) — check LOG tab", msg.result.ErrorCount))
			m.statusMsg = fmt.Sprintf("Build FAILED — %d error(s)", msg.result.ErrorCount)
			// Update editor diagnostics
			m.ed.SetDiagnostics(assembler.ExtractDiagnostics(msg.result))
		}
		return m, nil

	case runDoneMsg:
		m.running = false
		m.lastResult = msg.result
		if msg.result.Success {
			m.anim.SetState(spryzex.StateSuccess)
			m.notify("Emulation complete")
		} else {
			m.anim.SetState(spryzex.StateError)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global shortcuts (always active)
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+w":
		// Cycle focus
		m.focus = (m.focus + 1) % 3
		return m, nil
	case "ctrl+]":
		m.activeTab = (m.activeTab + 1) % 4
		return m, nil
	case "ctrl+[":
		m.activeTab = (m.activeTab - 1 + 4) % 4
		return m, nil
	case "ctrl+b":
		return m, m.triggerBuild()
	case "ctrl+r":
		return m, m.triggerRun()
	case "ctrl+s":
		_ = m.ed.SaveFile()
		m.notify("Saved")
		return m, nil
	}

	// Console scrolling when console is focused
	if m.focus == FocusConsole {
		switch key {
		case "j", "down":
			m.consoleScrl++
		case "k", "up":
			if m.consoleScrl > 0 {
				m.consoleScrl--
			}
		case "g":
			m.consoleScrl = 0
		case "G":
			m.consoleScrl = 999999
		case "esc":
			m.focus = FocusEditor
		}
		return m, nil
	}

	// Editor key handling
	if m.focus == FocusEditor {
		cmd := m.ed.HandleKey(key)
		switch cmd {
		case "build":
			return m, m.triggerBuild()
		case "run":
			return m, m.triggerRun()
		case "saved":
			m.notify("Saved")
		case "quit", "quit!":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) triggerBuild() tea.Cmd {
	if m.building {
		return nil
	}
	_ = m.ed.SaveFile()
	m.building = true
	m.anim.SetState(spryzex.StateBuilding)
	m.activeTab = TabLive
	m.consoleScrl = 0
	m.statusMsg = "Building..."

	asmBin := m.asmBin
	emuBin := m.emuBin
	filePath := m.ed.FilePath
	_ = emuBin

	return func() tea.Msg {
		result := assembler.Assemble(asmBin, filePath, nil)
		return buildDoneMsg{result}
	}
}

func (m *model) triggerRun() tea.Cmd {
	if m.running {
		return nil
	}
	// First save
	_ = m.ed.SaveFile()

	// Build first if no obj
	if m.lastResult == nil || !m.lastResult.Success {
		return m.triggerBuild()
	}

	m.running = true
	m.anim.SetState(spryzex.StateRunning)
	m.activeTab = TabLive
	m.consoleScrl = 0
	m.statusMsg = "Running..."

	emuBin := m.emuBin
	objPath := m.lastResult.ObjPath

	return func() tea.Msg {
		result := assembler.Run(emuBin, objPath, nil)
		return runDoneMsg{result}
	}
}

func (m *model) notify(msg string) {
	m.notification = msg
	m.notifyUntil = time.Now().Add(3 * time.Second)
}

func (m *model) updateEditorSize() {
	// Layout: [spryzex 28cols | editor rest] top half, console bottom 1/3
	spryzexW := 30
	if m.width < 100 {
		spryzexW = 0 // hide spryzex on small terminals
	}
	edW := m.width - spryzexW - 2 // 2 for borders
	consoleH := m.height / 3
	edH := m.height - consoleH - 4 // 4 for title+status bars

	if edW < 20 {
		edW = 20
	}
	if edH < 5 {
		edH = 5
	}
	m.ed.SetSize(edW, edH)
}

// ---- View ----

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Compute layout
	spryzexW := 30
	if m.width < 100 {
		spryzexW = 0
	}
	edW := m.width - spryzexW
	consoleH := m.height / 3
	topH := m.height - consoleH - 2 // 2 for title+status

	// --- Title bar ---
	titleBar := m.renderTitleBar()

	// --- Top section: Spryzex | Editor ---
	var topSection string
	if spryzexW > 0 {
		spryzexPanel := m.renderSpryzexPanel(spryzexW, topH)
		edPanel := m.renderEditorPanel(edW, topH)
		topSection = lipgloss.JoinHorizontal(lipgloss.Top, spryzexPanel, edPanel)
	} else {
		topSection = m.renderEditorPanel(m.width, topH)
	}

	// --- Console ---
	console := m.renderConsole(m.width, consoleH)

	// --- Status bar ---
	statusBar := m.renderStatusBar()

	// --- Notification overlay ---
	page := lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		topSection,
		console,
		statusBar,
	)

	// Overlay notification
	if m.notification != "" && time.Now().Before(m.notifyUntil) {
		page = m.overlayNotification(page)
	} else {
		m.notification = ""
	}

	return page
}

func (m model) renderTitleBar() string {
	// Left: logo
	logo := lipgloss.NewStyle().
		Foreground(theme.SpryzexGlow).
		Bold(true).
		Render(" ◈ SPRYZEX IDE")

	// Center: filename
	fname := m.ed.FilePath
	if fname == "" {
		fname = "untitled.asm"
	}
	dirty := ""
	if m.ed.Dirty {
		dirty = " ●"
	}
	fileInfo := lipgloss.NewStyle().
		Foreground(theme.TextPrimary).
		Render(fmt.Sprintf("  %s%s  ", filepath.Base(fname), dirty))

	// Right: build status + time
	var statusPill string
	switch {
	case m.building:
		statusPill = theme.StatusBuildStyle.Render(" BUILDING ")
	case m.running:
		statusPill = theme.StatusBuildStyle.Render(" RUNNING ")
	case m.lastResult != nil && m.lastResult.Success:
		statusPill = theme.StatusOKStyle.Render(" BUILD OK ")
	case m.lastResult != nil && !m.lastResult.Success:
		statusPill = theme.StatusErrStyle.Render(fmt.Sprintf(" %d ERRORS ", m.lastResult.ErrorCount))
	default:
		statusPill = lipgloss.NewStyle().Foreground(theme.TextMuted).Render(" READY ")
	}

	timeStr := lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Render(time.Now().Format(" 15:04:05 "))

	// Combine with spacing
	rightSection := lipgloss.JoinHorizontal(lipgloss.Center, statusPill, timeStr)
	totalW := m.width

	leftW := lipgloss.Width(logo)
	rightW := lipgloss.Width(rightSection)
	midW := totalW - leftW - rightW
	if midW < 0 {
		midW = 0
	}

	center := lipgloss.NewStyle().Width(midW).Align(lipgloss.Center).
		Foreground(theme.TextSecond).Render(fileInfo)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, logo, center, rightSection)
	return lipgloss.NewStyle().
		Background(theme.BgDeep).
		Width(m.width).
		Render(bar) + "\n"
}

func (m model) renderSpryzexPanel(w, h int) string {
	spryzexH := h - 2 // subtract borders
	spryzexW := w - 2
	if spryzexH < 5 {
		spryzexH = 5
	}
	if spryzexW < 10 {
		spryzexW = 10
	}

	content := m.anim.Render(spryzexW, spryzexH)

	borderColor := theme.BorderSubtle
	if m.focus == FocusSpryzex {
		borderColor = theme.SpryzexRed
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(theme.BgSurface).
		Render(content)
}

func (m model) renderEditorPanel(w, h int) string {
	innerW := w - 2
	_ = h

	editorContent := m.ed.View(m.focus == FocusEditor)

	// Mode pill
	var modePill string
	switch m.ed.Mode {
	case editor.ModeInsert:
		modePill = theme.ModeInsertStyle.Render(" INSERT ")
	case editor.ModeVisual:
		modePill = theme.ModeVisualStyle.Render(" VISUAL ")
	case editor.ModeCommand:
		modePill = theme.ModeCommandStyle.Render(" COMMAND ")
	default:
		modePill = theme.ModeNormalStyle.Render(" NORMAL ")
	}

	row, col := m.ed.Position()
	posInfo := lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Render(fmt.Sprintf(" %d:%d ", row, col))

	// Command bar
	cmdBar := ""
	if m.ed.Mode == editor.ModeCommand {
		cmdBar = lipgloss.NewStyle().
			Width(innerW).
			Background(theme.BgOverlay).
			Foreground(theme.TextPrimary).
			Render(fmt.Sprintf(":%s", m.ed.CommandBuf))
	} else if m.ed.CmdMsg != "" {
		cmdBar = lipgloss.NewStyle().
			Width(innerW).
			Foreground(theme.TextMuted).
			Render(m.ed.CmdMsg)
	}

	titleRight := lipgloss.JoinHorizontal(lipgloss.Right, posInfo, modePill)
	titleW := innerW - lipgloss.Width(titleRight)
	if titleW < 0 {
		titleW = 0
	}
	edTitle := lipgloss.NewStyle().Width(titleW).Foreground(theme.TextMuted).
		Render(" CODE ")
	header := lipgloss.JoinHorizontal(lipgloss.Right, edTitle, titleRight)

	var content strings.Builder
	content.WriteString(header)
	content.WriteRune('\n')
	content.WriteString(editorContent)
	if cmdBar != "" {
		content.WriteString("\n")
		content.WriteString(cmdBar)
	}

	borderColor := theme.BorderSubtle
	if m.focus == FocusEditor {
		borderColor = theme.PhobosBlue
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(theme.BgSurface).
		Render(content.String())
}

func (m model) renderConsole(w, h int) string {
	innerW := w - 2

	// Tab bar
	tabs := []OutputTab{TabLive, TabLog, TabListing, TabObj}
	var tabBar strings.Builder
	for _, tab := range tabs {
		if tab == m.activeTab {
			tabBar.WriteString(theme.TabActiveStyle.Render(tab.String()))
		} else {
			tabBar.WriteString(theme.TabStyle.Render(tab.String()))
		}
	}

	// Stats badge
	if m.lastResult != nil {
		var badge string
		if m.lastResult.ErrorCount > 0 {
			badge = lipgloss.NewStyle().
				Background(theme.ColorError).Foreground(theme.BgDeep).
				Padding(0, 1).Bold(true).
				Render(fmt.Sprintf(" %d errors ", m.lastResult.ErrorCount))
		}
		if m.lastResult.WarnCount > 0 {
			warnBadge := lipgloss.NewStyle().
				Background(theme.ColorWarn).Foreground(theme.BgDeep).
				Padding(0, 1).Bold(true).
				Render(fmt.Sprintf(" %d warns ", m.lastResult.WarnCount))
			badge += warnBadge
		}
		if badge != "" {
			tabBarStr := tabBar.String()
			gap := innerW - lipgloss.Width(tabBarStr) - lipgloss.Width(badge)
			if gap > 0 {
				tabBarStr += strings.Repeat(" ", gap)
			}
			tabBarStr += badge
			tabBar.Reset()
			tabBar.WriteString(tabBarStr)
		}
	}

	tabLine := lipgloss.NewStyle().
		Width(innerW).
		Background(theme.BgMuted).
		Render(tabBar.String())

	// Content
	contentH := h - 4 // tabs + border
	if contentH < 1 {
		contentH = 1
	}
	consoleContent := m.renderConsoleContent(innerW, contentH)

	borderColor := theme.BorderSubtle
	if m.focus == FocusConsole {
		borderColor = theme.AuroraGreen
	}

	consoleLabel := lipgloss.NewStyle().Foreground(theme.AuroraGreen).Bold(true).Render(" CONSOLE ")

	full := lipgloss.JoinVertical(lipgloss.Left, tabLine, consoleContent)

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(theme.BgSurface).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		Render(fmt.Sprintf("%s\n%s", consoleLabel, full))
}

func (m model) renderConsoleContent(w, h int) string {
	lines := m.getConsoleLines()

	// Scroll
	total := len(lines)
	start := m.consoleScrl
	if start > total-h {
		start = total - h
	}
	if start < 0 {
		start = 0
	}

	var sb strings.Builder
	shown := 0
	for i := start; i < total && shown < h; i++ {
		l := lines[i]
		var style lipgloss.Style
		switch l.Kind {
		case assembler.LineError:
			style = lipgloss.NewStyle().Foreground(theme.ColorError)
		case assembler.LineWarning:
			style = lipgloss.NewStyle().Foreground(theme.ColorWarn)
		case assembler.LineSuccess:
			style = lipgloss.NewStyle().Foreground(theme.ColorOK).Bold(true)
		case assembler.LineInfo:
			style = lipgloss.NewStyle().Foreground(theme.PhobosBlue)
		case assembler.LineSeparator:
			style = lipgloss.NewStyle().Foreground(theme.TextMuted)
		case assembler.LineTrace:
			style = lipgloss.NewStyle().Foreground(theme.NebulaPurp)
		default:
			style = lipgloss.NewStyle().Foreground(theme.TextPrimary)
		}

		rendered := style.Width(w).Render(truncate(l.Text, w))
		sb.WriteString(rendered)
		sb.WriteRune('\n')
		shown++
	}

	// Fill empty lines
	for shown < h {
		sb.WriteString(strings.Repeat(" ", w))
		sb.WriteRune('\n')
		shown++
	}

	return sb.String()
}

func (m model) getConsoleLines() []assembler.Line {
	if m.lastResult == nil {
		if m.building {
			return []assembler.Line{
				{Text: "Building...", Kind: assembler.LineInfo},
			}
		}
		return []assembler.Line{
			{Text: "No output yet. Press Ctrl+B to build, Ctrl+R to run.", Kind: assembler.LineInfo},
			{Text: "  :build  or  B  in normal mode to assemble", Kind: assembler.LineNormal},
			{Text: "  :run   or  R  to execute emulator", Kind: assembler.LineNormal},
		}
	}

	switch m.activeTab {
	case TabLog:
		return m.loadFileLines(m.lastResult.LogPath)
	case TabListing:
		return m.loadFileLines(m.lastResult.ListingPath)
	case TabObj:
		return m.renderObjLines()
	}
	return m.lastResult.Output
}

func (m model) loadFileLines(path string) []assembler.Line {
	if path == "" {
		return []assembler.Line{{Text: "No file available", Kind: assembler.LineInfo}}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []assembler.Line{{Text: fmt.Sprintf("Cannot read %s", path), Kind: assembler.LineError}}
	}
	var lines []assembler.Line
	for _, l := range strings.Split(string(data), "\n") {
		lines = append(lines, assembler.Line{Text: l, Kind: assembler.LineNormal})
	}
	return lines
}

func (m model) renderObjLines() []assembler.Line {
	if m.lastResult == nil || m.lastResult.ObjPath == "" {
		return []assembler.Line{{Text: "No .o file available yet", Kind: assembler.LineInfo}}
	}
	data, err := os.ReadFile(m.lastResult.ObjPath)
	if err != nil {
		return []assembler.Line{{Text: fmt.Sprintf("Cannot read %s: %v", m.lastResult.ObjPath, err), Kind: assembler.LineError}}
	}

	var lines []assembler.Line
	lines = append(lines, assembler.Line{
		Text: fmt.Sprintf("Object file: %s (%d bytes)", m.lastResult.ObjPath, len(data)),
		Kind: assembler.LineInfo,
	})
	lines = append(lines, assembler.Line{Text: strings.Repeat("─", 60), Kind: assembler.LineSeparator})
	lines = append(lines, assembler.Line{
		Text: fmt.Sprintf("%-6s  %-10s  %-12s  %s", "ADDR", "HEX", "MNEMONIC", "OPERAND"),
		Kind: assembler.LineInfo,
	})

	// Disassemble: each word is 4 bytes big-endian
	for i := 0; i+3 < len(data); i += 4 {
		addr := i / 4
		word := uint32(data[i])<<24 | uint32(data[i+1])<<16 | uint32(data[i+2])<<8 | uint32(data[i+3])
		mnem, operand := disassemble(word)
		lines = append(lines, assembler.Line{
			Text: fmt.Sprintf("%-6d  %08X  %-12s  %s", addr, word, mnem, operand),
			Kind: assembler.LineNormal,
		})
	}
	return lines
}

func (m model) renderStatusBar() string {
	// Left: key hints based on mode
	var hints string
	switch m.ed.Mode {
	case editor.ModeInsert:
		hints = " ESC Normal │ Ctrl+S Save │ Ctrl+B Build"
	case editor.ModeCommand:
		hints = " :w Save │ :q Quit │ :build Assemble │ :run Execute"
	case editor.ModeVisual:
		hints = " ESC Cancel │ d Delete │ y Yank"
	default:
		hints = " i Insert │ : Command │ Ctrl+B Build │ Ctrl+R Run │ Ctrl+W Focus │ Ctrl+] Tab"
	}

	left := lipgloss.NewStyle().Foreground(theme.TextMuted).Render(hints)

	// Right: status message
	right := lipgloss.NewStyle().Foreground(theme.TextSecond).Render(m.statusMsg + " ")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().
		Background(theme.BgMuted).
		Width(m.width).
		Render(bar)
}

func (m model) overlayNotification(page string) string {
	lines := strings.Split(page, "\n")

	msg := " " + m.notification + " "
	boxW := len(msg) + 4
	boxX := (m.width - boxW) / 2
	boxY := m.height / 2

	topLine := "╭" + strings.Repeat("─", boxW-2) + "╮"
	midLine := "│" + lipgloss.NewStyle().
		Bold(true).Foreground(theme.SpryzexGlow).
		Width(boxW-2).Align(lipgloss.Center).Render(m.notification) + "│"
	botLine := "╰" + strings.Repeat("─", boxW-2) + "╯"

	overlay := func(line, insert string, x int) string {
		runes := []rune(line)
		ins := []rune(insert)
		if x < 0 {
			x = 0
		}
		for x < len(runes) && x < len(runes) && len(ins) > 0 {
			if x+len(ins) <= len(runes) {
				copy(runes[x:], ins)
			}
			break
		}
		return string(runes)
	}
	_ = overlay

	// Simple: inject box into page string at correct Y position
	boxStyle := lipgloss.NewStyle().
		Background(theme.BgOverlay).
		Foreground(theme.SpryzexGlow).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.SpryzexRed).
		Padding(0, 2).
		Bold(true)

	notification := boxStyle.Render(m.notification)
	notifW := lipgloss.Width(notification)
	padLeft := (m.width - notifW) / 2

	row := boxY
	if row >= len(lines) {
		row = len(lines) / 2
	}

	_ = topLine
	_ = midLine
	_ = botLine
	_ = boxX

	lines[row] = strings.Repeat(" ", padLeft) + notification
	return strings.Join(lines, "\n")
}

// ---- Disassembler ----

var opNames = []string{
	"ldc", "adc", "ldl", "stl", "ldnl", "stnl", "add", "sub",
	"shl", "shr", "adj", "a2sp", "sp2a", "call", "return", "brz",
	"brlz", "br", "HALT", "data",
}

func disassemble(word uint32) (string, string) {
	opcode := word >> 24
	operand := int32(word & 0x00FFFFFF)
	// Sign extend 24-bit
	if operand&0x800000 != 0 {
		operand |= ^int32(0xFFFFFF)
	}

	if int(opcode) < len(opNames) {
		return opNames[opcode], fmt.Sprintf("%d", operand)
	}
	return fmt.Sprintf("0x%02X", opcode), fmt.Sprintf("0x%06X", uint32(operand))
}

// ---- Helpers ----

func findProjectRoot(filePath string) string {
	if filePath == "" {
		dir, _ := os.Getwd()
		return dir
	}
	dir := filepath.Dir(filePath)
	// Walk up looking for Makefile
	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(filePath)
}

func truncate(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	if maxW > 3 {
		return string(runes[:maxW-3]) + "..."
	}
	return string(runes[:maxW])
}

func defaultASM() []string {
	return []string{
		"; Welcome to SPRYZEX IDE",
		"; Assembly Language Editor",
		"; ─────────────────────────────────────",
		";",
		"; MODAL EDITOR — vim-style keybindings:",
		";   i         Enter INSERT mode",
		";   ESC       Return to NORMAL mode",
		";   :w        Save file",
		";   :build    Assemble",
		";   :run      Run emulator",
		";   Ctrl+B    Build",
		";   Ctrl+R    Run",
		";",
		"",
		"; Hello World example",
		"start:",
		"    ldc  72   ; H",
		"    outc",
		"    ldc  101  ; e",
		"    outc",
		"    ldc  108  ; l",
		"    outc",
		"    ldc  108  ; l",
		"    outc",
		"    ldc  111  ; o",
		"    outc",
		"    ldc  10   ; newline",
		"    outc",
		"    HALT",
	}
}

// ---- Build the C assembler/emulator from source ----
func buildCProject(root string) error {
	cmd := exec.Command("make", "-C", root, "all")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	var filePath string
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	// Optional: try to build the C project
	root := findProjectRoot(filePath)
	if _, err := os.Stat(filepath.Join(root, "Makefile")); err == nil {
		// Run make silently; don't block startup
		go func() {
			_ = buildCProject(root)
		}()
	}

	m := initialModel(filePath)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
