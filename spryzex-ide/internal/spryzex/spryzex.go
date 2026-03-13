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
	count := (w * h) / 20
	if count > 40 {
		count = 40
	}
	for i := 0; i < count; i++ {
		a.stars = append(a.stars, star{
			x:       rand.Intn(max1(w)),
			y:       rand.Intn(max1(h)),
			twinkle: rand.Intn(20),
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

	speed := 0.03
	switch a.State {
	case StateBuilding:
		speed = 0.15
	case StateSuccess:
		speed = 0.05
		a.burstFrame++
	case StateRunning:
		speed = 0.04
	}

	a.spryzexAngleA += speed
	a.spryzexAngleB += speed * 0.37

	phobosSpeed := 0.10
	if a.State == StateBuilding {
		phobosSpeed = 0.30
	}
	a.phobosAngle += phobosSpeed

	deimosSpeed := 0.035
	if a.State == StateBuilding {
		deimosSpeed = 0.10
	}
	a.deimosAngle += deimosSpeed

	a.fireFrame = (a.fireFrame + 1) % 8
}

func (a *Animator) SetState(s BuildState) {
	a.State = s
	a.burstFrame = 0
}

func (a *Animator) renderSphere(radius float64) ([][]rune, [][]int) {
	cols := int(radius * 4.0)
	rows := int(radius * 2.0)
	if rows < 3 {
		rows = 3
	}
	if cols < 6 {
		cols = 6
	}

	buf := make([][]rune, rows)
	light := make([][]int, rows)
	zbuf := make([][]float64, rows)
	for i := range buf {
		buf[i] = make([]rune, cols)
		light[i] = make([]int, cols)
		zbuf[i] = make([]float64, cols)
		for j := range buf[i] {
			buf[i][j] = ' '
			zbuf[i][j] = -1e9
		}
	}

	lx, ly, lz := 0.7, -0.5, 0.5
	lmag := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lmag, ly/lmag, lz/lmag

	cosA := math.Cos(a.spryzexAngleA)
	sinA := math.Sin(a.spryzexAngleA)
	cosB := math.Cos(a.spryzexAngleB)
	sinB := math.Sin(a.spryzexAngleB)

	shading := []rune{'█', '▓', '▓', '▒', '▒', '░', '·'}

	for theta := 0.0; theta < 2*math.Pi; theta += 0.035 {
		for phi := 0.0; phi < 2*math.Pi; phi += 0.018 {
			sinTheta := math.Sin(theta)
			cosTheta := math.Cos(theta)
			sinPhi := math.Sin(phi)
			cosPhi := math.Cos(phi)

			x := sinTheta * cosPhi
			y := sinTheta * sinPhi
			z := cosTheta

			x2 := x*cosA - z*sinA
			z2 := x*sinA + z*cosA
			y2 := y*cosB - z2*sinB
			z3 := y*sinB + z2*cosB

			sx := int(float64(cols)/2.0 + radius*2.0*x2)
			sy := int(float64(rows)/2.0 - radius*y2)

			if sx < 0 || sx >= cols || sy < 0 || sy >= rows {
				continue
			}

			if z3 <= zbuf[sy][sx] {
				continue
			}
			zbuf[sy][sx] = z3

			dot := x2*lx + y2*ly + z3*lz
			if dot < 0 {
				dot = 0
			}

			idx := int(dot * float64(len(shading)-1))
			if idx >= len(shading) {
				idx = len(shading) - 1
			}
			buf[sy][sx] = shading[idx]
			light[sy][sx] = idx
		}
	}

	return buf, light
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
	if artHeight < 3 {
		artHeight = panelHeight
		infoLines = 0
	}

	radiusByH := float64(artHeight) * 0.26
	radiusByW := float64(panelWidth) * 0.20
	radius := math.Min(radiusByH, radiusByW)
	if radius < 3 {
		radius = 3
	}

	sphere, lightmap := a.renderSphere(radius)
	sphereRows := len(sphere)
	sphereCols := 0
	if sphereRows > 0 {
		sphereCols = len(sphere[0])
	}

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

	phobosOrbitRx := radius * 2.4
	phobosOrbitRy := radius * 0.55
	deimosOrbitRx := radius * 3.6
	deimosOrbitRy := radius * 0.75

	cx := float64(panelWidth) / 2.0
	cy := float64(artHeight) / 2.0

	phobosStyle := lipgloss.NewStyle().Foreground(theme.PhobosBlue)
	deimosStyle := lipgloss.NewStyle().Foreground(theme.DeimosGold)

	for t := 0.0; t < 2*math.Pi; t += 0.12 {
		ox := int(cx + math.Cos(t)*phobosOrbitRx)
		oy := int(cy + math.Sin(t)*phobosOrbitRy)
		if ox >= 0 && ox < panelWidth && oy >= 0 && oy < artHeight && grid[oy][ox] == " " {
			grid[oy][ox] = phobosStyle.Render("·")
		}
	}

	for t := 0.0; t < 2*math.Pi; t += 0.18 {
		ox := int(cx + math.Cos(t)*deimosOrbitRx)
		oy := int(cy + math.Sin(t)*deimosOrbitRy)
		if ox >= 0 && ox < panelWidth && oy >= 0 && oy < artHeight && grid[oy][ox] == " " {
			grid[oy][ox] = deimosStyle.Render("·")
		}
	}

	for row := 0; row < sphereRows; row++ {
		for col := 0; col < sphereCols; col++ {
			gy := sphereOffY + row
			gx := sphereOffX + col
			if gy < 0 || gy >= artHeight || gx < 0 || gx >= panelWidth {
				continue
			}
			ch := sphere[row][col]
			if ch == ' ' {
				continue
			}
			li := lightmap[row][col]
			var color lipgloss.Color
			switch {
			case li == 0:
				color = theme.SpryzexGlow
			case li <= 1:
				color = theme.SpryzexBright
			case li <= 3:
				color = theme.SpryzexRed
			default:
				color = theme.SpryzexDust
			}
			style := lipgloss.NewStyle().Foreground(color)
			if li <= 1 {
				style = style.Bold(true)
			}
			grid[gy][gx] = style.Render(string(ch))
		}
	}

	phobosX := cx + math.Cos(a.phobosAngle)*phobosOrbitRx
	phobosY := cy + math.Sin(a.phobosAngle)*phobosOrbitRy
	deimosX := cx + math.Cos(a.deimosAngle+math.Pi)*deimosOrbitRx
	deimosY := cy + math.Sin(a.deimosAngle+math.Pi)*deimosOrbitRy

	py := int(phobosY)
	px := int(phobosX)
	if px >= 0 && px < panelWidth && py >= 0 && py < artHeight {
		grid[py][px] = phobosStyle.Bold(true).Render("◉")
	}

	dy := int(deimosY)
	dx := int(deimosX)
	if dx >= 0 && dx < panelWidth && dy >= 0 && dy < artHeight {
		grid[dy][dx] = deimosStyle.Bold(true).Render("◉")
	}

	switch a.State {
	case StateBuilding:
		a.drawFire(grid, panelWidth, artHeight, sphereOffX, sphereOffY, sphereRows, sphereCols)
	case StateSuccess:
		a.drawBurst(grid, panelWidth, artHeight)
	case StateError:
		a.drawErrorPulse(grid, panelWidth, artHeight)
	}

	for row := 0; row < artHeight; row++ {
		for col := 0; col < panelWidth; col++ {
			sb.WriteString(grid[row][col])
		}
		if row < artHeight-1 || infoLines > 0 {
			sb.WriteRune('\n')
		}
	}

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
	starChars := []string{"⋆", "·", "∘", "⋅"}
	starColors := []lipgloss.Color{"#2A3050", "#252840", "#1E2238", "#1A1D30"}
	for _, s := range a.stars {
		if s.x >= pw || s.y >= ph {
			continue
		}
		if grid[s.y][s.x] != " " {
			continue
		}
		twinklePhase := (a.Frame + s.twinkle) % 30
		if twinklePhase > 20 {
			continue
		}
		charIdx := (s.x + s.y) % len(starChars)
		colorIdx := (s.x*3 + s.y*7) % len(starColors)
		grid[s.y][s.x] = lipgloss.NewStyle().
			Foreground(starColors[colorIdx]).
			Render(starChars[charIdx])
	}
}

var fireChars = []rune{'▲', '△', '∧', '˄', '`', ' '}
var fireColors = []lipgloss.Color{"#FF0000", "#FF3300", "#FF6600", "#FF9900", "#FFCC00", "#FFEE44"}

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
