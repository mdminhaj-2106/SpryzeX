package theme

import "github.com/charmbracelet/lipgloss"

// SPRYZEX color palette — deep space meets red planet
var (
	// Base palette
	BgDeep    = lipgloss.Color("#0D0F14") // near-black space
	BgSurface = lipgloss.Color("#13161E") // panel background
	BgOverlay = lipgloss.Color("#1A1E2A") // elevated panels
	BgMuted   = lipgloss.Color("#1F2433") // subtle backgrounds
	BgSelect  = lipgloss.Color("#263045") // selected line

	// SPRYZEX reds/oranges
	SpryzexRed    = lipgloss.Color("#C1440E") // spryzex surface red
	SpryzexBright = lipgloss.Color("#E85D26") // bright rust orange
	SpryzexGlow   = lipgloss.Color("#FF6B35") // hot glow
	SpryzexDust   = lipgloss.Color("#8B4513") // dark dust

	// Accent colors
	PhobosBlue  = lipgloss.Color("#4A9EFF") // Phobos — icy blue
	DeimosGold  = lipgloss.Color("#FFD700") // Deimos — golden
	NebulaPurp  = lipgloss.Color("#B48EAD") // nebula purple
	AuroraGreen = lipgloss.Color("#A3BE8C") // aurora green
	CometCyan   = lipgloss.Color("#88C0D0") // comet cyan

	// Text
	TextPrimary = lipgloss.Color("#E8EAF0") // main text
	TextSecond  = lipgloss.Color("#9BA3B8") // secondary text
	TextMuted   = lipgloss.Color("#5C6478") // muted text
	TextComment = lipgloss.Color("#4A5268") // comments

	// Syntax highlighting colors
	SynKeyword  = lipgloss.Color("#E85D26") // mnemonics — spryzex orange
	SynLabel    = lipgloss.Color("#4A9EFF") // labels — phobos blue
	SynNumber   = lipgloss.Color("#FFD700") // numbers — deimos gold
	SynComment  = lipgloss.Color("#4A5268") // comments — dark
	SynDirectiv = lipgloss.Color("#B48EAD") // directives — purple
	SynReg      = lipgloss.Color("#88C0D0") // registers — cyan
	SynString   = lipgloss.Color("#A3BE8C") // strings — green

	// Status colors
	ColorOK       = lipgloss.Color("#A3BE8C") // green success
	ColorError    = lipgloss.Color("#BF616A") // red error
	ColorWarn     = lipgloss.Color("#EBCB8B") // yellow warning
	ColorInfo     = lipgloss.Color("#4A9EFF") // blue info
	ColorBuilding = lipgloss.Color("#FF6B35") // orange building

	// Border styles
	BorderSubtle  = lipgloss.Color("#1F2433")
	BorderActive  = lipgloss.Color("#E85D26")
	BorderFocused = lipgloss.Color("#4A9EFF")
)

// Styles — composed lipgloss styles used throughout
var (
	// Panel borders
	PanelStyle = lipgloss.NewStyle().
			Background(BgSurface).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderSubtle)

	ActivePanelStyle = lipgloss.NewStyle().
				Background(BgSurface).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(BorderActive)

	// Title bar
	TitleBarStyle = lipgloss.NewStyle().
			Background(BgDeep).
			Foreground(SpryzexGlow).
			Bold(true)

	// Mode indicator pill
	ModeNormalStyle = lipgloss.NewStyle().
			Background(SpryzexRed).
			Foreground(TextPrimary).
			Bold(true).
			Padding(0, 1)

	ModeInsertStyle = lipgloss.NewStyle().
			Background(PhobosBlue).
			Foreground(BgDeep).
			Bold(true).
			Padding(0, 1)

	ModeCommandStyle = lipgloss.NewStyle().
				Background(NebulaPurp).
				Foreground(BgDeep).
				Bold(true).
				Padding(0, 1)

	ModeVisualStyle = lipgloss.NewStyle().
			Background(DeimosGold).
			Foreground(BgDeep).
			Bold(true).
			Padding(0, 1)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Background(BgMuted).
			Foreground(TextSecond)

	StatusOKStyle = lipgloss.NewStyle().
			Background(ColorOK).
			Foreground(BgDeep).
			Bold(true).
			Padding(0, 1)

	StatusErrStyle = lipgloss.NewStyle().
			Background(ColorError).
			Foreground(TextPrimary).
			Bold(true).
			Padding(0, 1)

	StatusBuildStyle = lipgloss.NewStyle().
				Background(SpryzexBright).
				Foreground(BgDeep).
				Bold(true).
				Padding(0, 1)

	// Line numbers
	LineNumStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgSurface)

	LineNumActiveStyle = lipgloss.NewStyle().
				Foreground(SpryzexBright).
				Background(BgSurface).
				Bold(true)

	// Cursor line
	CursorLineStyle = lipgloss.NewStyle().
			Background(BgSelect)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Padding(0, 2)

	TabActiveStyle = lipgloss.NewStyle().
			Foreground(SpryzexGlow).
			Background(BgOverlay).
			Bold(true).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(false).
			BorderForeground(SpryzexRed)
)
