package spryzex

import (
	"fmt"
	"math"
	"math/rand"
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

type star struct {
	x, y    int
	twinkle int
}

type Animator struct {
	State         BuildState
	StartTime     time.Time
	Frame         int
	spryzexAngleA float64
	spryzexAngleB float64
	phobosAngle   float64
	deimosAngle   float64
	fireFrame     int
	burstFrame    int
	width         int
	height        int
	stars         []star
}

func NewAnimator(w, h int) *Animator {
	a := &Animator{
		State:     StateIdle,
		StartTime: time.Now(),
		width:     w,
		height:    h,
	}
	a.initStars(w, h)
	return a
}

func (a *Animator) initStars(w, h int) {
	a.stars = a.stars[:0]
	count := (w * h) / 18
	if count > 50 {
		count = 50
	}
	for i := 0; i < count; i++ {
		a.stars = append(a.stars, star{
			x:       rand.Intn(max1(w)),
			y:       rand.Intn(max1(h)),
			twinkle: rand.Intn(30),
		})
	}
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

func (a *Animator) Tick() {
	a.Frame++

	speed := 0.025
	switch a.State {
	case StateBuilding:
		speed = 0.12
	case StateSuccess:
		speed = 0.04
		a.burstFrame++
	case StateRunning:
		speed = 0.035
	}

	a.spryzexAngleA += speed
	a.spryzexAngleB += speed * 0.41

	phobosSpeed := 0.08
	if a.State == StateBuilding {
		phobosSpeed = 0.25
	}
	a.phobosAngle += phobosSpeed

	deimosSpeed := 0.03
	if a.State == StateBuilding {
		deimosSpeed = 0.08
	}
	a.deimosAngle += deimosSpeed

	a.fireFrame = (a.fireFrame + 1) % 8
}

func (a *Animator) SetState(s BuildState) {
	a.State = s
	a.burstFrame = 0
}

// sphereColor maps a Lambertian dot product [0,1] to a smooth orange/red background color.
func sphereColor(dot float64) lipgloss.Color {
	type rgb struct{ r, g, b float64 }

	stops := []struct {
		t float64
		c rgb
	}{
		{0.00, rgb{0.06, 0.01, 0.01}}, // deep shadow
		{0.18, rgb{0.20, 0.05, 0.02}}, // dark
		{0.38, rgb{0.50, 0.16, 0.05}}, // terminator
		{0.60, rgb{0.78, 0.36, 0.09}}, // mid-lit
		{0.80, rgb{0.94, 0.60, 0.20}}, // lit
		{1.00, rgb{1.00, 0.87, 0.60}}, // highlight
	}

	var c rgb
	for i := 1; i < len(stops); i++ {
		if dot <= stops[i].t {
			lo := stops[i-1]
			hi := stops[i]
			t := (dot - lo.t) / (hi.t - lo.t)
			c.r = lo.c.r + t*(hi.c.r-lo.c.r)
			c.g = lo.c.g + t*(hi.c.g-lo.c.g)
			c.b = lo.c.b + t*(hi.c.b-lo.c.b)
			break
		}
	}
	if dot > 1.0 {
		c = stops[len(stops)-1].c
	}

	ri := int(math.Min(c.r*255, 255))
	gi := int(math.Min(c.g*255, 255))
	bi := int(math.Min(c.b*255, 255))
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", ri, gi, bi))
}

// renderSphere returns a 2-D grid of background colors (empty string = no sphere here).
func (a *Animator) renderSphere(radius float64) ([][]lipgloss.Color, int, int) {
	cols := int(radius * 4.0)
	rows := int(radius * 2.2)
	if rows < 4 {
		rows = 4
	}
	if cols < 8 {
		cols = 8
	}

	colors := make([][]lipgloss.Color, rows)
	zbuf := make([][]float64, rows)
	for i := range colors {
		colors[i] = make([]lipgloss.Color, cols)
		zbuf[i] = make([]float64, cols)
		for j := range zbuf[i] {
			zbuf[i][j] = -1e9
		}
	}

	// Light direction — upper-left-front
	lx, ly, lz := 0.55, -0.55, 0.62
	lmag := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lmag, ly/lmag, lz/lmag

	cosA := math.Cos(a.spryzexAngleA)
	sinA := math.Sin(a.spryzexAngleA)
	cosB := math.Cos(a.spryzexAngleB)
	sinB := math.Sin(a.spryzexAngleB)

	for theta := 0.0; theta < 2*math.Pi; theta += 0.030 {
		for phi := 0.0; phi < 2*math.Pi; phi += 0.016 {
			sinT, cosT := math.Sin(theta), math.Cos(theta)
			sinP, cosP := math.Sin(phi), math.Cos(phi)

			// Point on unit sphere
			x := sinT * cosP
			y := sinT * sinP
			z := cosT

			// Rotate around Y then X
			x2 := x*cosA - z*sinA
			z2 := x*sinA + z*cosA
			y2 := y*cosB - z2*sinB
			z3 := y*sinB + z2*cosB

			// Project (orthographic) — compensate for 2:1 char aspect ratio
			sx := int(float64(cols)/2.0 + radius*2.0*x2)
			sy := int(float64(rows)/2.0 - radius*1.0*y2)

			if sx < 0 || sx >= cols || sy < 0 || sy >= rows {
				continue
			}

			if z3 <= zbuf[sy][sx] {
				continue
			}
			zbuf[sy][sx] = z3

			// Lambertian diffuse
			dot := x2*lx + y2*ly + z3*lz
			if dot < 0 {
				dot = 0
			}

			// Small specular spike
			hy := ly
			hz := lz + 1.0
			hlen := math.Sqrt(lx*lx + hy*hy + hz*hz)
			spec := x2*lx/hlen + y2*hy/hlen + z3*hz/hlen
			if spec < 0 {
				spec = 0
			}
			dot += math.Pow(spec, 12) * 0.45
			if dot > 1.0 {
				dot = 1.0
			}

			colors[sy][sx] = sphereColor(dot)
		}
	}

	return colors, rows, cols
}

func (a *Animator) Render(panelWidth, panelHeight int) string {
	var sb strings.Builder
	if panelWidth < 1 || panelHeight < 1 {
		return ""
	}

	if len(a.stars) == 0 || panelWidth != a.width || panelHeight != a.height {
		a.width = panelWidth
		a.height = panelHeight
		a.initStars(panelWidth, panelHeight)
	}

	infoLines := 0
	if panelHeight >= 10 {
		infoLines = 2
	} else if panelHeight >= 6 {
		infoLines = 1
	}
	artHeight := panelHeight - infoLines
	if artHeight < 4 {
		artHeight = panelHeight
		infoLines = 0
	}

	radiusByH := float64(artHeight) * 0.23
	radiusByW := float64(panelWidth) * 0.18
	radius := math.Min(radiusByH, radiusByW)
	if radius < 3.5 {
		radius = 3.5
	}

	sphereColors, sphereRows, sphereCols := a.renderSphere(radius)

	// Build a string grid: each cell is a single pre-rendered string
	grid := make([][]string, artHeight)
	for i := range grid {
		grid[i] = make([]string, panelWidth)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	a.drawStarfield(grid, panelWidth, artHeight)

	sphereOffY := (artHeight - sphereRows) / 2
	sphereOffX := (panelWidth - sphereCols) / 2

	phobosOrbitRx := radius * 2.5
	phobosOrbitRy := radius * 0.55
	deimosOrbitRx := radius * 3.8
	deimosOrbitRy := radius * 0.80

	cx := float64(panelWidth) / 2.0
	cy := float64(artHeight) / 2.0

	phobosStyle := lipgloss.NewStyle().Foreground(theme.PhobosBlue)
	deimosStyle := lipgloss.NewStyle().Foreground(theme.DeimosGold)

	// Draw orbit trails (behind sphere)
	for t := 0.0; t < 2*math.Pi; t += 0.10 {
		ox := int(cx + math.Cos(t)*phobosOrbitRx)
		oy := int(cy + math.Sin(t)*phobosOrbitRy)
		if ox >= 0 && ox < panelWidth && oy >= 0 && oy < artHeight && grid[oy][ox] == " " {
			grid[oy][ox] = phobosStyle.Render("·")
		}
	}
	for t := 0.0; t < 2*math.Pi; t += 0.16 {
		ox := int(cx + math.Cos(t)*deimosOrbitRx)
		oy := int(cy + math.Sin(t)*deimosOrbitRy)
		if ox >= 0 && ox < panelWidth && oy >= 0 && oy < artHeight && grid[oy][ox] == " " {
			grid[oy][ox] = deimosStyle.Render("·")
		}
	}

	// Draw sphere — background-colored spaces for smooth 3D shading
	for row := 0; row < sphereRows; row++ {
		for col := 0; col < sphereCols; col++ {
			gy := sphereOffY + row
			gx := sphereOffX + col
			if gy < 0 || gy >= artHeight || gx < 0 || gx >= panelWidth {
				continue
			}
			c := sphereColors[row][col]
			if c == "" {
				continue
			}
			grid[gy][gx] = lipgloss.NewStyle().Background(c).Render(" ")
		}
	}

	// Draw moons on top
	phobosX := cx + math.Cos(a.phobosAngle)*phobosOrbitRx
	phobosY := cy + math.Sin(a.phobosAngle)*phobosOrbitRy
	deimosX := cx + math.Cos(a.deimosAngle+math.Pi)*deimosOrbitRx
	deimosY := cy + math.Sin(a.deimosAngle+math.Pi)*deimosOrbitRy

	py, px := int(phobosY), int(phobosX)
	if px >= 0 && px < panelWidth && py >= 0 && py < artHeight {
		grid[py][px] = phobosStyle.Bold(true).Render("◉")
	}
	dy, dx := int(deimosY), int(deimosX)
	if dx >= 0 && dx < panelWidth && dy >= 0 && dy < artHeight {
		grid[dy][dx] = deimosStyle.Bold(true).Render("◉")
	}

	// State effects
	switch a.State {
	case StateBuilding:
		a.drawFire(grid, panelWidth, artHeight, sphereOffX, sphereOffY, sphereRows, sphereCols)
	case StateSuccess:
		a.drawBurst(grid, panelWidth, artHeight)
	case StateError:
		a.drawErrorPulse(grid, panelWidth, artHeight)
	}

	// Assemble rows
	for row := 0; row < artHeight; row++ {
		for col := 0; col < panelWidth; col++ {
			sb.WriteString(grid[row][col])
		}
		if row < artHeight-1 || infoLines > 0 {
			sb.WriteRune('\n')
		}
	}

	// Info lines
	if infoLines >= 1 {
		label := a.stateLabel()
		sb.WriteString(lipgloss.NewStyle().
			Width(panelWidth).
			Align(lipgloss.Center).
			Foreground(a.stateColor()).
			Bold(true).
			Render(label))
	}
	if infoLines >= 2 {
		sb.WriteRune('\n')
		phText := lipgloss.NewStyle().Foreground(theme.PhobosBlue).Bold(true).Render("◉ Phobos")
		dmText := lipgloss.NewStyle().Foreground(theme.DeimosGold).Bold(true).Render("◉ Deimos")
		sep := lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ·  ")
		moons := phText + sep + dmText
		moonsW := lipgloss.Width(moons)
		padLeft := (panelWidth - moonsW) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		sb.WriteString(strings.Repeat(" ", padLeft))
		sb.WriteString(moons)
	}

	return sb.String()
}

func (a *Animator) drawStarfield(grid [][]string, pw, ph int) {
	starChars := []string{"·", "⋆", "∘", "⋅", "·"}
	starColors := []lipgloss.Color{"#2A3060", "#1E2545", "#252840", "#1A1D35", "#303870"}
	for _, s := range a.stars {
		if s.x >= pw || s.y >= ph {
			continue
		}
		if grid[s.y][s.x] != " " {
			continue
		}
		twinklePhase := (a.Frame + s.twinkle) % 40
		if twinklePhase > 26 {
			continue
		}
		charIdx := (s.x + s.y) % len(starChars)
		colorIdx := (s.x*3 + s.y*7) % len(starColors)
		// Brighter twinkle
		if twinklePhase > 22 {
			colorIdx = 4 // brightest
		}
		grid[s.y][s.x] = lipgloss.NewStyle().
			Foreground(starColors[colorIdx]).
			Render(starChars[charIdx])
	}
}

var fireChars = []rune{'▲', '△', '∧', '˄', ' ', ' '}
var fireColors = []lipgloss.Color{"#FF1100", "#FF3300", "#FF6600", "#FF9900", "#FFCC00", "#FFEE44"}

func (a *Animator) drawFire(grid [][]string, pw, ph, sox, soy, sr, sc int) {
	baseY := soy + sr
	baseXStart := sox + sc/4
	baseXEnd := sox + 3*sc/4

	for x := baseXStart; x <= baseXEnd; x++ {
		if x < 0 || x >= pw {
			continue
		}
		fireH := int(3 + math.Sin(float64(a.fireFrame+x)*0.8)*2)
		for fy := 0; fy < fireH; fy++ {
			gy := baseY + fy
			if gy < 0 || gy >= ph {
				continue
			}
			flickerIdx := (a.fireFrame + x + fy) % len(fireChars)
			ch := fireChars[flickerIdx]
			if ch == ' ' {
				continue
			}
			colorIdx := fy
			if colorIdx >= len(fireColors) {
				colorIdx = len(fireColors) - 1
			}
			grid[gy][x] = lipgloss.NewStyle().
				Foreground(fireColors[colorIdx]).
				Bold(true).
				Render(string(ch))
		}
	}
}

var burstChars = []string{"✦", "✧", "★", "✸", "✹", "✺", "✼", "·"}
var burstColors = []lipgloss.Color{
	theme.DeimosGold, theme.SpryzexGlow, theme.AuroraGreen,
	theme.PhobosBlue, theme.NebulaPurp, theme.CometCyan,
	theme.SpryzexBright, theme.DeimosGold,
}

func (a *Animator) drawBurst(grid [][]string, pw, ph int) {
	cx := pw / 2
	cy := ph / 2
	frame := float64(a.burstFrame)

	for i := 0; i < 16; i++ {
		angle := float64(i) * math.Pi / 8.0
		dist := frame * 0.35
		x := int(float64(cx) + math.Cos(angle)*dist*2.0)
		y := int(float64(cy) + math.Sin(angle)*dist*0.6)
		if x >= 0 && x < pw && y >= 0 && y < ph {
			ch := burstChars[i%len(burstChars)]
			color := burstColors[i%len(burstColors)]
			grid[y][x] = lipgloss.NewStyle().Foreground(color).Bold(true).Render(ch)
		}
	}
}

func (a *Animator) drawErrorPulse(grid [][]string, pw, ph int) {
	pulse := math.Sin(float64(a.Frame) * 0.25)
	if pulse > 0.3 {
		markers := [][2]int{
			{ph/2 - 2, pw/2 - 5}, {ph/2 - 2, pw/2 + 5},
			{ph/2 + 2, pw/2 - 5}, {ph/2 + 2, pw/2 + 5},
		}
		intensity := lipgloss.Color("#BF616A")
		if pulse > 0.7 {
			intensity = lipgloss.Color("#FF4455")
		}
		for _, m := range markers {
			if m[0] >= 0 && m[0] < ph && m[1] >= 0 && m[1] < pw {
				grid[m[0]][m[1]] = lipgloss.NewStyle().
					Foreground(intensity).Bold(true).Render("✗")
			}
		}
	}
}

func (a *Animator) stateLabel() string {
	switch a.State {
	case StateBuilding:
		dots := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := dots[(a.Frame/2)%len(dots)]
		return fmt.Sprintf("%s ASSEMBLING %s", spinner, spinner)
	case StateSuccess:
		return "✦ BUILD OK — SPRYZEX PLEASED ✦"
	case StateError:
		return "✗ BUILD FAILED"
	case StateRunning:
		frames := []string{"▶ RUNNING ·", "▶ RUNNING ··", "▶ RUNNING ···"}
		return frames[(a.Frame/4)%len(frames)]
	default:
		return "◈ SPRYZEX  ·  READY"
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
		return theme.SpryzexBright
	}
}

func TickCmd() func() (string, error) {
	return func() (string, error) {
		time.Sleep(50 * time.Millisecond)
		return fmt.Sprintf("tick:%d", time.Now().UnixMilli()), nil
	}
}
