package tuichart

import "math"

type Point struct{ X, Y float64 }

func Seq(vals ...float64) []Point {
	out := make([]Point, len(vals))
	for i, v := range vals {
		out[i] = Point{X: float64(i), Y: v}
	}
	return out
}

func Zip(xs, ys []float64) []Point {
	n := minInt(len(xs), len(ys))
	out := make([]Point, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Point{X: xs[i], Y: ys[i]})
	}
	return out
}

type seriesI interface {
	seriesName() string
	bounds(db *dataBounds)
	hasColor() bool
	setColor(c Color)
	colorOf() Color
	draw(cv *Canvas, fr frame, st Style)
	legendEntry(st Style, uni bool) LegendEntry
}

type Line struct {
	name   string
	pts    []Point
	color  Color
	marker rune
	dashed bool
}

func NewLine(name string, pts ...Point) *Line {
	return &Line{name: name, pts: pts}
}

func NewLineVals(name string, vals []float64) *Line {
	return &Line{name: name, pts: Seq(vals...)}
}

func (l *Line) Points(pts ...Point) *Line    { l.pts = append(l.pts, pts...); return l }
func (l *Line) Values(vals ...float64) *Line { l.pts = append(l.pts, Seq(vals...)...); return l }

// SetValues replaces all points with the given y values at x = 0..n-1.
// Intended for live charts: assign Ring.Values() on every tick.
func (l *Line) SetValues(vals []float64) *Line { l.pts = Seq(vals...); return l }
func (l *Line) Color(c Color) *Line            { l.color = c; return l }
func (l *Line) Marker(m rune) *Line            { l.marker = m; return l }
func (l *Line) Dashed(d bool) *Line            { l.dashed = d; return l }
func (l *Line) SetName(s string) *Line         { l.name = s; return l }

func (l *Line) seriesName() string { return l.name }

func (l *Line) bounds(db *dataBounds) {
	for _, p := range l.pts {
		db.add(p.X, p.Y)
	}
}
func (l *Line) hasColor() bool { return !l.color.IsZero() }
func (l *Line) setColor(c Color) {
	if l.color.IsZero() {
		l.color = c
	}
}
func (l *Line) colorOf() Color { return l.color }

type projPt struct{ x, y float64 }

