package tuichart

import "math"

type dataBounds struct {
	x0, y0, x1, y1 float64
	empty          bool
}

func (db *dataBounds) add(x, y float64) {
	if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		return
	}
	if db.empty {
		db.x0, db.x1, db.y0, db.y1 = x, x, y, y
		db.empty = false
		return
	}
	db.x0 = math.Min(db.x0, x)
	db.y0 = math.Min(db.y0, y)
	db.x1 = math.Max(db.x1, x)
	db.y1 = math.Max(db.y1, y)
}

type frame struct {
	xt   []Tick
	yt   []Tick
	area Rect
	xsc  Scale
	ysc  Scale
	uni  bool
}

func (fr frame) mx(v float64) float64 {
	if fr.xsc.Max == fr.xsc.Min {
		return float64(fr.area.X)
	}
	t := (v - fr.xsc.Min) / (fr.xsc.Max - fr.xsc.Min)
	return float64(fr.area.X) + t*float64(fr.area.W-1)
}

func (fr frame) my(v float64) float64 {
	if fr.ysc.Max == fr.ysc.Min {
		return float64(fr.area.Y)
	}
	t := (v - fr.ysc.Min) / (fr.ysc.Max - fr.ysc.Min)
	return float64(fr.area.Y+fr.area.H-1) - t*float64(fr.area.H-1)
}

func (fr frame) myRow(v float64) int { return int(math.Round(fr.my(v))) }
func (fr frame) mxCol(v float64) int { return int(math.Round(fr.mx(v))) }

func resolveScales(b *chartBase, db dataBounds, xk, yk Kind) (Scale, Scale) {
	var xs, ys Scale
	if b.xSet {
		xs = fixedScale(xk, b.x0, b.x1)
	} else if db.empty {
		xs = fixedScale(xk, 0, 1)
	} else {
		xs = newScale(xk, db.x0, db.x1, defaultPad)
	}
	if b.ySet {
		ys = fixedScale(yk, b.y0, b.y1)
	} else if db.empty {
		ys = fixedScale(yk, 0, 1)
	} else {
		ys = newScale(yk, db.y0, db.y1, defaultPad)
	}
	return xs, ys
}

func fmtTicks(t []Tick, f func(float64) string) []Tick {
	if f == nil {
		return t
	}
	out := make([]Tick, len(t))
	for i, tk := range t {
		out[i] = Tick{Value: tk.Value, Label: f(tk.Value)}
	}
	return out
}

func prepareFrame(cv *Canvas, rc *Ctx, b *chartBase, db dataBounds, xk, yk Kind, autoX bool) frame {
	uni := rc.Info.Unicode
	inner := b.frameTitle(cv, uni)

	xsc, ysc := resolveScales(b, db, xk, yk)
	yt := b.yTicks
	if yt == nil {
		yt = ysc.Ticks(b.tickN)
	}
	yt = fmtTicks(yt, b.yFmt)
	xt := b.xTicks
	if xt == nil && autoX {
		xt = xsc.Ticks(b.tickN)
	}
	xt = fmtTicks(xt, b.xFmt)

	gutter := 1
	for _, t := range yt {
		if n := runeLen(t.Label) + 2; n > gutter {
			gutter = n
		}
	}
	bottom := 2
	top := 0
	if b.title != "" {
		top = 1
	}
	if b.xLabel != "" {
		bottom++
	}
	if inner.H <= top+bottom+2 || inner.W <= gutter+4 {
		return frame{area: inner.clip(cv.Rect()), xsc: xsc, ysc: ysc, xt: xt, yt: yt, uni: uni}
	}

	plot := Rect{
		X: inner.X + gutter,
		Y: inner.Y + top,
		W: inner.W - gutter - 1,
		H: inner.H - top - bottom,
	}

	dim := S(DimGray)
	gridCh := '·'
	if !uni {
		gridCh = '.'
	}
	if b.grid {
		for _, t := range yt {
			row := fr2myRow(plot, ysc, t.Value)
			cv.HLine(row, plot.X, plot.X2(), gridCh, dim)
		}
		for _, t := range xt {
			col := fr2mxCol(plot, xsc, t.Value)
			cv.VLine(col, plot.Y, plot.Y2(), gridCh, dim)
		}
	}

	axisSt := S(Silver)
	hz, vt := '─', '│'
	if !uni {
		hz, vt = '-', '|'
	}
	axisRow := plot.Y + plot.H
	cv.VLine(plot.X-1, plot.Y, axisRow-1, vt, axisSt)
	cv.HLine(axisRow, plot.X, plot.X2(), hz, axisSt)
	cv.Set(plot.X-1, axisRow, '└', axisSt)
	if !uni {
		cv.Set(plot.X-1, axisRow, '+', axisSt)
	}

	labelSt := S(Default)
	for _, t := range yt {
		row := fr2myRow(plot, ysc, t.Value)
		if row < plot.Y || row > axisRow-1 {
			continue
		}
		mark := '┤'
		if !uni {
			mark = '|'
		}
		cv.Set(plot.X-1, row, mark, axisSt)
		if t.Label != "" {
			cv.TextRight(plot.X-2, row, t.Label, labelSt)
		}
	}
	lastEnd := -1 << 30
	for _, t := range xt {
		col := fr2mxCol(plot, xsc, t.Value)
		if col < plot.X || col > plot.X2() {
			continue
		}
		mark := '┬'
		if !uni {
			mark = '+'
		}
		cv.Set(col, axisRow, mark, axisSt)
		if t.Label == "" {
			continue
		}
		w := runeLen(t.Label)
		start := col - w/2
		if start < plot.X-1 {
			start = plot.X - 1
		}
		if start+w-1 > plot.X2()+1 {
			start = plot.X2() + 1 - w
		}
		if start <= lastEnd {
			continue
		}
		if start >= plot.X-1 && start+w-1 <= plot.X2()+2 {
			cv.Text(start, axisRow+1, t.Label, labelSt)
			lastEnd = start + w
		}
	}
	if b.xLabel != "" {
		cv.TextCenter(plot.X+plot.W/2, axisRow+2, ellipTrunc(b.xLabel, plot.W, uni), S(Silver))
	}

	return frame{area: plot, xsc: xsc, ysc: ysc, xt: xt, yt: yt, uni: uni}
}

func fr2myRow(r Rect, s Scale, v float64) int {
	if s.Max == s.Min {
		return r.Y
	}
	t := (v - s.Min) / (s.Max - s.Min)
	return int(math.Round(float64(r.Y+r.H-1) - t*float64(r.H-1)))
}

func fr2mxCol(r Rect, s Scale, v float64) int {
	if s.Max == s.Min {
		return r.X
	}
	t := (v - s.Min) / (s.Max - s.Min)
	return int(math.Round(float64(r.X) + t*float64(r.W-1)))
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
