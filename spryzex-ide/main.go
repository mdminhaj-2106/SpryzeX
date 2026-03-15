package main

import (
        "fmt"
        "io"
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
type buildDoneMsg struct {
        result *assembler.Result
        asmBin string
        emuBin string
}
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
        building       bool
        running        bool
        lastResult     *assembler.Result  // latest build or run (for LIVE tab + status)
        lastBuildResult *assembler.Result // last successful build (for LOG/LISTING/OBJ paths)
        activeTab      OutputTab
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

type layoutMetrics struct {
        spryzexW  int
        editorW   int
        topH      int
        consoleH  int
        titleH    int
        statusH   int
        editorX   int
        editorY   int
        consoleY  int
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
                m.lastBuildResult = msg.result
                if msg.asmBin != "" {
                        m.asmBin = msg.asmBin
                }
                if msg.emuBin != "" {
                        m.emuBin = msg.emuBin
                }
                m.buildCount++
                m.ed.SetDiagnostics(assembler.ExtractDiagnostics(msg.result))
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

        case tea.MouseMsg:
                return m.handleMouse(msg)
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

func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
        layout := m.layout()
        if layout.topH <= 0 || layout.consoleH <= 0 {
                return m, nil
        }

        hoverEditor := msg.Y >= layout.editorY && msg.Y < layout.editorY+layout.topH &&
                msg.X >= layout.editorX && msg.X < layout.editorX+layout.editorW
        hoverConsole := msg.Y >= layout.consoleY && msg.Y < layout.consoleY+layout.consoleH
        hoverSpryzex := layout.spryzexW > 0 &&
                msg.Y >= layout.editorY && msg.Y < layout.editorY+layout.topH &&
                msg.X >= 0 && msg.X < layout.spryzexW

        if msg.Action == tea.MouseActionPress {
                switch msg.Button {
                case tea.MouseButtonWheelUp:
                        switch {
                        case hoverEditor:
                                m.focus = FocusEditor
                                m.ed.ScrollBy(-3)
                        case hoverConsole:
                                m.focus = FocusConsole
                                m.scrollConsole(-3)
                        }
                        return m, nil
                case tea.MouseButtonWheelDown:
                        switch {
                        case hoverEditor:
                                m.focus = FocusEditor
                                m.ed.ScrollBy(3)
                        case hoverConsole:
                                m.focus = FocusConsole
                                m.scrollConsole(3)
                        }
                        return m, nil
                case tea.MouseButtonLeft:
                        switch {
                        case hoverEditor:
                                m.focus = FocusEditor
                                localX := msg.X - layout.editorX - 1
                                localY := msg.Y - layout.editorY - 2
                                m.ed.MoveToViewPosition(localX, localY)
                        case hoverConsole:
                                m.focus = FocusConsole
                        case hoverSpryzex:
                                m.focus = FocusSpryzex
                        }
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

        projectRoot := m.projectRoot
        asmBin := m.asmBin
        filePath := m.ed.FilePath

        return func() tea.Msg {
                if asmBin == "" {
                        _ = buildCProject(projectRoot)
                        asmBin = assembler.FindAssembler(projectRoot)
                }
                emuBin := assembler.FindEmulator(projectRoot)
                result := assembler.Assemble(asmBin, filePath, nil)
                return buildDoneMsg{result: result, asmBin: asmBin, emuBin: emuBin}
        }
}

func (m *model) triggerRun() tea.Cmd {
        if m.running {
                return nil
        }
        // First save
        _ = m.ed.SaveFile()

        // Build first if no successful build yet
        if m.lastBuildResult == nil || !m.lastBuildResult.Success {
                return m.triggerBuild()
        }

        m.running = true
        m.anim.SetState(spryzex.StateRunning)
        m.activeTab = TabLive
        m.consoleScrl = 0
        m.statusMsg = "Running..."

        projectRoot := m.projectRoot
        emuBin := m.emuBin
        objPath := m.lastBuildResult.ObjPath

        return func() tea.Msg {
                if emuBin == "" {
                        _ = buildCProject(projectRoot)
                        emuBin = assembler.FindEmulator(projectRoot)
                }
                result := assembler.Run(emuBin, objPath, []string{"-trace", "-before", "-after"})
                return runDoneMsg{result}
        }
}

func (m *model) notify(msg string) {
        m.notification = msg
        m.notifyUntil = time.Now().Add(3 * time.Second)
}

func (m *model) updateEditorSize() {
        layout := m.layout()
        editorInnerW := layout.editorW - 2
        editorInnerH := layout.topH - 4 // border + header + command bar reserve
        if editorInnerW < 20 {
                editorInnerW = 20
        }
        if editorInnerH < 3 {
                editorInnerH = 3
        }
        m.ed.SetSize(editorInnerW, editorInnerH)
}

// ---- View ----

func (m model) View() string {
        if m.width == 0 {
                return "Initializing..."
        }

        layout := m.layout()

        // --- Title bar ---
        titleBar := m.renderTitleBar()

        // --- Top section: Spryzex | Editor ---
        var topSection string
        if layout.spryzexW > 0 {
                spryzexPanel := m.renderSpryzexPanel(layout.spryzexW, layout.topH)
                edPanel := m.renderEditorPanel(layout.editorW, layout.topH)
                topSection = lipgloss.JoinHorizontal(lipgloss.Top, spryzexPanel, edPanel)
        } else {
                topSection = m.renderEditorPanel(m.width, layout.topH)
        }

        // --- Console ---
        console := m.renderConsole(m.width, layout.consoleH)

        // --- Status bar ---
        statusBar := m.renderStatusBar()

        return lipgloss.JoinVertical(lipgloss.Left,
                titleBar,
                topSection,
                console,
                statusBar,
        )
}

func (m model) renderTitleBar() string {
        divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A3050")).Render(" │ ")

        // Left: logo
        logoIcon := lipgloss.NewStyle().Foreground(theme.SpryzexGlow).Bold(true).Render("◈")
        logoText := lipgloss.NewStyle().Foreground(lipgloss.Color("#C8CCDC")).Bold(true).Render("SPRYZEX")
        logoSub := lipgloss.NewStyle().Foreground(theme.TextMuted).Render(" IDE")
        logo := " " + logoIcon + " " + logoText + logoSub

        // File tab
        fname := m.ed.FilePath
        if fname == "" {
                fname = "untitled.asm"
        }
        base := filepath.Base(fname)
        dirtyMark := ""
        if m.ed.Dirty {
                dirtyMark = lipgloss.NewStyle().Foreground(theme.SpryzexGlow).Render(" ●")
        }
        fileTab := lipgloss.NewStyle().
                Foreground(theme.TextPrimary).
                Bold(true).
                Render(base) + dirtyMark

        left := logo + divider + fileTab

        // Right: build status pill + time
        var statusPill string
        switch {
        case m.building:
                spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴"}
                sp := spinner[(m.buildCount+int(time.Now().UnixMilli()/100))%len(spinner)]
                statusPill = theme.StatusBuildStyle.Render(" " + sp + " BUILDING ")
        case m.running:
                statusPill = lipgloss.NewStyle().
                        Background(theme.PhobosBlue).Foreground(theme.BgDeep).
                        Bold(true).Padding(0, 1).Render(" ▶ RUNNING ")
        case m.lastResult != nil && m.lastResult.Success:
                statusPill = theme.StatusOKStyle.Render(" ✓ OK ")
        case m.lastResult != nil && !m.lastResult.Success:
                statusPill = theme.StatusErrStyle.Render(fmt.Sprintf(" ✗ %d ERR ", m.lastResult.ErrorCount))
        default:
                statusPill = lipgloss.NewStyle().
                        Foreground(theme.TextMuted).
                        Background(theme.BgMuted).
                        Padding(0, 1).
                        Render("READY")
        }

        timeStr := lipgloss.NewStyle().
                Foreground(lipgloss.Color("#3A4560")).
                Render(time.Now().Format(" 15:04:05 "))

        right := statusPill + timeStr

        leftW := lipgloss.Width(left)
        rightW := lipgloss.Width(right)
        gap := m.width - leftW - rightW
        if gap < 0 {
                gap = 0
        }
        filler := lipgloss.NewStyle().Background(theme.BgDeep).Render(strings.Repeat(" ", gap))

        bar := lipgloss.NewStyle().Background(theme.BgDeep).Render(left) +
                filler +
                lipgloss.NewStyle().Background(theme.BgDeep).Render(right)

        return bar + "\n"
}

func (m model) renderSpryzexPanel(w, h int) string {
        spryzexH := h - 2
        spryzexW := w - 2
        if spryzexH < 5 {
                spryzexH = 5
        }
        if spryzexW < 10 {
                spryzexW = 10
        }

        // Tell the face where the cursor is so eyes can track it
        layout := m.layout()
        edH := layout.topH - 4
        edW := layout.editorW - 2
        if edH < 1 {
                edH = 1
        }
        if edW < 1 {
                edW = 1
        }
        m.anim.SetCursor(m.ed.CursorRow, m.ed.CursorCol, edH, edW)

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
                Background(theme.BgDeep).
                Render(content)
}

func (m model) renderEditorPanel(w, h int) string {
        innerW := w - 2
        innerH := h - 2

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
        diagMsg := m.ed.DiagnosticAt(m.ed.CursorRow)

        // Command bar
        cmdBar := ""
        if m.ed.Mode == editor.ModeCommand {
                cmdBar = lipgloss.NewStyle().
                        Width(innerW).
                        Background(theme.BgOverlay).
                        Foreground(theme.TextPrimary).
                        Render(truncate(":"+m.ed.CommandBuf, innerW))
        } else if diagMsg != "" {
                cmdBar = lipgloss.NewStyle().
                        Width(innerW).
                        Foreground(theme.ColorError).
                        Render(truncate(diagMsg, innerW))
        } else if m.ed.CmdMsg != "" {
                cmdBar = lipgloss.NewStyle().
                        Width(innerW).
                        Foreground(theme.TextMuted).
                        Render(truncate(m.ed.CmdMsg, innerW))
        }

        editorBodyH := innerH - 1 // header
        if cmdBar != "" {
                editorBodyH--
        }
        if editorBodyH < 1 {
                editorBodyH = 1
        }
        m.ed.SetSize(innerW, editorBodyH)
        editorContent := m.ed.View(m.focus == FocusEditor)

        // Header: left = file icon + name, right = position + mode
        fileIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A6080")).Render("󰘡 ")
        headerFileName := lipgloss.NewStyle().Foreground(theme.TextSecond).Render(filepath.Base(m.ed.FilePath))
        headerLeft := fileIcon + headerFileName

        posInfo := lipgloss.NewStyle().
                Foreground(theme.TextMuted).
                Render(fmt.Sprintf("  %d:%d  ", row, col))

        rightSection := posInfo + modePill

        headerLeftW := lipgloss.Width(headerLeft)
        rightW := lipgloss.Width(rightSection)
        headerGap := innerW - headerLeftW - rightW
        if headerGap < 0 {
                headerGap = 0
        }
        headerFill := lipgloss.NewStyle().Background(theme.BgOverlay).Render(strings.Repeat(" ", headerGap))
        header := lipgloss.NewStyle().Background(theme.BgOverlay).Render(headerLeft) +
                headerFill +
                rightSection

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
                r := m.lastBuildResult
                if r == nil {
                        r = m.lastResult
                }
                return m.loadFileLines(r.LogPath)
        case TabListing:
                r := m.lastBuildResult
                if r == nil {
                        r = m.lastResult
                }
                return m.loadFileLines(r.ListingPath)
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
        r := m.lastBuildResult
        if r == nil {
                r = m.lastResult
        }
        if r == nil || r.ObjPath == "" {
                return []assembler.Line{{Text: "No .o file available yet", Kind: assembler.LineInfo}}
        }
        data, err := os.ReadFile(r.ObjPath)
        if err != nil {
                return []assembler.Line{{Text: fmt.Sprintf("Cannot read %s: %v", r.ObjPath, err), Kind: assembler.LineError}}
        }

        var lines []assembler.Line
        lines = append(lines, assembler.Line{
                Text: fmt.Sprintf("Object file: %s (%d bytes)", r.ObjPath, len(data)),
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
        sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A3050")).Render(" │ ")

        keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A6090")).Bold(true)
        valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#323C58"))

        key := func(k, v string) string {
                return keyStyle.Render(k) + valStyle.Render(v)
        }

        var hints string
        switch m.ed.Mode {
        case editor.ModeInsert:
                hints = key("ESC", " Normal") + sep + key("^S", " Save") + sep + key("^B", " Build")
        case editor.ModeCommand:
                hints = key(":w", " Save") + sep + key(":q", " Quit") + sep + key(":build", " Assemble") + sep + key(":run", " Execute")
        case editor.ModeVisual:
                hints = key("ESC", " Cancel") + sep + key("d", " Delete") + sep + key("y", " Yank")
        default:
                hints = key("i", " Insert") + sep + key(":", " Cmd") + sep + key("^B", " Build") + sep + key("^R", " Run") + sep + key("^W", " Focus") + sep + key("^]", " Tab")
        }
        left := " " + hints

        // Right: status message
        statusColor := theme.TextMuted
        if m.statusMsg != "" {
                if strings.Contains(m.statusMsg, "FAILED") || strings.Contains(m.statusMsg, "error") {
                        statusColor = theme.ColorError
                } else if strings.Contains(m.statusMsg, "OK") || strings.Contains(m.statusMsg, "ok") {
                        statusColor = theme.ColorOK
                } else {
                        statusColor = theme.CometCyan
                }
        }
        right := lipgloss.NewStyle().Foreground(statusColor).Render(m.statusMsg + " ")

        leftW := lipgloss.Width(left)
        rightW := lipgloss.Width(right)
        gap := m.width - leftW - rightW
        if gap < 0 {
                gap = 0
        }

        bar := left + strings.Repeat(" ", gap) + right
        return lipgloss.NewStyle().
                Background(theme.BgDeep).
                Width(m.width).
                Render(bar)
}

func (m model) overlayNotification(page string) string {
        lines := strings.Split(page, "\n")
        boxY := m.height / 2

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
        if padLeft < 0 {
                padLeft = 0
        }
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
        cmd.Stdout = io.Discard
        cmd.Stderr = io.Discard
        return cmd.Run()
}

func main() {
        var filePath string
        if len(os.Args) > 1 {
                filePath = os.Args[1]
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

func (m model) layout() layoutMetrics {
        availableH := m.height - 2 // title + status
        if availableH < 2 {
                availableH = 2
        }

        spryzexW := 30
        if m.width < 100 {
                spryzexW = 0
        }

        consoleH := availableH / 3
        topH := availableH - consoleH

        if availableH >= 10 && topH < 6 {
                topH = 6
                consoleH = availableH - topH
        }
        if availableH >= 8 && consoleH < 3 {
                consoleH = 3
                topH = availableH - consoleH
        }
        if topH < 1 {
                topH = 1
        }
        if consoleH < 1 {
                consoleH = 1
        }
        editorW := m.width - spryzexW
        if editorW < 20 {
                editorW = 20
        }

        return layoutMetrics{
                spryzexW: spryzexW,
                editorW:  editorW,
                topH:     topH,
                consoleH: consoleH,
                titleH:   1,
                statusH:  1,
                editorX:  spryzexW,
                editorY:  1,
                consoleY: 1 + topH,
        }
}

func (m *model) scrollConsole(delta int) {
        lines := m.getConsoleLines()
        maxScroll := len(lines) - max(m.layout().consoleH-4, 1)
        if maxScroll < 0 {
                maxScroll = 0
        }
        m.consoleScrl += delta
        if m.consoleScrl < 0 {
                m.consoleScrl = 0
        }
        if m.consoleScrl > maxScroll {
                m.consoleScrl = maxScroll
        }
}
