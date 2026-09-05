package tuichart

import (
	"math"
	"strings"
)

var (
	sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	sparkASCII  = []rune{'_', '.', '-', '=', '+', '*', '#', '%', '@'}
)

// Sparkline renders a compact one-row line of block characters representing
// a sequence of values — useful for inline or embedded mini-charts.
type Sparkline struct {
	vals []float64
	chartBase
	color Color
}

// NewSpark creates a Sparkline with the given initial values.
func NewSpark(vals ...float64) *Sparkline {
	return &Sparkline{chartBase: newChartBase(), vals: vals}
}

// Title is accepted for symmetry with other diagrams; it renders above the
// sparkline when placed in a Chart row by itself only if height allows.
func (s *Sparkline) Title(t string) *Sparkline { s.SetTitle(t); return s }

// Values appends additional data points to the sparkline.
func (s *Sparkline) Values(
	vals ...float64,
) *Sparkline {
	s.vals = append(s.vals, vals...)
	return s
}

// SetValues replaces all data points in the sparkline.
func (s *Sparkline) SetValues(vals []float64) *Sparkline {
	s.vals = vals
	return s
}

// Color sets the sparkline fill color; defaults to SkyBlue.
func (s *Sparkline) Color(c Color) *Sparkline { s.color = c; return s }

// HeightHint returns the suggested height in rows (1, or 2 with a title).
func (s *Sparkline) HeightHint(int) int {
	if s.title != "" {
		return 2
	}
	return 1
}

// Draw renders the sparkline into the canvas.
func (s *Sparkline) Draw(rc *Ctx, cv *Canvas) {
	row := cv.Height() - 1
	if s.title != "" && cv.Height() >= 2 {
		cv.TextCenter(cv.Width()/2, 0, ellipTrunc(s.title, cv.Width(), rc.Info.Unicode), S(Gray))
	}
	line := renderSpark(s.vals, rc.Info.Unicode, s.color, rc)
	for x, r := range line {
		if x >= cv.Width() {
			break
		}
		cv.Set(x, row, r.r, S(r.c))
	}
}

type sparkRune struct {
	r rune
	c Color
}

func renderSpark(vals []float64, uni bool, c Color, rc *Ctx) []sparkRune {
	n := len(vals)
	out := make([]sparkRune, 0, n)
	blocks := sparkBlocks
	if !uni {
		blocks = sparkASCII
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range vals {
		if math.IsNaN(v) {
			continue
		}
		lo = math.Min(lo, v)
		hi = math.Max(hi, v)
	}
	col := c
	if col.IsZero() && rc != nil {
		col = SkyBlue
	}
	for i := 0; i < n; i++ {
		v := vals[i]
		if math.IsNaN(v) || hi == lo {
			out = append(out, sparkRune{r: ' ', c: col})
			continue
		}
		idx := int((v-lo)/(hi-lo)*float64(len(blocks)-1) + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		out = append(out, sparkRune{r: blocks[idx], c: col})
	}
	return out
}

// Spark renders values as a one-line sparkline using the current terminal profile.
func Spark(vals []float64) string {
	rc := newCtx(Detect())
	runes := renderSpark(vals, rc.Info.Unicode, SkyBlue, rc)
	var b strings.Builder
	cur := Style{}
	active := false
	for _, sr := range runes {
		st := S(sr.c)
		if !st.eq(cur) {
			if active {
				b.WriteString(ansiReset)
				active = false
			}
			seq := rc.Info.Level.seq(st)
			if seq != "" {
				b.WriteString(seq)
				active = true
				cur = st
			}
		}
		b.WriteRune(sr.r)
	}
	if active {
		b.WriteString(ansiReset)
	}
	return b.String()
}
