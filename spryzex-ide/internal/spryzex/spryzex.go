package spryzex

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"spryzex-ide/internal/theme"
)

type BuildState int

const (
	StateIdle     BuildState = iota
	StateBuilding
	StateSuccess
	StateError
	StateRunning
)

type Animator struct {
	State BuildState
	Frame int

	// Cursor position as fractions 0.0–1.0
	gazeX float64 // 0=left edge, 1=right edge of editor
	gazeY float64 // 0=top, 1=bottom

	blinkTimer  int
	burstFrame  int
	shakeOffset int

	width  int
	height int
}

func NewAnimator(w, h int) *Animator {
	return &Animator{
		State:  StateIdle,
		width:  w,
		height: h,
	}
}

// SetCursor is called each frame with the editor's cursor position.
func (a *Animator) SetCursor(cursorRow, cursorCol, editorH, editorW int) {
	if editorW > 0 {
		a.gazeX = math.Max(0, math.Min(1, float64(cursorCol)/float64(editorW)))
	}
	if editorH > 0 {
		a.gazeY = math.Max(0, math.Min(1, float64(cursorRow)/float64(editorH)))
	}
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

// ---- Internal helpers ----

// pupilChar returns the character to draw inside each eye.
func (a *Animator) pupilChar() rune {
	// Blink every ~4 seconds (80 frames at 50ms)
	if a.blinkTimer%80 < 3 {
		return '─'
	}
	switch a.State {
	case StateError:
		if a.Frame%20 < 10 {
			return '×'
		}
		return '▪'
	case StateBuilding:
		dots := []rune{'·', '·', '○', '○', '·', '·', '●', '●'}
		return dots[a.Frame%len(dots)]
	case StateSuccess:
		return '^'
	case StateRunning:
		return '◉'
	default:
		return '●'
	}
}

// eyebrowChar returns left and right eyebrow characters.
func (a *Animator) eyebrows() (left, right string) {
	neutral := lipgloss.NewStyle().Foreground(theme.TextMuted)
	angry := lipgloss.NewStyle().Foreground(theme.ColorError).Bold(true)
	happy := lipgloss.NewStyle().Foreground(theme.DeimosGold).Bold(true)
	build := lipgloss.NewStyle().Foreground(theme.SpryzexGlow)

	switch a.State {
	case StateError:
		return angry.Render("╲"), angry.Render("╱")
	case StateSuccess:
		return happy.Render("╱"), happy.Render("╲")
	case StateBuilding:
		if a.Frame%12 < 6 {
			return build.Render("─"), build.Render("─")
		}
		return build.Render("╱"), build.Render("╲")
	default:
		return neutral.Render("─"), neutral.Render("─")
	}
}

// mouthRows returns 1–2 rows of the mouth string, padded to mouthW.
func (a *Animator) mouthRows(mouthW int) []string {
	if mouthW < 4 {
		mouthW = 4
	}

	inner := mouthW - 2
	if inner < 2 {
		inner = 2
	}

	fill := strings.Repeat("─", inner)

	happy := lipgloss.NewStyle().Foreground(theme.AuroraGreen).Bold(true)
	neutral := lipgloss.NewStyle().Foreground(theme.TextSecond)
	angry := lipgloss.NewStyle().Foreground(theme.ColorError).Bold(true)
	build := lipgloss.NewStyle().Foreground(theme.SpryzexGlow)
	run := lipgloss.NewStyle().Foreground(theme.PhobosBlue)

	switch a.State {
	case StateSuccess:
		return []string{
			happy.Render("╰" + fill + "╯"),
		}
	case StateError:
		if a.Frame%16 < 8 {
			return []string{
				angry.Render("╭" + fill + "╮"),
			}
		}
		return []string{
			angry.Render("╭" + strings.Repeat("~", inner) + "╮"),
		}
	case StateBuilding:
		dots := []string{"·   ", " ·  ", "  · ", "   ·", "  · ", " ·  "}
		d := dots[(a.Frame/3)%len(dots)]
		padded := d
		if len([]rune(padded)) < inner {
			padded = d + strings.Repeat(" ", inner-len([]rune(d)))
		}
		return []string{
			build.Render("│" + padded[:inner] + "│"),
		}
	case StateRunning:
		frames := []string{"▶   ", " ▶  ", "  ▶ ", "   ▶"}
		d := frames[(a.Frame/4)%len(frames)]
		if len([]rune(d)) < inner {
			d = d + strings.Repeat(" ", inner-len([]rune(d)))
		}
		return []string{
			run.Render("│" + d[:inner] + "│"),
		}
	default:
		// Idle — gentle smile
		return []string{
			neutral.Render("╰" + fill + "╯"),
		}
	}
}

// ledColor returns the color for the antenna LED.
func (a *Animator) ledColor() lipgloss.Color {
	switch a.State {
	case StateBuilding:
		cols := []lipgloss.Color{theme.SpryzexGlow, theme.SpryzexBright, theme.DeimosGold}
		return cols[(a.Frame/4)%len(cols)]
	case StateSuccess:
		cols := []lipgloss.Color{theme.AuroraGreen, theme.DeimosGold}
		return cols[(a.Frame/6)%len(cols)]
	case StateError:
		if a.Frame%10 < 5 {
			return theme.ColorError
		}
		return lipgloss.Color("#5A1018")
	case StateRunning:
		return theme.PhobosBlue
	default:
		if a.Frame%60 < 3 {
			return theme.SpryzexGlow
		}
		return lipgloss.Color("#1A2035")
	}
}

// eyeColor returns the pupil/iris color.
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

// ---- Main render ----

func (a *Animator) Render(W, H int) string {
	if W < 4 || H < 4 {
		return ""
	}

	// Reserve bottom for status label
	artH := H
	hasLabel := H >= 8
	if hasLabel {
		artH = H - 1
	}

	// Build string grid
	grid := make([][]string, artH)
	for y := range grid {
		grid[y] = make([]string, W)
		for x := range grid[y] {
			grid[y][x] = " "
		}
	}

	a.drawFace(grid, W, artH)

	var sb strings.Builder
	for y, row := range grid {
		for _, cell := range row {
			sb.WriteString(cell)
		}
		sb.WriteRune('\n')
	}

	if hasLabel {
		label := a.stateLabel()
		sb.WriteString(lipgloss.NewStyle().
			Width(W).
			Align(lipgloss.Center).
			Foreground(a.stateColor()).
			Bold(true).
			Render(label))
	}

	return sb.String()
}

// setCell writes a single styled string to grid[y][x], bounds-checked.
func setCell(grid [][]string, y, x int, s string) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) {
		return
	}
	grid[y][x] = s
}

