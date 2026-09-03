package tuichart

const (
	defaultTickTarget = 5
	defaultPad        = 0.06
)

// Align controls where a title is placed along its row.
type Align uint8

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// Orientation selects which axis carries values.
type Orientation uint8

const (
	// OrientAuto keeps each diagram's default (bars grow upward,
	// plots map x→columns).
	OrientAuto Orientation = iota
	// OrientHorizontal puts values along X: bars grow rightward, plots
	// swap axes so y=f(x) becomes x=f(y).
	OrientHorizontal
	// OrientVertical puts values along Y: bars grow upward, plots keep
	// the standard orientation even if legacy flags said otherwise.
	OrientVertical
)

// chartBase holds configuration shared by all built-in diagrams, including
// the Set*/Reset* API for size, scale and axis marks.
type chartBase struct {
	yFmt       func(float64) string
	xFmt       func(float64) string
	title      string
	xLabel     string
	yLabel     string
	yTicks     []Tick
	xTicks     []Tick
	cellW      int
	tickN      int
	y1         float64
	height     int
	y0         float64
	x1         float64
	x0         float64
	grid       bool
	titleAlign Align
	xSet       bool
	ySet       bool
	frame      bool
	legend     bool
	orient     Orientation
	showVals   bool
}

func newChartBase() chartBase {
	return chartBase{grid: true, frame: true, legend: true, tickN: defaultTickTarget}
}

func (b *chartBase) SetTitle(t string) { b.title = t }
func (b *chartBase) ResetTitle()       { b.title = "" }

// SetShowValues toggles numeric value annotations (bar tops, heatmap
// cells, pie legend values).
func (b *chartBase) SetShowValues(v bool) { b.showVals = v }
func (b *chartBase) ResetShowValues()     { b.showVals = false }

// SetCellWidth fixes the heatmap cell width in columns; 0 restores the
// automatic stretch that fills the frame width.
func (b *chartBase) SetCellWidth(n int)    { b.cellW = n }
func (b *chartBase) ResetCellWidth()       { b.cellW = 0 }
func (b *chartBase) SetTitleAlign(a Align) { b.titleAlign = a }

// ResetTitleAlign restores left alignment.
func (b *chartBase) ResetTitleAlign() { b.titleAlign = AlignLeft }

// SetOrientation switches the value axis; see Orientation for semantics.
// Diagrams that cannot swap axes ignore it.
func (b *chartBase) SetOrientation(o Orientation) { b.orient = o }

// ResetOrientation restores each diagram's default axis layout.
func (b *chartBase) ResetOrientation()  { b.orient = OrientAuto }
func (b *chartBase) SetXLabel(l string) { b.xLabel = l }
func (b *chartBase) SetYLabel(l string) { b.yLabel = l }
func (b *chartBase) ResetLabels()       { b.xLabel, b.yLabel = "", "" }

func (b *chartBase) SetXTicks(t []Tick) { b.xTicks = t }
func (b *chartBase) SetYTicks(t []Tick) { b.yTicks = t }
func (b *chartBase) ResetTicks()        { b.xTicks, b.yTicks = nil, nil }

func (b *chartBase) SetXFormatter(f func(float64) string) { b.xFmt = f }
func (b *chartBase) SetYFormatter(f func(float64) string) { b.yFmt = f }
func (b *chartBase) ResetFormatters()                     { b.xFmt, b.yFmt = nil, nil }

// SetScale fixes both axis ranges, disabling auto-scaling.
func (b *chartBase) SetScale(x0, x1, y0, y1 float64) {
	b.x0, b.x1, b.y0, b.y1 = x0, x1, y0, y1
	b.xSet, b.ySet = true, true
}

func (b *chartBase) SetXRange(lo, hi float64) { b.x0, b.x1, b.xSet = lo, hi, true }
func (b *chartBase) SetYRange(lo, hi float64) { b.y0, b.y1, b.ySet = lo, hi, true }

// ResetScale re-enables automatic scaling from data.
func (b *chartBase) ResetScale() {
	b.xSet, b.ySet = false, false
	b.xTicks, b.yTicks = nil, nil
}

// SetSize pins this diagram's height in rows when rendered in a Chart.
func (b *chartBase) SetSize(rows int) { b.height = rows }
func (b *chartBase) ResetSize()       { b.height = 0 }
func (b *chartBase) HeightHint(int) int {
	if b.height > 0 {
		return b.height
	}
	return 0
}

func (b *chartBase) SetGrid(on bool)   { b.grid = on }
func (b *chartBase) SetBorder(on bool) { b.frame = on }
func (b *chartBase) SetLegend(on bool) { b.legend = on }
func (b *chartBase) SetTickCount(n int) {
	if n < 2 {
		n = 2
	}
	b.tickN = n
}

// Reset restores every configurable property to its default.
func (b *chartBase) Reset() {
	title := b.title
	*b = newChartBase()
	b.title = title
}

func (b *chartBase) frameTitle(cv *Canvas, uni bool) Rect {
	r := cv.Rect()
	if !b.frame {
		if b.title != "" && r.H > 1 {
			drawAlignedText(
				cv,
				0,
				ellipTrunc(b.title, r.W, uni),
				S(Default).Bolder(),
				b.titleAlign,
				r.W,
			)
			return Rect{X: 0, Y: 1, W: r.W, H: r.H - 1}.clip(r)
		}
		return r
	}
	st := S(Gray)
	cv.Border(st, uni)
	if b.title != "" && r.H > 2 {
		t := " " + b.title + " "
		if runeLen(t)+4 > r.W {
			t = " " + ellipTrunc(b.title, maxInt(r.W-6, 1), uni) + " "
		}
		drawAlignedText(cv, 0, t, S(Default).Bolder(), b.titleAlign, r.W)
		return Rect{X: 1, Y: 1, W: r.W - 2, H: r.H - 2}.clip(r)
	}
	return Rect{X: 1, Y: 1, W: r.W - 2, H: r.H - 2}.clip(r)
}

// drawAlignedText places s on row y honoring a within width w. Left-aligned
// text keeps a small margin so it clears frame borders.
func drawAlignedText(cv *Canvas, y int, s string, st Style, a Align, w int) {
	n := runeLen(s)
	switch a {
	case AlignRight:
		x := w - n - 1
		if x < 0 {
			x = 0
		}
		cv.Text(x, y, s, st)
	case AlignCenter:
		cv.TextCenter(w/2, y, s, st)
	default:
		x := 0
		if w > n+3 {
			x = 2
		}
		cv.Text(x, y, s, st)
	}
}

func (r Rect) clip(o Rect) Rect {
	if r.X < o.X {
		r.W -= o.X - r.X
		r.X = o.X
	}
	if r.Y < o.Y {
		r.H -= o.Y - r.Y
		r.Y = o.Y
	}
	if r.W > o.W-(r.X-o.X) {
		r.W = o.W - (r.X - o.X)
	}
	if r.H > o.H-(r.Y-o.Y) {
		r.H = o.H - (r.Y - o.Y)
	}
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}
