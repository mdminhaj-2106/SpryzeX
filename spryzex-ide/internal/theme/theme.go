package theme

import "github.com/charmbracelet/lipgloss"

var (
	BgDeep    = lipgloss.Color("#090B10")
	BgSurface = lipgloss.Color("#0F1219")
	BgOverlay = lipgloss.Color("#161B26")
	BgMuted   = lipgloss.Color("#1C2232")
	BgSelect  = lipgloss.Color("#1E2840")

	SpryzexRed    = lipgloss.Color("#C1440E")
	SpryzexBright = lipgloss.Color("#E85D26")
	SpryzexGlow   = lipgloss.Color("#FF6B35")
	SpryzexDust   = lipgloss.Color("#7A3A0A")

	PhobosBlue  = lipgloss.Color("#4A9EFF")
	DeimosGold  = lipgloss.Color("#FFD700")
	NebulaPurp  = lipgloss.Color("#B48EAD")
	AuroraGreen = lipgloss.Color("#A3BE8C")
	CometCyan   = lipgloss.Color("#88C0D0")

	TextPrimary = lipgloss.Color("#DCE0EC")
	TextSecond  = lipgloss.Color("#8A93AE")
	TextMuted   = lipgloss.Color("#485068")
	TextComment = lipgloss.Color("#3A4256")

	SynKeyword  = lipgloss.Color("#FF6B35")
	SynLabel    = lipgloss.Color("#4A9EFF")
	SynNumber   = lipgloss.Color("#FFD700")
	SynComment  = lipgloss.Color("#3E4A62")
	SynDirectiv = lipgloss.Color("#B48EAD")
	SynReg      = lipgloss.Color("#88C0D0")
	SynString   = lipgloss.Color("#A3BE8C")

	ColorOK       = lipgloss.Color("#A3BE8C")
	ColorError    = lipgloss.Color("#BF616A")
	ColorWarn     = lipgloss.Color("#EBCB8B")
	ColorInfo     = lipgloss.Color("#4A9EFF")
	ColorBuilding = lipgloss.Color("#FF6B35")

	BorderSubtle  = lipgloss.Color("#1A2035")
	BorderActive  = lipgloss.Color("#2A3550")
	BorderFocused = lipgloss.Color("#4A9EFF")
)

var (
	PanelStyle = lipgloss.NewStyle().
			Background(BgSurface).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderSubtle)

	ActivePanelStyle = lipgloss.NewStyle().
				Background(BgSurface).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(SpryzexRed)

	TitleBarStyle = lipgloss.NewStyle().
			Background(BgDeep).
			Foreground(SpryzexGlow).
			Bold(true)

	ModeNormalStyle = lipgloss.NewStyle().
			Background(SpryzexRed).
			Foreground(lipgloss.Color("#FFFFFF")).
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

	StatusBarStyle = lipgloss.NewStyle().
			Background(BgDeep).
			Foreground(TextSecond)

	StatusOKStyle = lipgloss.NewStyle().
			Background(ColorOK).
			Foreground(BgDeep).
			Bold(true).
			Padding(0, 1)

	StatusErrStyle = lipgloss.NewStyle().
			Background(ColorError).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	StatusBuildStyle = lipgloss.NewStyle().
				Background(SpryzexBright).
				Foreground(BgDeep).
				Bold(true).
				Padding(0, 1)

	LineNumStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgSurface)

	LineNumActiveStyle = lipgloss.NewStyle().
				Foreground(SpryzexBright).
				Background(BgSurface).
				Bold(true)

	CursorLineStyle = lipgloss.NewStyle().
			Background(BgSelect)

	TabStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgMuted).
			Padding(0, 2)

	TabActiveStyle = lipgloss.NewStyle().
			Foreground(SpryzexGlow).
			Background(BgOverlay).
			Bold(true).
			Padding(0, 2).
			Underline(true)
)
