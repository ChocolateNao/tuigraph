package tuichart

import "math"

var (
	barEighths  = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	barEighthsH = []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}
)

// BarSeries holds the data and color for one series in a bar chart.
type BarSeries struct {
	Name   string
	Values []float64
	Color  Color
}

// BarChart renders categorical data as vertical or horizontal bars.
type BarChart struct {
	cats   []string
	series []BarSeries
	chartBase
	stacked    bool
	horizontal bool
}

// NewBar creates a bar chart with the given categories and series.
func NewBar(cats []string, series ...BarSeries) *BarChart {
	return &BarChart{chartBase: newChartBase(), cats: cats, series: series}
}

// NewBarValues creates a bar chart with a single series from the given values.
func NewBarValues(cats []string, vals []float64) *BarChart {
	return NewBar(cats, BarSeries{Values: vals})
}

// Title sets the chart title.
func (b *BarChart) Title(t string) *BarChart { b.SetTitle(t); return b }

// Add appends a series to the chart.
func (b *BarChart) Add(s BarSeries) *BarChart { b.series = append(b.series, s); return b }

// Stacked stacks multiple series on top of each other instead of side by side.
func (b *BarChart) Stacked(on bool) *BarChart { b.stacked = on; return b }

// Horizontal renders categories down the left edge with bars growing right.
func (b *BarChart) Horizontal(on bool) *BarChart { b.horizontal = on; return b }

// ShowValues prints each bar's numeric value above its top (vertical bars).
func (b *BarChart) ShowValues(on bool) *BarChart { b.SetShowValues(on); return b }

// Orientation switches bar direction: OrientHorizontal renders categories
// down the left edge with bars growing right; OrientVertical forces the
// classic column layout; OrientAuto defers to Horizontal().
func (b *BarChart) Orientation(o Orientation) *BarChart { b.SetOrientation(o); return b }

func (b *BarChart) isHorizontal() bool {
	switch b.orient {
	case OrientHorizontal:
		return true
	case OrientVertical:
		return false
	}
	return b.horizontal
}

// HeightHint returns the suggested height for the given width.
func (b *BarChart) HeightHint(width int) int {
	if h := b.chartBase.HeightHint(width); h > 0 {
		return h
	}
	if b.isHorizontal() {
		// Height follows the number of category rows, not the width;
		// otherwise flipped charts waste most of their frame.
		return horizontalBarHeight(len(b.cats))
	}
	h := width / 3
	if h > 20 {
		h = 20
	}
	if h < 6 {
		h = 6
	}
	return h
}

// horizontalBarHeight sizes a left-to-right bar chart: one row per
// category plus border/margin allowance, bounded like the other diagrams.
func horizontalBarHeight(cats int) int {
	h := cats + 3
	if h < 6 {
		h = 6
	}
	if h > 30 {
		h = 30
	}
	return h
}

// Draw renders the bar chart onto the canvas.
func (b *BarChart) Draw(rc *Ctx, cv *Canvas) {
	if b.isHorizontal() {
		b.drawHorizontal(rc, cv)
		return
	}
	b.drawVertical(rc, cv)
}

