package tuichart

import "math"

type FunctionPlot struct {
	fn   func(float64) float64
	name string
	chartBase
	lo        float64
	hi        float64
	samples   int
	color     Color
	domainSet bool
	yKind     Kind
}

func NewFunction(f func(float64) float64) *FunctionPlot {
	return &FunctionPlot{
		chartBase: newChartBase(),
		fn:        f,
		lo:        -10,
		hi:        10,
		name:      "f(x)",
	}
}

// Title sets the chart title.
func (p *FunctionPlot) Title(t string) *FunctionPlot { p.SetTitle(t); return p }

// Domain sets the plotted x range.
func (p *FunctionPlot) Domain(lo, hi float64) *FunctionPlot {
	p.lo, p.hi, p.domainSet = lo, hi, true
	return p
}

func (p *FunctionPlot) ResetDomain() *FunctionPlot {
	p.lo, p.hi, p.domainSet = -10, 10, false
	return p
}

func (p *FunctionPlot) Samples(n int) *FunctionPlot {
	if n > 8 {
		p.samples = n
	}
	return p
}

func (p *FunctionPlot) Name(s string) *FunctionPlot { p.name = s; return p }
func (p *FunctionPlot) Color(c Color) *FunctionPlot { p.color = c; return p }

// LogY switches the y axis to a logarithmic scale.
func (p *FunctionPlot) LogY(on bool) *FunctionPlot {
	if on {
		p.yKind = Logarithmic
	} else {
		p.yKind = Linear
	}
	return p
}

func (p *FunctionPlot) sample(width int) [][]Point {
	n := p.samples
	if n == 0 {
		n = maxInt(width*4, 100)
	}
	var segs [][]Point
	cur := make([]Point, 0, n)
	for i := 0; i <= n; i++ {
		x := p.lo + (p.hi-p.lo)*float64(i)/float64(n)
		y := p.fn(x)
		if math.IsNaN(y) || math.IsInf(y, 0) {
			if len(cur) > 0 {
				segs = append(segs, cur)
				cur = make([]Point, 0, n)
			}
			continue
		}
		cur = append(cur, Point{X: x, Y: y})
	}
	if len(cur) > 0 {
		segs = append(segs, cur)
	}
	return segs
}

func (p *FunctionPlot) HeightHint(width int) int {
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

func (p *FunctionPlot) Draw(rc *Ctx, cv *Canvas) {
	segs := p.sample(cv.Width())
	color := p.color
	if color.IsZero() {
		color = rc.Palette[0]
	}
	line := &Line{name: p.name}
	line.color = color
	var db dataBounds
	db.empty = true
	db.add(p.lo, 0)
	db.add(p.hi, 0)
	for _, seg := range segs {
		for _, pt := range seg {
			db.add(pt.X, pt.Y)
			line.pts = append(line.pts, pt)
		}
		for i := 0; i < len(seg)-1; i++ {
			midY := (seg[i].Y + seg[i+1].Y) / 2
			midX := (seg[i].X + seg[i+1].X) / 2
			fy := p.fn(midX)
			if !math.IsNaN(fy) && math.Abs(fy-midY) > 5*math.Abs(seg[i].Y)+5 {
				line.pts = append(line.pts, Point{X: midX, Y: NaN()})
			}
		}
	}
	fr := prepareFrame(cv, rc, &p.chartBase, db, Linear, p.yKind, true)
	line.draw(cv, fr, S(color))
	drawLegendInside(cv, fr.area, []LegendEntry{{
		Label: p.name,
		Style: S(color),
		Glyph: tern(fr.uni, "───", "---"),
	}}, fr.uni)
}

func NaN() float64 { return math.NaN() }

func tern(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
