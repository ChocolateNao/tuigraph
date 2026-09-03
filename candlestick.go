package tuichart

import (
	"math"
	"sort"
	"time"
)

// Candle holds the four OHLC prices of one time frame.
type Candle struct {
	At    time.Time
	Open  float64
	High  float64
	Low   float64
	Close float64
}

// Candlestick renders financial OHLC data: each candle is one time frame
// with a body spanning open..close — colored up (default green) when the
// close is >= the open, down (default red) otherwise — and thin wicks
// reaching to the period high and low. Colors are configurable via
// UpColor/DownColor; X ticks use the shared time-formatting rules.
type Candlestick struct {
	layout  string
	candles []Candle
	chartBase
	up   Color
	down Color
}

func NewCandlestick() *Candlestick {
	return &Candlestick{
		chartBase: newChartBase(),
		up:        Indexed(2), // green
		down:      Indexed(1), // red
	}
}

// Title sets the chart title.
func (c *Candlestick) Title(s string) *Candlestick { c.SetTitle(s); return c }

// Candle appends one OHLC observation at time at.
func (c *Candlestick) Candle(at time.Time, open, high, low, close float64) *Candlestick {
	c.candles = append(c.candles, Candle{At: at, Open: open, High: high, Low: low, Close: close})
	return c
}

// Candles appends several Candle structs at once.
func (c *Candlestick) Candles(cs ...Candle) *Candlestick {
	c.candles = append(c.candles, cs...)
	return c
}

// UpColor sets the bullish (close >= open) color.
func (c *Candlestick) UpColor(col Color) *Candlestick { c.up = col; return c }

// DownColor sets the bearish (close < open) color.
func (c *Candlestick) DownColor(col Color) *Candlestick { c.down = col; return c }

// ResetColors restores the default green/red pair.
func (c *Candlestick) ResetColors() {
	c.up, c.down = Indexed(2), Indexed(1)
}

// Format sets the Go time layout for X tick labels; empty restores auto.
func (c *Candlestick) Format(layout string) *Candlestick { c.layout = layout; return c }

func (c *Candlestick) HeightHint(width int) int {
	if h := c.chartBase.HeightHint(width); h > 0 {
		return h
	}
	p := NewPlot()
	p.chartBase = c.chartBase
	return p.HeightHint(width)
}

func (c *Candlestick) Draw(rc *Ctx, cv *Canvas) {
	uni := rc.Info.Unicode
	mono := rc.Info.Level == LevelNone
	inner := c.frameTitle(cv, uni)

	if len(c.candles) == 0 || inner.W < 8 || inner.H < 4 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}

	cs := make([]Candle, len(c.candles))
	copy(cs, c.candles)
	sort.Slice(cs, func(i, j int) bool { return cs[i].At.Before(cs[j].At) })

	hi := math.Inf(-1)
	lo := math.Inf(1)
	for _, k := range cs {
		if math.IsNaN(k.Low) || math.IsNaN(k.High) || k.High < k.Low {
			continue
		}
		lo = math.Min(lo, k.Low)
		hi = math.Max(hi, k.High)
	}
	if lo > hi { // no valid candle
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}
	padv := (hi - lo) * 0.05
	if padv <= 0 {
		padv = 1
	}

	fmtFn := c.xFmt
	if fmtFn == nil {
		layout := c.layout
		if layout == "" {
			layout = autoTimeLayout(float64(cs[len(cs)-1].At.Unix() - cs[0].At.Unix()))
		}
		fmtFn = func(v float64) string { return time.Unix(int64(v), 0).Format(layout) }
	}

	var db dataBounds
	db.empty = false
	db.y0, db.y1 = lo-padv, hi+padv
	db.x0 = float64(cs[0].At.Unix())
	db.x1 = float64(cs[len(cs)-1].At.Unix())
	if db.x1 == db.x0 {
		db.x1 = db.x0 + 1
	}
	savedFmt := c.xFmt
	c.xFmt = fmtFn
	fr := prepareFrame(cv, rc, &c.chartBase, db, Linear, Linear, true)
	c.xFmt = savedFmt

	wickCh, dashCh := '│', '─'
	bodyUp, bodyDown := '█', '█'
	if !uni {
		wickCh, dashCh = '|', '-'
	}
	if mono { // colorless: distinguish direction by glyph
		bodyUp, bodyDown = '#', '%'
	}

	n := len(cs)
	gaps := make([]int, 0, n-1)
	for i := 1; i < n; i++ {
		d := fr.mxCol(float64(cs[i].At.Unix())) - fr.mxCol(float64(cs[i-1].At.Unix()))
		if d > 0 {
			gaps = append(gaps, d)
		}
	}
	slot := inner.W / 3
	if len(gaps) > 0 {
		sort.Ints(gaps)
		slot = gaps[len(gaps)/2]
	}
	bw := slot * 7 / 10
	if bw < 1 {
		bw = 1
	}
	if bw > 9 {
		bw = 9
	}

	for _, k := range cs {
		if math.IsNaN(k.Low) || math.IsNaN(k.High) || k.High < k.Low {
			continue
		}
		bull := k.Close >= k.Open
		col := c.up
		bodyCh := bodyUp
		if !bull {
			col = c.down
			bodyCh = bodyDown
		}
		st := S(col)
		cx := fr.mxCol(float64(k.At.Unix()))

		yHi := fr.myRow(math.Max(k.High, math.Max(k.Open, k.Close)))
		yLo := fr.myRow(math.Min(k.Low, math.Min(k.Open, k.Close)))
		for y := yHi; y <= yLo; y++ {
			if y >= fr.area.Y && y <= fr.area.Y2() {
				cv.Set(cx, y, wickCh, st)
			}
		}

		bHiV := math.Max(k.Open, k.Close)
		bLoV := math.Min(k.Open, k.Close)
		yb0 := fr.myRow(bHiV)
		yb1 := fr.myRow(bLoV)
		x0 := cx - bw/2
		xs := maxInt(x0, fr.area.X)
		xe := minInt(x0+bw-1, fr.area.X2())
		if bHiV == bLoV || yb1-yb0+1 < 1 {
			// doji: flat line instead of an empty box
			if yb0 >= fr.area.Y && yb0 <= fr.area.Y2() && xs <= xe {
				cv.HLine(yb0, xs, xe, dashCh, st)
			}
			continue
		}
		for y := yb0; y <= yb1; y++ {
			if y < fr.area.Y || y > fr.area.Y2() {
				continue
			}
			for x := xs; x <= xe; x++ {
				cv.Set(x, y, bodyCh, st)
			}
		}
	}

	if c.legend {
		gu, gd := "██", "██"
		if mono {
			gu, gd = "#", "%"
		}
		drawLegendInside(cv, fr.area, []LegendEntry{
			{Label: "up", Style: S(c.up), Glyph: gu},
			{Label: "down", Style: S(c.down), Glyph: gd},
		}, uni)
	}
}
