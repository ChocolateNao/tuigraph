package tuichart

import "math"

type pieSlice struct {
	name  string
	val   float64
	color Color
}

type PieChart struct {
	slices []pieSlice
	chartBase
	donut bool
}

func NewPie() *PieChart {
	return &PieChart{chartBase: newChartBase(), donut: false}
}

// Slice appends a slice; the color is assigned from the palette when omitted.
func (p *PieChart) Slice(name string, val float64) *PieChart {
	p.slices = append(p.slices, pieSlice{name: name, val: val})
	return p
}

func (p *PieChart) SliceColor(c Color) *PieChart {
	if len(p.slices) > 0 {
		p.slices[len(p.slices)-1].color = c
	}
	return p
}

func (p *PieChart) Donut(on bool) *PieChart { p.donut = on; return p }

// ShowValues includes each slice's raw value in the legend next to its
// percentage (terminal pie legends are the practical spot for numbers).
func (p *PieChart) ShowValues(v bool) *PieChart { p.SetShowValues(v); return p }

// Title sets the chart title.
func (p *PieChart) Title(t string) *PieChart { p.SetTitle(t); return p }

func (p *PieChart) HeightHint(width int) int {
	if h := p.chartBase.HeightHint(width); h > 0 {
		return h
	}
	h := width * 3 / 5
	if h > 22 {
		h = 22
	}
	if h < 8 {
		h = 8
	}
	return h
}

var pieASCIIChars = []rune{'#', '@', '*', 'o', '=', '+', '~', '%'}

func (p *PieChart) Draw(rc *Ctx, cv *Canvas) {
	total := 0.0
	ci := 0
	for i := range p.slices {
		if p.slices[i].color.IsZero() {
			p.slices[i].color = rc.Palette[ci%len(rc.Palette)]
			ci++
		}
		total += math.Max(p.slices[i].val, 0)
	}
	inner := p.frameTitle(cv, rc.Info.Unicode)

	if total <= 0 || inner.W < 6 || inner.H < 4 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}

	reserve := (inner.W*2)/5 + 1
	if reserve > len(p.slices)*14+6 {
		reserve = len(p.slices)*14 + 6
	}
	cx := inner.X + (inner.W-reserve)/2
	cy := inner.Y + inner.H/2
	rx := (inner.W - reserve - 1) / 2
	ry := (inner.H - 1) / 2
	if rx < 1 || ry < 1 {
		rx = maxInt(rx, 1)
		ry = maxInt(ry, 1)
	}

	var entries []LegendEntry
	acc := 0.0
	for i := range p.slices {
		s := &p.slices[i]
		v := math.Max(s.val, 0)
		from := acc / total
		acc += v
		to := acc / total
		st := S(s.color)
		mono := rc.Info.Level == LevelNone
		fillCh := '█'
		if mono || !rc.Info.Unicode {
			fillCh = pieASCIIChars[i%len(pieASCIIChars)]
		}
		drawPieSector(cv, cx, cy, rx, ry, from, to, fillCh, st, p.donut)
		pct := v / total * 100
		label := s.name + " " + trimPct(pct) + "%"
		if p.showVals {
			label = s.name + " " + FormatValue(v) + " (" + trimPct(pct) + "%)"
		}
		glyph := "██"
		if mono || !rc.Info.Unicode {
			glyph = string(fillCh) + string(fillCh)
		}
		entries = append(entries, LegendEntry{
			Label: label,
			Style: st,
			Glyph: glyph,
		})
	}
	lx := minInt(cx+rx+2, inner.X2()-12)
	drawLegendColumn(
		cv,
		Rect{X: lx, Y: cy - len(entries)/2, W: inner.X2() - lx + 1, H: len(entries)},
		entries,
		rc.Info.Unicode,
	)
}

func drawLegendColumn(cv *Canvas, r Rect, entries []LegendEntry, uni bool) {
	for i, e := range entries {
		y := r.Y + i
		if y > r.Y2() || y < 0 {
			continue
		}
		x := r.X
		n := cv.Text(x, y, e.Glyph, e.Style)
		cv.Text(x+n+1, y, ellipTrunc(e.Label, r.W-n-2, uni), e.Style)
	}
}

func trimPct(v float64) string {
	s := FormatValue(math.Round(v*10) / 10)
	return s
}

func drawPieSector(
	cv *Canvas,
	cx, cy, rx, ry int,
	from, to float64,
	ch rune,
	st Style,
	donut bool,
) {
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			nx := float64(dx) / float64(rx)
			ny := float64(dy) / float64(ry)
			d := nx*nx + ny*ny
			if d > 1 {
				continue
			}
			if donut && d < 0.30 {
				continue
			}
			ang := math.Atan2(nx, -ny)
			if ang < 0 {
				ang += 2 * math.Pi
			}
			t := ang / (2 * math.Pi)
			if t >= from-1e-9 && t < to+1e-9 {
				cv.Set(cx+dx, cy+dy, ch, st)
			}
		}
	}
}
