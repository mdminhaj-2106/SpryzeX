package spryzex

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"spryzex-ide/internal/theme"
)

// BuildState controls animation mode
type BuildState int

const (
	StateIdle     BuildState = iota
	StateBuilding            // SPRYZEX spins fast, fire trails
	StateSuccess             // Celebration burst
	StateError               // Red pulse
	StateRunning             // Smooth orbit
)

// Animator holds all animation state
type Animator struct {
	State         BuildState
	StartTime     time.Time
	Frame         int
	spryzexAngleA float64 // SPRYZEX rotation angle A
	spryzexAngleB float64 // SPRYZEX rotation angle B
	phobosAngle   float64 // Phobos orbit angle
	deimosAngle   float64 // Deimos orbit angle
	fireFrame     int     // fire animation frame
	burstFrame    int     // success burst frame
	width         int
	height        int
}

// NewAnimator creates the SPRYZEX animator
func NewAnimator(w, h int) *Animator {
	return &Animator{
		State:     StateIdle,
		StartTime: time.Now(),
		width:     w,
		height:    h,
	}
}

// Tick advances animation by one frame
func (a *Animator) Tick() {
	a.Frame++

	// Speed varies by state
	speed := 0.04
	switch a.State {
	case StateBuilding:
		speed = 0.18 // crazy fast spin while building
	case StateSuccess:
		speed = 0.06
		a.burstFrame++
	case StateRunning:
		speed = 0.05
	}

	a.spryzexAngleA += speed
	a.spryzexAngleB += speed * 0.4

	// Phobos orbits fast (close to spryzex)
	phobosSpeed := 0.12
	if a.State == StateBuilding {
		phobosSpeed = 0.35
	}
	a.phobosAngle += phobosSpeed

	// Deimos orbits slow (far out)
	deimosSpeed := 0.04
	if a.State == StateBuilding {
		deimosSpeed = 0.12
	}
	a.deimosAngle += deimosSpeed

	a.fireFrame = (a.fireFrame + 1) % 8
}

// SetState changes animation state
func (a *Animator) SetState(s BuildState) {
	a.State = s
	a.burstFrame = 0
}

// renderSpryzex renders the 3D spinning SPRYZEX sphere
func (a *Animator) renderSpryzex(radius float64) [][]rune {
	// Sphere size in terminal chars
	rows := int(radius * 2.2)
	cols := int(radius * 4.4)
	if rows < 3 {
		rows = 3
	}
	if cols < 6 {
		cols = 6
	}

	// Output buffer
	buf := make([][]rune, rows)
	zbuf := make([][]float64, rows)
	for i := range buf {
		buf[i] = make([]rune, cols)
		zbuf[i] = make([]float64, cols)
		for j := range buf[i] {
			buf[i][j] = ' '
			zbuf[i][j] = -1e9
		}
	}

	// 3D sphere rendering — Andy Sloane style
	// Light source direction
	lx, ly, lz := 0.6, -0.4, 0.7
	lmag := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lmag, ly/lmag, lz/lmag

	cosA := math.Cos(a.spryzexAngleA)
	sinA := math.Sin(a.spryzexAngleA)
	cosB := math.Cos(a.spryzexAngleB)
	sinB := math.Sin(a.spryzexAngleB)

	// ASCII shading chars — bright to dark
	shading := []rune{'@', '#', 'S', '%', '?', '*', '+', ';', ':', ',', '.', ' '}

	for theta := 0.0; theta < 2*math.Pi; theta += 0.04 {
		for phi := 0.0; phi < 2*math.Pi; phi += 0.02 {
			sinTheta := math.Sin(theta)
			cosTheta := math.Cos(theta)
			sinPhi := math.Sin(phi)
			cosPhi := math.Cos(phi)

			// Point on sphere
			x := sinTheta * cosPhi
			y := sinTheta * sinPhi
			z := cosTheta

			// Rotate around Y (A) then X (B)
			x2 := x*cosA - z*sinA
			z2 := x*sinA + z*cosA
			y2 := y*cosB - z2*sinB
			z3 := y*sinB + z2*cosB

			// Project to screen
			zProj := z3 + 3.5
			if zProj <= 0 {
				continue
			}
			inv := 1.0 / zProj
			sx := int(float64(cols)/2 + radius*2.0*x2*inv)
			sy := int(float64(rows)/2 - radius*1.0*y2*inv)

			if sx < 0 || sx >= cols || sy < 0 || sy >= rows {
				continue
			}

			if z3 <= zbuf[sy][sx] {
				continue
			}
			zbuf[sy][sx] = z3

			// Normal at this point (same as position on unit sphere)
			// Dot product with light
			dot := x2*lx + y2*ly + z3*lz
			if dot < 0 {
				dot = 0
			}

			idx := int(dot * float64(len(shading)-2))
			if idx >= len(shading)-1 {
				idx = len(shading) - 2
			}
			buf[sy][sx] = shading[idx]
		}
	}

	return buf
}