func hline(grid [][]string, y, x1, x2 int, s lipgloss.Style, ch string) {
	for x := x1; x <= x2; x++ {
		setCell(grid, y, x, s.Render(ch))
	}
}

func vline(grid [][]string, x, y1, y2 int, s lipgloss.Style, ch string) {
	for y := y1; y <= y2; y++ {
		setCell(grid, y, x, s.Render(ch))
	}
}

func (a *Animator) drawFace(grid [][]string, W, H int) {
	// ---- Sizing ----
	// Head is the central rectangle; make it fill most of the panel.
	headW := W - 4
	if headW > 22 {
		headW = 22
	}
	if headW < 10 {
		headW = 10
	}
	// Ensure odd width so center is clean
	if headW%2 == 0 {
		headW--
	}

	headH := H - 3 // leave room for antenna above and chin padding
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
	headTop := cy - headH/2 + 1
	headBot := headTop + headH - 1

	// ---- Styles ----
	headStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2E3D5A")).Bold(true)
	rimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5E82")).Bold(true)

	// ---- Antenna ----
	antRow := headTop - 2
	antCol := cx
	if antRow >= 0 {
		ledStyle := lipgloss.NewStyle().Foreground(a.ledColor()).Bold(true)
		setCell(grid, antRow, antCol, ledStyle.Render("◆"))
		setCell(grid, antRow+1, antCol, headStyle.Render("┃"))
	}

	// ---- Head border ----
	// Corners
	setCell(grid, headTop, headL, rimStyle.Render("╭"))
	setCell(grid, headTop, headR, rimStyle.Render("╮"))
	setCell(grid, headBot, headL, rimStyle.Render("╰"))
	setCell(grid, headBot, headR, rimStyle.Render("╯"))

	// Antenna port on top edge
	hline(grid, headTop, headL+1, headR-1, headStyle, "─")
	setCell(grid, headTop, antCol, rimStyle.Render("┬"))

	// Sides and bottom
	vline(grid, headL, headTop+1, headBot-1, headStyle, "│")
	vline(grid, headR, headTop+1, headBot-1, headStyle, "│")
	hline(grid, headBot, headL+1, headR-1, headStyle, "─")

	// ---- Bolt screws in head corners ----
	boltStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#334060"))
	setCell(grid, headTop+1, headL+1, boltStyle.Render("·"))
	setCell(grid, headTop+1, headR-1, boltStyle.Render("·"))
	setCell(grid, headBot-1, headL+1, boltStyle.Render("·"))
	setCell(grid, headBot-1, headR-1, boltStyle.Render("·"))

	// ---- Eyes ----
	eyeRow := headTop + 2
	eyeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2C3B56"))
	pupilStyle := lipgloss.NewStyle().Foreground(a.eyeColor()).Bold(true)

	// Eyes are 5 chars wide: ┌───┐
	//                        │ ● │
	//                        └───┘
	// Pupil can shift ±1 within the inner 3 chars

	// Eye X positions (centre of each eye)
	lEyeCX := cx - headW/4 - 1
	rEyeCX := cx + headW/4 + 1

	// Gaze offset: map gazeX (0–1) → -1,0,1
	gx := int(math.Round((a.gazeX*2 - 1) * 1.2)) // -1 to +1
	if gx < -1 {
		gx = -1
	}
	if gx > 1 {
		gx = 1
	}
	gy := int(math.Round((a.gazeY*2-1)*0.8)) // -1 to +1
	if gy < -1 {
		gy = -1
	}
	if gy > 1 {
		gy = 1
	}

	// Blink override
	blinking := a.blinkTimer%80 < 3

	// Left eyebrow
	lBrow, rBrow := a.eyebrows()
	browRow := eyeRow - 1
	setCell(grid, browRow, lEyeCX, lBrow)
	setCell(grid, browRow, rEyeCX, rBrow)

	// Draw each eye
	for _, ecx := range []int{lEyeCX, rEyeCX} {
		// Eye frame top
		setCell(grid, eyeRow, ecx-2, eyeStyle.Render("┌"))
		setCell(grid, eyeRow, ecx-1, eyeStyle.Render("─"))
		setCell(grid, eyeRow, ecx, eyeStyle.Render("─"))
		setCell(grid, eyeRow, ecx+1, eyeStyle.Render("─"))
		setCell(grid, eyeRow, ecx+2, eyeStyle.Render("┐"))

		// Eye frame sides
		setCell(grid, eyeRow+1, ecx-2, eyeStyle.Render("│"))
		setCell(grid, eyeRow+1, ecx+2, eyeStyle.Render("│"))

		// Eye frame bottom
		setCell(grid, eyeRow+2, ecx-2, eyeStyle.Render("└"))
		setCell(grid, eyeRow+2, ecx-1, eyeStyle.Render("─"))
		setCell(grid, eyeRow+2, ecx, eyeStyle.Render("─"))
		setCell(grid, eyeRow+2, ecx+1, eyeStyle.Render("─"))
		setCell(grid, eyeRow+2, ecx+2, eyeStyle.Render("┘"))

		if blinking {
			// Blink: replace interior with ─
			setCell(grid, eyeRow+1, ecx-1, eyeStyle.Render("─"))
			setCell(grid, eyeRow+1, ecx, eyeStyle.Render("─"))
			setCell(grid, eyeRow+1, ecx+1, eyeStyle.Render("─"))
		} else {
			p := a.pupilChar()
			// Pupil with gaze offset
			px := ecx + gx
			py := eyeRow + 1 + gy
			// Keep within eye bounds
			if px < ecx-1 {
				px = ecx - 1
			}
			if px > ecx+1 {
				px = ecx + 1
			}
			if py < eyeRow+1 {
				py = eyeRow + 1
			}
			if py > eyeRow+1 {
				py = eyeRow + 1
			}

			// Fill non-pupil interior with dim dots
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1C2538"))
			for dx := -1; dx <= 1; dx++ {
				setCell(grid, eyeRow+1, ecx+dx, dimStyle.Render("·"))
			}
			setCell(grid, py, px, pupilStyle.Render(string(p)))
		}
	}

	// ---- Nose ----
	noseRow := eyeRow + 3
	noseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A3550"))
	if noseRow < headBot-1 {
		setCell(grid, noseRow, cx-1, noseStyle.Render("▪"))
		setCell(grid, noseRow, cx+1, noseStyle.Render("▪"))
	}

	// ---- Mouth ----
	mouthRow := headBot - 2
	mouthW := headW - 6
	if mouthW < 4 {
		mouthW = 4
	}
	mouthLines := a.mouthRows(mouthW)
	for i, ml := range mouthLines {
		row := mouthRow + i
		if row >= headBot {
			break
		}
		mouthStartX := cx - lipgloss.Width(ml)/2
		for dx, ch := range ml {
			setCell(grid, row, mouthStartX+dx, string(ch))
		}
		// Actually render the whole mouth string centered
		setCell(grid, row, mouthStartX, ml)
		// Clear used columns so they don't double-render
		_ = mouthStartX
	}
	// Better: draw as one string at the center column
	if len(mouthLines) > 0 {
		ml := mouthLines[0]
		mlW := lipgloss.Width(ml)
		mouthX := cx - mlW/2
		// Clear the row first, then write
		for x := mouthX; x < mouthX+mlW && x < W; x++ {
			setCell(grid, mouthRow, x, " ")
		}
		setCell(grid, mouthRow, mouthX, ml)
	}

	// ---- Success burst ----
	if a.State == StateSuccess && a.burstFrame < 30 {
		stars := []string{"★", "✦", "✧", "✸"}
		starColor := lipgloss.NewStyle().Foreground(theme.DeimosGold).Bold(true)
		angles := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2, math.Pi / 4, 3 * math.Pi / 4, 5 * math.Pi / 4, 7 * math.Pi / 4}
		dist := float64(a.burstFrame) * 0.22
		for i, ang := range angles {
			bx := cx + int(math.Cos(ang)*dist*2)
			by := cy + int(math.Sin(ang)*dist)
			if bx >= 0 && bx < W && by >= 0 && by < len(grid) {
				setCell(grid, by, bx, starColor.Render(stars[i%len(stars)]))
			}
		}
	}

	// ---- Status dots on chin ----
	if headBot+0 < H {
		dotStyle := lipgloss.NewStyle().Foreground(a.stateColor()).Bold(true)
		period := 6
		for i := 0; i < 3; i++ {
			dotOn := a.Frame%period < period/2
			if a.State == StateIdle {
				dotOn = (a.Frame/20+i)%3 == 0
			}
			if dotOn {
				setCell(grid, headBot-1, cx-2+i*2, dotStyle.Render("·"))
			}
		}
	}
}

func (a *Animator) stateLabel() string {
	switch a.State {
	case StateBuilding:
		dots := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := dots[(a.Frame/2)%len(dots)]
		return fmt.Sprintf("%s BUILDING %s", spinner, spinner)
	case StateSuccess:
		return "✦ BUILD OK ✦"
	case StateError:
		return "✗ ERRORS"
	case StateRunning:
		frames := []string{"▶ RUN ·", "▶ RUN ··", "▶ RUN ···"}
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
		return theme.ColorOK
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