func (b *BarChart) drawVertical(rc *Ctx, cv *Canvas) {
	var db dataBounds
	db.empty = true
	if b.stacked {
		nC := len(b.cats)
		pos := make([]float64, nC)
		neg := make([]float64, nC)
		for _, s := range b.series {
			for c, v := range s.Values {
				if c >= nC || math.IsNaN(v) {
					continue
				}
				if v > 0 {
					pos[c] += v
				} else {
					neg[c] += v
				}
			}
		}
		for c := 0; c < nC; c++ {
			db.add(0, pos[c])
			db.add(0, neg[c])
		}
	} else {
		for _, s := range b.series {
			for _, v := range s.Values {
				db.add(0, v)
				db.add(0, 0)
			}
		}
	}
	db.x0, db.x1 = 0, 1
	fr := prepareFrame(cv, rc, &b.chartBase, db, Linear, Linear, false)
	area := fr.area
	if area.W < 2 || area.H < 2 || len(b.cats) == 0 || len(b.series) == 0 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}
	nC := len(b.cats)
	slot := area.W / nC
	if slot < 1 {
		slot = 1
	}

	ci := 0
	for i := range b.series {
		if b.series[i].Color.IsZero() {
			b.series[i].Color = rc.Palette[ci%len(rc.Palette)]
		} else {
			ci++
		}
	}

	posOff := make([]float64, nC)
	negOff := make([]float64, nC)
	type valLabel struct {
		text string
		x    int
		y    int
	}
	var labels []valLabel
	recordLabel := func(bx, bw, topY int, v float64) {
		if !b.showVals || math.IsNaN(v) || v == 0 {
			return
		}
		if topY-1 < area.Y || topY > area.Y2() {
			return
		}
		labels = append(labels, valLabel{
			x:    bx + bw/2,
			y:    topY - 1,
			text: FormatValue(v),
		})
	}

	for si := range b.series {
		s := &b.series[si]
		st := S(s.Color)
		for c := 0; c < nC && c < len(s.Values); c++ {
			v := s.Values[c]
			if math.IsNaN(v) || v == 0 {
				continue
			}
			var bw, bx int
			if b.stacked || len(b.series) == 1 {
				bw = slot - slot/4
				if bw < 1 {
					bw = 1
				}
				bx = area.X + c*slot + (slot-bw)/2
				from := 0.0
				if b.stacked {
					if v > 0 {
						from, posOff[c] = posOff[c], posOff[c]+v
					} else {
						from, negOff[c] = negOff[c], negOff[c]+v
					}
				}
				paintColumn(cv, fr, bx, bw, from, from+v, st)
				if !b.stacked && v > 0 {
					recordLabel(bx, bw, fr.myRow(from+v), from+v)
				}
			} else {
				gw := slot * 3 / 4
				bw = gw / len(b.series)
				if bw < 1 {
					bw = 1
				}
				bx = area.X + c*slot + (slot-gw)/2 + si*bw
				paintColumn(cv, fr, bx, bw, 0, v, st)
				if v > 0 {
					recordLabel(bx, bw, fr.myRow(v), v)
				}
			}
		}
	}

	if b.stacked && b.showVals {
		for c := 0; c < nC; c++ {
			if tot := posOff[c]; tot > 0 {
				bw := slot - slot/4
				if bw < 1 {
					bw = 1
				}
				recordLabel(area.X+c*slot+(slot-bw)/2, bw, fr.myRow(tot), tot)
			}
		}
	}

	valSt := S(Gray)
	for _, l := range labels {
		cv.TextCenter(l.x, l.y, l.text, valSt)
	}

	labelSt := S(Default)
	lastEnd := -1 << 30
	maxCatLen := 0
	for _, cat := range b.cats {
		if n := runeLen(cat); n > maxCatLen {
			maxCatLen = n
		}
	}
	stride := 1
	if maxCatLen+1 > slot && slot > 0 {
		stride = (maxCatLen + 2) / slot
		if stride < 1 {
			stride = 1
		}
	}
	axisRow := area.Y + area.H
	mark := '┬'
	if !fr.uni {
		mark = '+'
	}
	for c := 0; c < nC; c++ {
		col := area.X + c*slot + slot/2
		cv.Set(col, axisRow, mark, S(Silver))
		if stride > 1 && c%stride != 0 {
			continue
		}
		lbl := ellipTrunc(b.cats[c], slot, fr.uni)
		start := col - runeLen(lbl)/2
		if start < area.X || start <= lastEnd {
			continue
		}
		cv.Text(start, axisRow+1, lbl, labelSt)
		lastEnd = start + runeLen(lbl) - 1
	}

	var entries []LegendEntry
	for _, s := range b.series {
		glyph := "██"
		if !fr.uni {
			glyph = "##"
		}
		entries = append(entries, LegendEntry{Label: s.Name, Style: S(s.Color), Glyph: glyph})
	}
	drawLegendInside(cv, area, entries, fr.uni)
}