// Render returns the full SPRYZEX panel as a styled string
func (a *Animator) Render(panelWidth, panelHeight int) string {
	var sb strings.Builder
	if panelWidth < 1 || panelHeight < 1 {
		return ""
	}

	infoLines := 0
	if panelHeight >= 10 {
		infoLines = 2
	} else if panelHeight >= 6 {
		infoLines = 1
	}
	artHeight := panelHeight - infoLines
	if artHeight < 3 {
		artHeight = panelHeight
		infoLines = 0
	}

	// Dynamic radius based on panel size
	radiusByHeight := float64(artHeight) * 0.30
	radiusByWidth := float64(panelWidth) * 0.24
	radius := math.Min(radiusByHeight, radiusByWidth)
	if radius < 4 {
		radius = 4
	}

	sphere := a.renderSpryzex(radius)
	sphereRows := len(sphere)
	var sphereCols int
	if sphereRows > 0 {
		sphereCols = len(sphere[0])
	}

	// Phobos: small, close, fast orbit
	phobosOrbitR := radius * 2.2
	phobosX := float64(panelWidth/2) + math.Cos(a.phobosAngle)*phobosOrbitR*2
	phobosY := float64(artHeight/2) + math.Sin(a.phobosAngle)*phobosOrbitR*0.5

	// Deimos: tiny, far, slow orbit
	deimosOrbitR := radius * 3.2
	deimosX := float64(panelWidth/2) + math.Cos(a.deimosAngle+math.Pi)*deimosOrbitR*2
	deimosY := float64(artHeight/2) + math.Sin(a.deimosAngle+math.Pi)*deimosOrbitR*0.5

	// Build char grid
	grid := make([][]string, artHeight)
	for i := range grid {
		grid[i] = make([]string, panelWidth)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// SPRYZEX sphere offset so it's centered
	sphereOffY := (artHeight - sphereRows) / 2
	sphereOffX := (panelWidth - sphereCols) / 2

	// Color function for SPRYZEX sphere chars
	spryzexColor := func(ch rune) string {
		s := string(ch)
		if ch == '@' || ch == '#' {
			return lipgloss.NewStyle().Foreground(theme.SpryzexRed).Bold(true).Render(s)
		} else if ch == 'S' || ch == '%' {
			return lipgloss.NewStyle().Foreground(theme.SpryzexBright).Render(s)
		} else if ch == '?' || ch == '*' {
			return lipgloss.NewStyle().Foreground(theme.SpryzexGlow).Render(s)
		} else if ch == '+' || ch == ';' {
			return lipgloss.NewStyle().Foreground(theme.SpryzexDust).Render(s)
		} else if ch != ' ' {
			return lipgloss.NewStyle().Foreground(theme.SpryzexDust).Render(s)
		}
		return " "
	}

	// Fill sphere
	for row := 0; row < sphereRows; row++ {
		for col := 0; col < sphereCols; col++ {
			gy := sphereOffY + row
			gx := sphereOffX + col
			if gy >= 0 && gy < artHeight && gx >= 0 && gx < panelWidth {
				ch := sphere[row][col]
				if ch != ' ' {
					grid[gy][gx] = spryzexColor(ch)
				}
			}
		}
	}

	// Orbit trails (dotted ellipses)
	phobosStyle := lipgloss.NewStyle().Foreground(theme.PhobosBlue)
	deimosStyle := lipgloss.NewStyle().Foreground(theme.DeimosGold)

	// Draw Phobos orbit dots
	for t := 0.0; t < 2*math.Pi; t += 0.15 {
		ox := int(float64(panelWidth/2) + math.Cos(t)*phobosOrbitR*2)
		oy := int(float64(artHeight/2) + math.Sin(t)*phobosOrbitR*0.5)
		if ox >= 0 && ox < panelWidth && oy >= 0 && oy < artHeight && grid[oy][ox] == " " {
			grid[oy][ox] = phobosStyle.Render("·")
		}
	}

	// Draw Deimos orbit dots
	for t := 0.0; t < 2*math.Pi; t += 0.25 {
		ox := int(float64(panelWidth/2) + math.Cos(t)*deimosOrbitR*2)
		oy := int(float64(artHeight/2) + math.Sin(t)*deimosOrbitR*0.5)
		if ox >= 0 && ox < panelWidth && oy >= 0 && oy < artHeight && grid[oy][ox] == " " {
			grid[oy][ox] = deimosStyle.Render("·")
		}
	}

	// Draw Phobos moon
	py := int(phobosY)
	px := int(phobosX)
	if px >= 0 && px < panelWidth-1 && py >= 0 && py < artHeight {
		moonStr := phobosStyle.Bold(true).Render("◉")
		grid[py][px] = moonStr
	}

	// Draw Deimos moon
	dy := int(deimosY)
	dx := int(deimosX)
	if dx >= 0 && dx < panelWidth-1 && dy >= 0 && dy < artHeight {
		moonStr := deimosStyle.Bold(true).Render("●")
		grid[dy][dx] = moonStr
	}

	// Fire particles when building
	if a.State == StateBuilding {
		a.drawFire(grid, panelWidth, artHeight, sphereOffX, sphereOffY, sphereRows, sphereCols)
	}

	// Success burst
	if a.State == StateSuccess {
		a.drawBurst(grid, panelWidth, artHeight)
	}

	// Error pulse
	if a.State == StateError {
		a.drawErrorPulse(grid, panelWidth, artHeight)
	}

	// Render grid to string
	for row := 0; row < artHeight; row++ {
		for col := 0; col < panelWidth; col++ {
			sb.WriteString(grid[row][col])
		}
		if row < artHeight-1 || infoLines > 0 {
			sb.WriteRune('\n')
		}
	}

	if infoLines >= 1 {
		sb.WriteString(lipgloss.NewStyle().
			Width(panelWidth).
			Align(lipgloss.Center).
			Foreground(a.stateColor()).
			Bold(true).
			Render(a.stateLabel()))
	}

	if infoLines >= 2 {
		sb.WriteRune('\n')
		sb.WriteString(lipgloss.NewStyle().
			Width(panelWidth).
			Align(lipgloss.Center).
			Foreground(theme.TextMuted).
			Render("Phobos / Deimos"))
	}

	return sb.String()
}

var fireChars = []rune{'▲', '△', '∧', '^', '`', '\'', ' '}
var fireColors = []lipgloss.Color{"#FF0000", "#FF3300", "#FF6600", "#FF9900", "#FFCC00", "#FFFF00"}

func (a *Animator) drawFire(grid [][]string, pw, ph, sox, soy, sr, sc int) {
	// Emit fire particles below the sphere
	baseY := soy + sr - 1
	baseXStart := sox + sc/4
	baseXEnd := sox + 3*sc/4

	for x := baseXStart; x <= baseXEnd; x++ {
		if x < 0 || x >= pw {
			continue
		}
		// How many rows of fire
		fireH := int(3 + math.Sin(float64(a.fireFrame+x)*0.7)*2)
		for fy := 0; fy < fireH; fy++ {
			gy := baseY + 1 + fy
			if gy < 0 || gy >= ph {
				continue
			}
			charIdx := fy
			if charIdx >= len(fireChars)-1 {
				charIdx = len(fireChars) - 2
			}
			colorIdx := fy
			if colorIdx >= len(fireColors) {
				colorIdx = len(fireColors) - 1
			}
			flickerIdx := (a.fireFrame + x + fy) % len(fireChars)
			ch := fireChars[flickerIdx%len(fireChars)]
			if ch == ' ' {
				continue
			}
			grid[gy][x] = lipgloss.NewStyle().
				Foreground(fireColors[colorIdx]).
				Bold(true).
				Render(string(ch))
		}
	}
}

var burstChars = []string{"✦", "✧", "★", "☆", "*", "+", "·"}
var burstColors = []lipgloss.Color{
	theme.DeimosGold, theme.SpryzexGlow, theme.AuroraGreen,
	theme.PhobosBlue, theme.NebulaPurp, theme.CometCyan,
}

func (a *Animator) drawBurst(grid [][]string, pw, ph int) {
	cx := pw / 2
	cy := ph / 2
	frame := float64(a.burstFrame)

	for i := 0; i < 12; i++ {
		angle := float64(i) * math.Pi / 6.0
		dist := frame * 0.4
		x := int(float64(cx) + math.Cos(angle)*dist*2)
		y := int(float64(cy) + math.Sin(angle)*dist*0.5)
		if x >= 0 && x < pw && y >= 0 && y < ph {
			ch := burstChars[i%len(burstChars)]
			color := burstColors[i%len(burstColors)]
			grid[y][x] = lipgloss.NewStyle().Foreground(color).Bold(true).Render(ch)
		}
	}
}

func (a *Animator) drawErrorPulse(grid [][]string, pw, ph int) {
	pulse := math.Sin(float64(a.Frame) * 0.3)
	if pulse > 0.5 {
		// Flash red X markers at corners of spryzex
		markers := [][2]int{
			{ph/2 - 3, pw/2 - 6}, {ph/2 - 3, pw/2 + 6},
			{ph/2 + 3, pw/2 - 6}, {ph/2 + 3, pw/2 + 6},
		}
		for _, m := range markers {
			if m[0] >= 0 && m[0] < ph && m[1] >= 0 && m[1] < pw {
				grid[m[0]][m[1]] = lipgloss.NewStyle().
					Foreground(theme.ColorError).Bold(true).Render("✗")
			}
		}
	}
}

func (a *Animator) stateLabel() string {
	switch a.State {
	case StateBuilding:
		frames := []string{"ASSEMBLING ▪▪▪", "ASSEMBLING ▪▪ ", "ASSEMBLING ▪  ", "ASSEMBLING    "}
		return frames[(a.Frame/3)%len(frames)]
	case StateSuccess:
		return "✓ BUILD OK — SPRYZEX IS PLEASED"
	case StateError:
		return "✗ BUILD FAILED"
	case StateRunning:
		return "▶ EMULATOR RUNNING"
	default:
		return "SPRYZEX IDE  ·  READY"
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
		return theme.TextSecond
	}
}

// BuildingFrames returns animation frames as a tea.Cmd tick
func TickCmd() func() (string, error) {
	return func() (string, error) {
		time.Sleep(50 * time.Millisecond)
		return fmt.Sprintf("tick:%d", time.Now().UnixMilli()), nil
	}
}
