package spryzex

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"spryzex-ide/internal/theme"
)

// ---- State ----

type BuildState int

const (
	StateIdle     BuildState = iota
	StateBuilding
	StateSuccess
	StateError
	StateRunning
)

// ---- Animator ----

type Animator struct {
	State      BuildState
	Frame      int
	blinkTimer int
	burstFrame int

	// Cursor tracking: fractions 0.0–1.0 within the editor
	gazeX float64
	gazeY float64

	width  int
	height int
}

func NewAnimator(w, h int) *Animator {
	return &Animator{State: StateIdle, width: w, height: h}
}

// SetCursor is called each render with the editor cursor position.
func (a *Animator) SetCursor(cursorRow, cursorCol, editorH, editorW int) {
	if editorW > 1 {
		a.gazeX = clamp01(float64(cursorCol) / float64(editorW))
	}
	if editorH > 1 {
		a.gazeY = clamp01(float64(cursorRow) / float64(editorH))
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (a *Animator) Tick() {
	a.Frame++
	a.blinkTimer++
	if a.State == StateSuccess {
		a.burstFrame++
	} else {
		a.burstFrame = 0
	}
}

func (a *Animator) SetState(s BuildState) {
	a.State = s
	a.burstFrame = 0
}

// ---- Face parts ----

func (a *Animator) pupilRune() rune {
	if a.blinkTimer%80 < 3 { // blink ~every 4s
		return '─'
	}
	switch a.State {
	case StateError:
		if a.Frame%16 < 8 {
			return '×'
		}
		return '▪'
	case StateBuilding:
		seq := []rune{'·', '○', '●', '○'}
		return seq[(a.Frame/4)%len(seq)]
	case StateSuccess:
		return '^'
	case StateRunning:
		return '◉'
	default:
		return '●'
	}
}

func (a *Animator) eyeColor() lipgloss.Color {
	switch a.State {
	case StateError:
		return theme.ColorError
	case StateSuccess:
		return theme.AuroraGreen
	case StateBuilding:
		return theme.SpryzexGlow
	case StateRunning:
		return theme.PhobosBlue
	default:
		return theme.CometCyan
	}
}

// eyebrowChars returns (leftChar, rightChar) for the eyebrows.
func (a *Animator) eyebrowChars() (string, string) {
	s := func(c string, col lipgloss.Color) string {
		return lipgloss.NewStyle().Foreground(col).Bold(true).Render(c)
	}
	switch a.State {
	case StateError:
		return s("╲", theme.ColorError), s("╱", theme.ColorError)
	case StateSuccess:
		return s("╱", theme.AuroraGreen), s("╲", theme.AuroraGreen)
	case StateBuilding:
		if a.Frame%10 < 5 {
			return s("╱", theme.SpryzexGlow), s("╲", theme.SpryzexGlow)
		}
		return s("─", theme.SpryzexGlow), s("─", theme.SpryzexGlow)
	default:
		col := lipgloss.Color("#3A4A68")
		return s("─", col), s("─", col)
	}
}

// ledColor returns the color for the blinking antenna LED.
func (a *Animator) ledColor() lipgloss.Color {
	switch a.State {
	case StateBuilding:
		seq := []lipgloss.Color{theme.SpryzexGlow, theme.DeimosGold, theme.SpryzexBright}
		return seq[(a.Frame/3)%len(seq)]
	case StateSuccess:
		seq := []lipgloss.Color{theme.AuroraGreen, theme.DeimosGold}
		return seq[(a.Frame/5)%len(seq)]
	case StateError:
		if a.Frame%8 < 4 {
			return theme.ColorError
		}
		return lipgloss.Color("#400810")
	case StateRunning:
		return theme.PhobosBlue
	default:
		if a.Frame%70 < 3 {
			return theme.SpryzexGlow // rare twinkle
		}
		return lipgloss.Color("#181E30")
	}
}

// drawMouth writes the mouth characters one-per-cell into the grid.
func (a *Animator) drawMouth(grid [][]string, mouthRow, cx, mouthW int) {
	inner := mouthW - 2
	if inner < 2 {
		inner = 2
	}

	type mouthDef struct {
		left, right rune
		fill        rune
		col         lipgloss.Color
	}

	var def mouthDef
	switch a.State {
	case StateSuccess:
		def = mouthDef{'╰', '╯', '─', theme.AuroraGreen}
	case StateError:
		if a.Frame%14 < 7 {
			def = mouthDef{'╭', '╮', '─', theme.ColorError}
		} else {
			def = mouthDef{'╭', '╮', '~', theme.ColorError}
		}
	case StateBuilding:
		// Animated dots mouth
		st := lipgloss.NewStyle().Foreground(theme.SpryzexGlow)
		dotSeq := []string{" ·    ", "  ·   ", "   ·  ", "  ·   "}
		d := dotSeq[(a.Frame/4)%len(dotSeq)]
		runes := []rune("│" + fmt.Sprintf("%-*s", inner, d) + "│")
		startX := cx - len(runes)/2
		for i, r := range runes {
			setCell(grid, mouthRow, startX+i, st.Render(string(r)))
		}
		return
	case StateRunning:
		st := lipgloss.NewStyle().Foreground(theme.PhobosBlue)
		frames := []string{"▷──", "─▷─", "──▷"}
		d := frames[(a.Frame/4)%len(frames)]
		mid := d
		if len([]rune(mid)) < inner {
			mid += strings.Repeat("─", inner-len([]rune(mid)))
		}
		runes := []rune("│" + string([]rune(mid)[:inner]) + "│")
		startX := cx - len(runes)/2
		for i, r := range runes {
			setCell(grid, mouthRow, startX+i, st.Render(string(r)))
		}
		return
	default:
		def = mouthDef{'╰', '╯', '─', lipgloss.Color("#4A6090")}
	}

	st := lipgloss.NewStyle().Foreground(def.col).Bold(true)
	runes := make([]rune, inner+2)
	runes[0] = def.left
	runes[inner+1] = def.right
	for i := 1; i <= inner; i++ {
		runes[i] = def.fill
	}
	startX := cx - len(runes)/2
	for i, r := range runes {
		setCell(grid, mouthRow, startX+i, st.Render(string(r)))
	}
}

// ---- Grid helpers ----

func setCell(grid [][]string, y, x int, s string) {
	if y < 0 || y >= len(grid) {
		return
	}
	if x < 0 || x >= len(grid[y]) {
		return
	}
	grid[y][x] = s
}

func hlineStyle(grid [][]string, y, x1, x2 int, st lipgloss.Style, ch rune) {
	for x := x1; x <= x2; x++ {
		setCell(grid, y, x, st.Render(string(ch)))
	}
}

func vlineStyle(grid [][]string, x, y1, y2 int, st lipgloss.Style, ch rune) {
	for y := y1; y <= y2; y++ {
		setCell(grid, y, x, st.Render(string(ch)))
	}
}

// ---- Main draw ----

func (a *Animator) drawFace(grid [][]string, W, H int) {
	// Head dimensions
	headW := W - 2
	if headW > 22 {
		headW = 22
	}
	if headW < 10 {
		headW = 10
	}
	if headW%2 == 0 {
		headW--
	}

	headH := H - 2
	if headH < 7 {
		headH = 7
	}
	if headH > 13 {
		headH = 13
	}

	cx := W / 2
	cy := H / 2
	headL := cx - headW/2
	headR := headL + headW - 1
	headTop := cy - headH/2
	headBot := headTop + headH - 1

	// ── Styles ──
	rimSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5E82")).Bold(true)
	wallSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#2E3D5A"))
	boltSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#283448"))
	eyeFrameSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A4E72"))
	eyeDimSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#1C2538"))
	pupilSt := lipgloss.NewStyle().Foreground(a.eyeColor()).Bold(true)
	noseSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A3550"))

	// ── Antenna ──
	antRow := headTop - 2
	if antRow >= 0 {
		ledSt := lipgloss.NewStyle().Foreground(a.ledColor()).Bold(true)
		setCell(grid, antRow, cx, ledSt.Render("◆"))
		setCell(grid, antRow+1, cx, wallSt.Render("┃"))
	}

	// ── Head border ──
	setCell(grid, headTop, headL, rimSt.Render("╭"))
	setCell(grid, headTop, headR, rimSt.Render("╮"))
	setCell(grid, headBot, headL, rimSt.Render("╰"))
	setCell(grid, headBot, headR, rimSt.Render("╯"))
	hlineStyle(grid, headTop, headL+1, headR-1, wallSt, '─')
	hlineStyle(grid, headBot, headL+1, headR-1, wallSt, '─')
	vlineStyle(grid, headL, headTop+1, headBot-1, wallSt, '│')
	vlineStyle(grid, headR, headTop+1, headBot-1, wallSt, '│')
	// Antenna port on top edge
	setCell(grid, headTop, cx, rimSt.Render("┬"))

	// ── Corner bolts ──
	setCell(grid, headTop+1, headL+1, boltSt.Render("·"))
	setCell(grid, headTop+1, headR-1, boltSt.Render("·"))
	setCell(grid, headBot-1, headL+1, boltSt.Render("·"))
	setCell(grid, headBot-1, headR-1, boltSt.Render("·"))

	// ── Eyes ──
	eyeRow := headTop + 2

	// Map gazeX (0–1) to pupil offset -1,0,+1
	gx := int(math.Round((a.gazeX*2 - 1) * 1.3))
	if gx < -1 {
		gx = -1
	}
	if gx > 1 {
		gx = 1
	}

	// Eye center columns
	lEyeCX := cx - headW/4 - 1
	rEyeCX := cx + headW/4 + 1

	// Eyebrows
	lBrow, rBrow := a.eyebrowChars()
	setCell(grid, eyeRow-1, lEyeCX, lBrow)
	setCell(grid, eyeRow-1, rEyeCX, rBrow)

	blinking := a.blinkTimer%80 < 3

	for _, ecx := range []int{lEyeCX, rEyeCX} {
		// Eye frame: ┌───┐ / │···│ / └───┘
		setCell(grid, eyeRow, ecx-2, eyeFrameSt.Render("┌"))
		setCell(grid, eyeRow, ecx+2, eyeFrameSt.Render("┐"))
		hlineStyle(grid, eyeRow, ecx-1, ecx+1, eyeFrameSt, '─')
		setCell(grid, eyeRow+2, ecx-2, eyeFrameSt.Render("└"))
		setCell(grid, eyeRow+2, ecx+2, eyeFrameSt.Render("┘"))
		hlineStyle(grid, eyeRow+2, ecx-1, ecx+1, eyeFrameSt, '─')
		setCell(grid, eyeRow+1, ecx-2, eyeFrameSt.Render("│"))
		setCell(grid, eyeRow+1, ecx+2, eyeFrameSt.Render("│"))

		if blinking {
			hlineStyle(grid, eyeRow+1, ecx-1, ecx+1, eyeFrameSt, '─')
		} else {
			// Dim fill, then pupil at gaze offset
			hlineStyle(grid, eyeRow+1, ecx-1, ecx+1, eyeDimSt, '·')
			px := ecx + gx
			if px < ecx-1 {
				px = ecx - 1
			}
			if px > ecx+1 {
				px = ecx + 1
			}
			setCell(grid, eyeRow+1, px, pupilSt.Render(string(a.pupilRune())))
		}
	}

	// ── Nose ──
	noseRow := eyeRow + 3
	if noseRow < headBot-1 {
		setCell(grid, noseRow, cx-1, noseSt.Render("▪"))
		setCell(grid, noseRow, cx+1, noseSt.Render("▪"))
	}

	// ── Mouth ──
	mouthRow := headBot - 2
	mouthW := headW - 4
	if mouthW < 4 {
		mouthW = 4
	}
	a.drawMouth(grid, mouthRow, cx, mouthW)

	// ── Success burst ──
	if a.State == StateSuccess && a.burstFrame < 35 {
		starSt := lipgloss.NewStyle().Foreground(theme.DeimosGold).Bold(true)
		stars := []string{"★", "✦", "✧", "✸"}
		dist := float64(a.burstFrame) * 0.25
		for i := 0; i < 8; i++ {
			ang := float64(i) * math.Pi / 4.0
			bx := cx + int(math.Cos(ang)*dist*2.2)
			by := cy + int(math.Sin(ang)*dist*0.9)
			setCell(grid, by, bx, starSt.Render(stars[i%len(stars)]))
		}
	}

	// ── Error shake sparks ──
	if a.State == StateError && a.Frame%20 < 10 {
		sparkSt := lipgloss.NewStyle().Foreground(theme.ColorError).Bold(true)
		sparks := []string{"✗", "×", "!"}
		positions := [][2]int{
			{headTop + 1, headL - 1}, {headTop + 1, headR + 1},
		}
		for i, p := range positions {
			setCell(grid, p[0], p[1], sparkSt.Render(sparks[i%len(sparks)]))
		}
	}
}

// ---- Render ----

func (a *Animator) Render(W, H int) string {
	if W < 6 || H < 6 {
		return ""
	}

	artH := H
	hasLabel := H >= 8
	if hasLabel {
		artH = H - 1
	}

	grid := make([][]string, artH)
	for i := range grid {
		grid[i] = make([]string, W)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	a.drawFace(grid, W, artH)

	var sb strings.Builder
	for _, row := range grid {
		for _, cell := range row {
			sb.WriteString(cell)
		}
		sb.WriteRune('\n')
	}

	if hasLabel {
		sb.WriteString(lipgloss.NewStyle().
			Width(W).
			Align(lipgloss.Center).
			Foreground(a.stateColor()).
			Bold(true).
			Render(a.stateLabel()))
	}

	return sb.String()
}

// ---- State label ----

func (a *Animator) stateLabel() string {
	switch a.State {
	case StateBuilding:
		spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		s := spinners[(a.Frame/2)%len(spinners)]
		return s + " BUILDING " + s
	case StateSuccess:
		return "✦ BUILD OK ✦"
	case StateError:
		return "✗ BUILD FAILED"
	case StateRunning:
		frames := []string{"▶ RUNNING ·", "▶ RUNNING ··", "▶ RUNNING ···"}
		return frames[(a.Frame/4)%len(frames)]
	default:
		return "◈ READY"
	}
}

func (a *Animator) stateColor() lipgloss.Color {
	switch a.State {
	case StateBuilding:
		return theme.SpryzexGlow
	case StateSuccess:
		return theme.AuroraGreen
	case StateError:
		return theme.ColorError
	case StateRunning:
		return theme.PhobosBlue
	default:
		return theme.TextMuted
	}
}

func TickCmd() func() (string, error) {
	return func() (string, error) {
		time.Sleep(50 * time.Millisecond)
		return fmt.Sprintf("tick:%d", time.Now().UnixMilli()), nil
	}
}
