package tuichart

import (
	"time"
)

// TimeSeriesPoint is one measurement: a timestamp and its value.
type TimeSeriesPoint struct {
	At    time.Time
	Value float64
}

// TimeSeriesLine accumulates the points of one series.
type TimeSeriesLine struct {
	ts     *TimeSeries
	name   string
	pts    []TimeSeriesPoint
	marker rune
	color  Color
}

// Add appends a measurement.
func (l *TimeSeriesLine) Add(at time.Time, v float64) *TimeSeriesLine {
	l.pts = append(l.pts, TimeSeriesPoint{At: at, Value: v})
	return l
}

// Point appends a TimeSeriesPoint.
func (l *TimeSeriesLine) Point(p TimeSeriesPoint) *TimeSeriesLine {
	l.pts = append(l.pts, p)
	return l
}

// Color fixes the series color; palette colors cycle when unset.
func (l *TimeSeriesLine) Color(c Color) *TimeSeriesLine { l.color = c; return l }

// Marker sets the point marker rune (0 = none).
func (l *TimeSeriesLine) Marker(m rune) *TimeSeriesLine { l.marker = m; return l }

// TimeSeries renders measurements over time on an X/Y chart: the X axis is
// time (formatted through Go layouts, with span-based automatic presets),
// the Y axis carries the measured values, and consecutive points are
// connected chronologically — the classic peaks-and-troughs view used for
// reports and forecasts.
type TimeSeries struct {
	layout string
	lines  []*TimeSeriesLine
	chartBase
}

// NewTimeSeries creates an empty TimeSeries chart.
func NewTimeSeries() *TimeSeries {
	return &TimeSeries{chartBase: newChartBase()}
}

// Title sets the chart title.
func (ts *TimeSeries) Title(s string) *TimeSeries { ts.SetTitle(s); return ts }

// Line starts a new named series and returns it for chaining Add calls.
func (ts *TimeSeries) Line(name string) *TimeSeriesLine {
	l := &TimeSeriesLine{ts: ts, name: name}
	ts.lines = append(ts.lines, l)
	return l
}

// Format sets the Go time layout used for X tick labels (e.g.
// "2006-01-02", "15:04"). Empty restores automatic presets chosen from the
// covered span.
func (ts *TimeSeries) Format(layout string) *TimeSeries { ts.layout = layout; return ts }

// SetXFormatter overrides tick labels completely (receives unix seconds).
func (ts *TimeSeries) SetXFormatter(f func(float64) string) { ts.xFmt = f }

// ResetFormatters restores default tick formatting.
func (ts *TimeSeries) ResetFormatters() { ts.xFmt, ts.layout = nil, "" }

// HeightHint returns the suggested height in rows for the given width.
func (ts *TimeSeries) HeightHint(width int) int {
	if h := ts.chartBase.HeightHint(width); h > 0 {
		return h
	}
	p := NewPlot()
	p.chartBase = ts.chartBase
	return p.HeightHint(width)
}

// Draw renders the time-series lines onto an X/Y plot in the canvas.
func (ts *TimeSeries) Draw(rc *Ctx, cv *Canvas) {
	p := NewPlot()
	p.chartBase = ts.chartBase

	minU, maxU := float64(0), float64(0)
	first := true
	for _, l := range ts.lines {
		if len(l.pts) == 0 {
			continue
		}
		line := &Line{name: l.name, color: l.color, marker: l.marker}
		for _, pt := range l.pts {
			u := float64(pt.At.Unix())
			line.pts = append(line.pts, Point{X: u, Y: pt.Value})
			if first {
				minU, maxU = u, u
				first = false
			} else {
				minU = realMin(minU, u)
				maxU = realMax(maxU, u)
			}
		}
		p.Add(line)
	}

	if !first {
		fmtFn := ts.xFmt
		if fmtFn == nil {
			layout := ts.layout
			if layout == "" {
				layout = autoTimeLayout(maxU - minU)
			}
			fmtFn = func(v float64) string { return time.Unix(int64(v), 0).Format(layout) }
		}
		p.SetXFormatter(fmtFn)
	}
	p.Draw(rc, cv)
}

func realMin(a, b float64) float64 {
	if b < a {
		return b
	}
	return a
}

func realMax(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