// paintColumn fills cells for a bar spanning values [from, to].
// Each touched row is filled proportionally to how much of it the bar covers,
// producing smooth eighth-block tops. Baseline-rooted edges (value 0) claim
// the whole zero row so bars sit ON the axis — except bars shorter than one
// cell row, which render as a single proportional segment (▁..▇) on the
// axis row instead of an overstated full block.
func paintColumn(cv *Canvas, fr frame, bx, bw int, from, to float64, st Style) {
	if bw < 1 || to == from {
		return
	}
	vLow, vHigh := math.Min(from, to), math.Max(from, to)
	y0 := fr.myRow(vHigh)
	y1 := fr.myRow(vLow)
	fillFull := '█'
	if !fr.uni {
		fillFull = '#'
	}
	if vLow == 0 {
		rowH := math.Abs(valueAtRow(fr, y1) - valueAtRow(fr, y1+1))
		if vHigh-vLow <= rowH {
			// Bar fits inside a single cell: one partial segment.
			idx := int((vHigh - vLow) / rowH * 8)
			if idx < 1 {
				idx = 1
			}
			if idx > 8 {
				idx = 8
			}
			ch := fillFull
			if fr.uni && idx < 8 {
				ch = barEighths[idx-1]
			}
			cv.HLine(y1, bx, bx+bw-1, ch, st)
			return
		}
		// Occupy the zero row down to its bottom edge (flush axis base).
		vLow = math.Min(vLow, valueAtRow(fr, y1+1))
	}
	if vHigh == 0 {
		rowH := math.Abs(valueAtRow(fr, y0) - valueAtRow(fr, y0+1))
		if vHigh-vLow <= rowH {
			idx := int((vHigh - vLow) / rowH * 8)
			if idx < 1 {
				idx = 1
			}
			if idx > 8 {
				idx = 8
			}
			ch := fillFull
			if fr.uni && idx < 8 {
				ch = barEighths[idx-1]
			}
			cv.HLine(y0, bx, bx+bw-1, ch, st)
			return
		}
		vHigh = math.Max(vHigh, valueAtRow(fr, y0))
	}
	for y := y0; y <= y1; y++ {
		if y < fr.area.Y || y > fr.area.Y2() {
			continue
		}
		rowTopV := valueAtRow(fr, y)
		rowBotV := valueAtRow(fr, y+1)
		height := rowTopV - rowBotV
		if height <= 0 {
			cv.HLine(y, bx, bx+bw-1, fillFull, st)
			continue
		}
		frac := (math.Min(vHigh, rowTopV) - math.Max(vLow, rowBotV)) / height
		if frac <= 0 {
			continue
		}
		ch := fillFull
		if fr.uni && frac < 0.995 {
			idx := int(frac * 8)
			if idx < 1 {
				idx = 1
			}
			if idx > 8 {
				idx = 8
			}
			ch = barEighths[idx-1]
		}
		cv.HLine(y, bx, bx+bw-1, ch, st)
	}
}

func valueAtRow(fr frame, y int) float64 {
	if fr.area.H <= 1 {
		return fr.ysc.Min
	}
	t := float64(fr.area.Y+fr.area.H-1-y) / float64(fr.area.H-1)
	return fr.ysc.Min + t*(fr.ysc.Max-fr.ysc.Min)
}

func (b *BarChart) drawHorizontal(rc *Ctx, cv *Canvas) {
	inner := b.frameTitle(cv, rc.Info.Unicode)
	uni := rc.Info.Unicode
	if len(b.cats) == 0 || len(b.series) == 0 || inner.W < 10 || inner.H < 2 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}
	s := b.series[0]
	if s.Color.IsZero() {
		s.Color = rc.Palette[0]
	}
	st := S(s.Color)
	hi := math.SmallestNonzeroFloat64
	for _, v := range s.Values {
		if !math.IsNaN(v) {
			hi = math.Max(hi, v)
		}
	}
	if hi <= 0 {
		hi = 1
	}
	gutter := 1
	for _, cat := range b.cats {
		if n := runeLen(cat) + 1; n > gutter && n*2 < inner.W {
			gutter = n
		}
	}
	valPad := len(FormatValue(hi)) + 1
	barW := inner.W - gutter - valPad - 1
	if barW < 2 {
		barW = 2
	}
	rows := inner.H - 1
	stride := 1
	if len(b.cats) > rows {
		stride = (len(b.cats) + rows - 1) / rows
	}
	ri := 0
	for c, cat := range b.cats {
		if c%stride != 0 || ri >= rows {
			break
		}
		y := inner.Y + ri
		ri++
		cv.TextRight(inner.X+gutter-2, y, ellipTrunc(cat, gutter-1, uni), S(Default))
		v := 0.0
		if c < len(s.Values) && !math.IsNaN(s.Values[c]) {
			v = s.Values[c]
		}
		wf := (v / hi) * float64(barW)
		if wf < 0 {
			wf = 0
		}
		w := int(wf)
		if w > barW {
			w = barW
		}
		for i := 0; i < w; i++ {
			ch := '#'
			if uni {
				ch = '█'
			}
			cv.Set(inner.X+gutter+i, y, ch, st)
		}
		if uni && w < barW {
			idx := int((wf - float64(w)) * 8)
			if idx >= 1 && idx <= 8 {
				cv.Set(inner.X+gutter+w, y, barEighthsH[idx-1], st)
				w++
			}
		}
		cv.Text(inner.X+gutter+w+1, y, FormatValue(v), S(Gray))
	}
}