func project(fr frame, p Point) projPt {
	return projPt{x: fr.mx(p.X), y: fr.my(p.Y)}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (l *Line) draw(cv *Canvas, fr frame, st Style) {
	if len(l.pts) == 0 {
		return
	}
	proj := make([]projPt, len(l.pts))
	for i, p := range l.pts {
		proj[i] = project(fr, p)
	}
	area := fr.area
	for i := 0; i < len(proj)-1; i++ {
		a, b := proj[i], proj[i+1]
		if math.IsNaN(a.x) || math.IsNaN(a.y) || math.IsNaN(b.x) || math.IsNaN(b.y) {
			continue
		}
		if fr.uni && fr.ysc.Kind == Linear && fr.xsc.Kind == Linear {
			gx0 := int(math.Round(clampF(a.x, float64(area.X-1), float64(area.X2()+1)) * 2))
			gy0 := int(math.Round(clampF(a.y, float64(area.Y-1), float64(area.Y2()+1)) * 4))
			gx1 := int(math.Round(clampF(b.x, float64(area.X-1), float64(area.X2()+1)) * 2))
			gy1 := int(math.Round(clampF(b.y, float64(area.Y-1), float64(area.Y2()+1)) * 4))
			dotLine(cv, gx0, gy0, gx1, gy1, st.Fg, l.dashed)
		} else {
			x0 := int(math.Round(clampF(a.x, float64(area.X), float64(area.X2()))))
			y0 := int(math.Round(clampF(a.y, float64(area.Y), float64(area.Y2()))))
			x1 := int(math.Round(clampF(b.x, float64(area.X), float64(area.X2()))))
			y1 := int(math.Round(clampF(b.y, float64(area.Y), float64(area.Y2()))))
			asciiLine(cv, x0, y0, x1, y1, st)
		}
	}
	l.drawMarkers(cv, fr, st)
}

func (l *Line) drawMarkers(cv *Canvas, fr frame, st Style) {
	if l.marker == 0 {
		return
	}
	m := l.marker
	area := fr.area
	for _, p := range l.pts {
		q := project(fr, p)
		if math.IsNaN(q.x) || math.IsNaN(q.y) {
			continue
		}
		x := int(math.Round(clampF(q.x, float64(area.X), float64(area.X2()))))
		y := int(math.Round(clampF(q.y, float64(area.Y), float64(area.Y2()))))
		cv.Set(x, y, m, st)
	}
}

func (l *Line) legendEntry(st Style, uni bool) LegendEntry {
	glyph := "───"
	if !uni {
		glyph = "---"
	}
	return LegendEntry{Label: l.name, Style: st, Glyph: glyph}
}

type Scatter struct {
	name   string
	pts    []Point
	color  Color
	marker rune
}

func NewScatter(name string, pts ...Point) *Scatter {
	return &Scatter{name: name, pts: pts}
}

func NewScatterVals(name string, vals []float64) *Scatter {
	return &Scatter{name: name, pts: Seq(vals...)}
}

func (s *Scatter) Points(pts ...Point) *Scatter { s.pts = append(s.pts, pts...); return s }
func (s *Scatter) Color(c Color) *Scatter       { s.color = c; return s }
func (s *Scatter) Marker(m rune) *Scatter       { s.marker = m; return s }
func (s *Scatter) SetName(n string) *Scatter    { s.name = n; return s }

func (s *Scatter) seriesName() string { return s.name }
func (s *Scatter) bounds(db *dataBounds) {
	for _, p := range s.pts {
		db.add(p.X, p.Y)
	}
}
func (s *Scatter) hasColor() bool { return !s.color.IsZero() }
func (s *Scatter) setColor(c Color) {
	if s.color.IsZero() {
		s.color = c
	}
}
func (s *Scatter) colorOf() Color { return s.color }

func (s *Scatter) draw(cv *Canvas, fr frame, st Style) {
	area := fr.area
	m := s.marker
	if m == 0 {
		if fr.uni {
			m = '●'
		} else {
			m = 'o'
		}
	}
	for _, p := range s.pts {
		q := project(fr, p)
		if math.IsNaN(q.x) || math.IsNaN(q.y) {
			continue
		}
		x := int(math.Round(clampF(q.x, float64(area.X), float64(area.X2()))))
		y := int(math.Round(clampF(q.y, float64(area.Y), float64(area.Y2()))))
		cv.Set(x, y, m, st)
	}
}

func (s *Scatter) legendEntry(st Style, uni bool) LegendEntry {
	glyph := "●●"
	if !uni {
		glyph = "oo"
	}
	return LegendEntry{Label: s.name, Style: st, Glyph: glyph}
}

type Plot struct {
	order []seriesI
	chartBase
	xKind Kind
	yKind Kind
}

func NewPlot() *Plot {
	return &Plot{chartBase: newChartBase()}
}

// Title sets the plot title.
func (p *Plot) Title(t string) *Plot { p.SetTitle(t); return p }

// XLabel sets the x axis label.
func (p *Plot) XLabel(l string) *Plot { p.SetXLabel(l); return p }

// YLabel sets the y axis label.
func (p *Plot) YLabel(l string) *Plot { p.SetYLabel(l); return p }

// Grid toggles grid lines.
func (p *Plot) Grid(on bool) *Plot { p.SetGrid(on); return p }

// Legend toggles the series legend.
func (p *Plot) Legend(on bool) *Plot { p.SetLegend(on); return p }

// Height pins the plot height in rows.
func (p *Plot) Height(rows int) *Plot { p.SetSize(rows); return p }

// Orientation swaps the axes when given OrientHorizontal; OrientVertical
// or OrientAuto restore the standard x→columns layout.
func (p *Plot) Orientation(o Orientation) *Plot { p.SetOrientation(o); return p }

// Add appends one or more *Line or *Scatter series to the plot.
func (p *Plot) Add(ss ...any) *Plot {
	for _, s := range ss {
		switch v := s.(type) {
		case *Line:
			p.order = append(p.order, v)
		case *Scatter:
			p.order = append(p.order, v)
		}
	}
	return p
}

func (p *Plot) LogX(on bool) *Plot {
	if on {
		p.xKind = Logarithmic
	} else {
		p.xKind = Linear
	}
	return p
}

func (p *Plot) LogY(on bool) *Plot {
	if on {
		p.yKind = Logarithmic
	} else {
		p.yKind = Linear
	}
	return p
}

func (p *Plot) HeightHint(width int) int {
	if h := p.chartBase.HeightHint(width); h > 0 {
		return h
	}
	h := width * 2 / 5
	if h > 24 {
		h = 24
	}
	if h < 9 {
		h = 9
	}
	return h
}

func (p *Plot) Draw(rc *Ctx, cv *Canvas) {
	base := &p.chartBase
	xk, yk := p.xKind, p.yKind
	swap := base.orient == OrientHorizontal
	if swap {
		// Present the transposed view: series points are swapped and all
		// per-axis configuration follows them.
		cp := *base
		cp.xLabel, cp.yLabel = base.yLabel, base.xLabel
		cp.xTicks, cp.yTicks = base.yTicks, base.xTicks
		cp.xFmt, cp.yFmt = base.yFmt, base.xFmt
		cp.x0, cp.x1 = base.y0, base.y1
		cp.y0, cp.y1 = base.x0, base.x1
		cp.xSet, cp.ySet = base.ySet, base.xSet
		xk, yk = yk, xk
		base = &cp
	}

	var db dataBounds
	db.empty = true
	for _, s := range p.order {
		maybeSwap(s, swap).bounds(&db)
	}
	ci := 0
	for _, s := range p.order {
		if !s.hasColor() {
			s.setColor(rc.Palette[ci%len(rc.Palette)])
		}
		if s.hasColor() {
			ci++
		}
	}
	fr := prepareFrame(cv, rc, base, db, xk, yk, true)
	for _, s := range p.order {
		st := S(s.colorOf())
		maybeSwap(s, swap).draw(cv, fr, st)
	}
	entries := make([]LegendEntry, 0, len(p.order))
	for _, s := range p.order {
		entries = append(entries, s.legendEntry(S(s.colorOf()), rc.Info.Unicode))
	}
	drawLegendInside(cv, fr.area, entries, rc.Info.Unicode)

	if db.empty && p.title == "" && len(p.order) == 0 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
	}
}

// maybeSwap returns a shallow copy of the series with transposed points
// when swap is set; Line and Scatter are supported, anything else passes
// through untouched.
func maybeSwap(s seriesI, swap bool) seriesI {
	if !swap {
		return s
	}
	switch v := s.(type) {
	case *Line:
		c := *v
		c.pts = transposePts(v.pts)
		return &c
	case *Scatter:
		c := *v
		c.pts = transposePts(v.pts)
		return &c
	}
	return s
}

func transposePts(pts []Point) []Point {
	out := make([]Point, len(pts))
	for i, p := range pts {
		out[i] = Point{X: p.Y, Y: p.X}
	}
	return out
}
