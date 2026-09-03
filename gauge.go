package tuichart

import (
	"fmt"
	"math"
)

// GaugeStyle selects the visual language of a Gauge bar.
type GaugeStyle uint8

const (
	// GaugeBlocks draws ████░░░ with eighth-block precision at the tip.
	GaugeBlocks GaugeStyle = iota
	// GaugeASCII is the plain-terminal variant: ####-----.
	GaugeASCII
	// GaugeBrackets wraps the bar in brackets: [██████░░░░].
	GaugeBrackets
	// GaugeArrow draws ━━━━╸─── with an arrowhead tip.
	GaugeArrow
	// GaugeSegments renders ten discrete segments: ▰▰▰▰▱▱▱▱▱▱.
	GaugeSegments
)

// Gauge shows a single value against its maximum as a one-row progress
// bar. Works in live dashboards (mutate Value between frames) and embeds
// like any other Drawable.
type Gauge struct {
	label string
	chartBase
	value   float64
	max     float64
	color   Color
	style   GaugeStyle
	showPct bool
}

func NewGauge(value, max float64) *Gauge {
	return &Gauge{
		chartBase: newChartBase(),
		value:     value,
		max:       max,
		showPct:   true,
	}
}

// Title sets the chart title.
func (g *Gauge) Title(s string) *Gauge { g.SetTitle(s); return g }

// Value sets the current value; it is clamped to [0, max].
func (g *Gauge) Value(v float64) *Gauge { g.value = v; return g }

// Max sets the scale maximum (values above are clamped to full).
func (g *Gauge) Max(m float64) *Gauge { g.max = m; return g }

// Style selects one of the GaugeStyle renderings.
func (g *Gauge) Style(s GaugeStyle) *Gauge { g.style = s; return g }

// Label appends a text suffix after the percent readout.
func (g *Gauge) Label(s string) *Gauge { g.label = s; return g }

// Color fixes the fill color; defaults to the first palette entry.
func (g *Gauge) Color(c Color) *Gauge { g.color = c; return g }

// ShowPercent toggles the numeric percentage suffix (default on).
func (g *Gauge) ShowPercent(on bool) *Gauge { g.showPct = on; return g }

func (g *Gauge) HeightHint(width int) int {
	if h := g.chartBase.HeightHint(width); h > 0 {
		return h
	}
	if g.title != "" {
		return 3
	}
	return 1
}

func (g *Gauge) Draw(rc *Ctx, cv *Canvas) {
	inner := g.frameTitle(cv, rc.Info.Unicode)
	if inner.W < 3 || inner.H < 1 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}
	row := inner.Y + inner.H/2

	st := S(g.color)
	if g.color.IsZero() {
		st = S(rc.Palette[0])
	}

	suffix := ""
	pct := 0.0
	if g.max > 0 && !math.IsNaN(g.value) {
		pct = math.Max(0, math.Min(1, g.value/g.max)) * 100
	}
	if g.showPct {
		suffix = fmt.Sprintf(" %.0f%%", pct)
	} else if g.label != "" {
		suffix = " " + g.label
	}
	if g.label != "" && g.showPct {
		suffix += " " + g.label
	}
	suffix = ellipTrunc(suffix, maxInt(inner.W/2-1, 0), rc.Info.Unicode)

	barW := inner.W - runeLen(suffix)
	if barW < 1 {
		return
	}
	g.drawBar(cv, row, inner.X, barW, pct/100, st, rc.Info.Unicode)
	if suffix != "" {
		cv.Text(inner.X+barW, row, suffix, S(Default))
	}
}

func (g *Gauge) drawBar(cv *Canvas, row, x, w int, frac float64, st Style, uni bool) {
	switch g.style {
	case GaugeASCII:
		gaugeFillRunes(cv, row, x, w, frac, st, '#', '-')
	case GaugeBrackets:
		if w < 5 {
			if uni {
				gaugeFillEighths(cv, row, x, w, frac, st, '█', '░')
			} else {
				gaugeFillRunes(cv, row, x, w, frac, st, '#', '-')
			}
			return
		}
		br := S(Default)
		cv.Set(x, row, '[', br)
		cv.Set(x+w-1, row, ']', br)
		if uni {
			gaugeFillEighths(cv, row, x+1, w-2, frac, st, '█', '░')
		} else {
			gaugeFillRunes(cv, row, x+1, w-2, frac, st, '#', '-')
		}
	case GaugeArrow:
		if uni {
			const fill, tip, track = '━', '╸', '─'
			full := int(math.Ceil(frac * float64(w)))
			if frac > 0 && full == 0 {
				full = 1
			}
			for i := 0; i < w; i++ {
				switch {
				case i < full-1:
					cv.Set(x+i, row, fill, st)
				case i == full-1:
					cv.Set(x+i, row, tip, st.Bolder())
				default:
					cv.Set(x+i, row, track, S(Indexed(240)))
				}
			}
		} else {
			gaugeFillRunes(cv, row, x, w, frac, st, '=', '-')
		}
	case GaugeSegments:
		// Stretch the segments so they span the whole bar instead of
		// leaving dead space before the suffix.
		n := 10
		if w < n {
			n = w
		}
		base := w / n
		rem := w - base*n
		filled := int(math.Round(frac * float64(n)))
		cur := x
		for i := 0; i < n; i++ {
			chunk := base
			if i < rem {
				chunk++
			}
			fullCh, emptyCh := '▰', '▱'
			if !uni {
				fullCh, emptyCh = '#', '-'
			}
			ch := emptyCh
			if i < filled {
				ch = fullCh
			}
			for j := 0; j < chunk; j++ {
				cv.Set(cur, row, ch, st)
				cur++
			}
		}
	default:
		if uni {
			gaugeFillEighths(cv, row, x, w, frac, st, '█', '░')
		} else {
			gaugeFillRunes(cv, row, x, w, frac, st, '#', '-')
		}
	}
}

func gaugeFillRunes(cv *Canvas, row, x, w int, frac float64, st Style, fill, empty rune) {
	full := int(frac * float64(w))
	for i := 0; i < w; i++ {
		ch := empty
		if i < full {
			ch = fill
		}
		cv.Set(x+i, row, ch, st)
	}
}

func gaugeFillEighths(cv *Canvas, row, x, w int, frac float64, st Style, fullRune, empty rune) {
	exact := frac * float64(w)
	nFull := int(exact)
	for i := 0; i < w; i++ {
		ch := empty
		if i < nFull {
			ch = fullRune
		} else if i == nFull {
			idx := int((exact - float64(nFull)) * 8)
			if idx > 0 {
				ch = barEighthsH[idx-1]
			}
		}
		cv.Set(x+i, row, ch, st)
	}
}
