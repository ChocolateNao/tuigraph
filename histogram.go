package tuichart

import (
	"math"
	"sort"
)

// Histogram bins numeric data and renders a bar chart of the counts.
type Histogram struct {
	name string
	data []float64
	chartBase
	bins  int
	color Color
}

// NewHistogram creates a Histogram from the given data slice.
func NewHistogram(data []float64) *Histogram {
	return &Histogram{
		chartBase: newChartBase(),
		data:      data,
		bins:      defaultBins(len(data)),
		name:      "count",
	}
}

func defaultBins(n int) int {
	if n < 2 {
		return 1
	}
	b := int(math.Ceil(math.Log2(float64(n)) + 1))
	if b < 5 {
		b = 5
	}
	if b > 32 {
		b = 32
	}
	return b
}

// Bins sets the number of bins (1–64) for grouping the data.
func (h *Histogram) Bins(n int) *Histogram {
	if n < 1 {
		n = 1
	}
	if n > 64 {
		n = 64
	}
	h.bins = n
	return h
}

// Color sets the bar fill color; defaults to the first palette entry.
func (h *Histogram) Color(c Color) *Histogram { h.color = c; return h }

// Orientation switches bar direction; see BarChart.Orientation.
func (h *Histogram) Orientation(o Orientation) *Histogram { h.SetOrientation(o); return h }

// Title sets the chart title.
func (h *Histogram) Title(t string) *Histogram { h.SetTitle(t); return h }

// Name sets the series label shown in the legend (default "count").
func (h *Histogram) Name(s string) *Histogram { h.name = s; return h }

// ShowValues toggles printing the numeric count inside each bar (when it fits).
func (h *Histogram) ShowValues(on bool) *Histogram { h.SetShowValues(true); return h }

// Counts bins the data and returns counts plus bin edges.
func (h *Histogram) Counts() (counts []int, edges []float64) {
	if len(h.data) == 0 {
		return nil, nil
	}
	sorted := append([]float64(nil), h.data...)
	sort.Float64s(sorted)
	lo, hi := sorted[0], sorted[len(sorted)-1]
	if hi == lo {
		hi = lo + 1
	}
	width := (hi - lo) / float64(h.bins)
	counts = make([]int, h.bins)
	edges = make([]float64, h.bins+1)
	for i := 0; i <= h.bins; i++ {
		edges[i] = lo + float64(i)*width
	}
	for _, v := range h.data {
		if math.IsNaN(v) {
			continue
		}
		bi := int((v - lo) / width)
		if bi < 0 {
			bi = 0
		}
		if bi >= h.bins {
			bi = h.bins - 1
		}
		counts[bi]++
	}
	return counts, edges
}

// HeightHint returns the suggested height in rows for the given width.
func (h *Histogram) HeightHint(width int) int {
	if hh := h.chartBase.HeightHint(width); hh > 0 {
		return hh
	}
	if h.orient == OrientHorizontal {
		// Delegated draw flips bars sideways; size for bin rows.
		return horizontalBarHeight(h.bins)
	}
	hh := width / 3
	if hh > 20 {
		hh = 20
	}
	if hh < 8 {
		hh = 8
	}
	return hh
}

// Draw bins the data and renders the histogram as a bar chart.
func (h *Histogram) Draw(rc *Ctx, cv *Canvas) {
	counts, edges := h.Counts()
	if counts == nil {
		h.frameTitle(cv, rc.Info.Unicode)
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}
	step := edges[1] - edges[0]
	cats := make([]string, len(counts))
	vals := make([]float64, len(counts))
	for i := range counts {
		if h.xFmt != nil {
			cats[i] = h.xFmt(edges[i])
		} else {
			cats[i] = fmtTick(edges[i], step)
		}
		vals[i] = float64(counts[i])
	}
	bc := NewBar(cats, BarSeries{Name: h.name, Values: vals, Color: h.color})
	bc.chartBase = h.chartBase
	bc.Draw(rc, cv)
}
