package tuichart

import (
	"math"
	"time"
)

// GanttBar is one activity on a Gantt chart.
type GanttBar struct {
	Start time.Time
	End   time.Time
	Name  string
	Color Color
}

// Gantt renders activities as horizontal bars spread across a time axis:
// each bar spans from its start date to its end date, one activity per row.
// Tick labels use Go time layouts — set one with Format, override ticks via
// SetXFormatter, or leave unset for span-based presets.
type Gantt struct {
	layout string
	bars   []GanttBar
	chartBase
}

// NewGantt creates an empty Gantt chart.
func NewGantt() *Gantt {
	return &Gantt{chartBase: newChartBase()}
}

// Title sets the chart title.
func (g *Gantt) Title(s string) *Gantt { g.SetTitle(s); return g }

// Bar appends an activity spanning [start, end].
func (g *Gantt) Bar(name string, start, end time.Time) *Gantt {
	g.bars = append(g.bars, GanttBar{Name: name, Start: start, End: end})
	return g
}

// Duration appends an activity given a start and a duration.
func (g *Gantt) Duration(name string, start time.Time, d time.Duration) *Gantt {
	return g.Bar(name, start, start.Add(d))
}

// Task appends a full GanttBar (allows per-bar color).
func (g *Gantt) Task(b GanttBar) *Gantt {
	g.bars = append(g.bars, b)
	return g
}

// Format sets the Go time layout for axis ticks; empty restores auto.
func (g *Gantt) Format(layout string) *Gantt { g.layout = layout; return g }

// HeightHint returns the suggested height in rows for the given width.
func (g *Gantt) HeightHint(width int) int {
	if h := g.chartBase.HeightHint(width); h > 0 {
		return h
	}
	rows := len(g.bars)
	if rows < 4 {
		rows = 4
	}
	if rows > 24 {
		rows = 24
	}
	return rows + 3 // grid + tick row + frame
}

// Draw renders the activity bars and time axis into the canvas.
func (g *Gantt) Draw(rc *Ctx, cv *Canvas) {
	inner := g.frameTitle(cv, rc.Info.Unicode)
	uni := rc.Info.Unicode
	if len(g.bars) == 0 || inner.W < 10 || inner.H < 3 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}

	minU := math.Inf(1)
	maxU := math.Inf(-1)
	for _, b := range g.bars {
		s, e := float64(b.Start.Unix()), float64(b.End.Unix())
		if e < s {
			s, e = e, s
		}
		minU = math.Min(minU, s)
		maxU = math.Max(maxU, e)
	}
	pad := (maxU - minU) * 0.03
	if pad <= 0 {
		pad = 60
	}
	minU -= pad
	maxU += pad

	fmtFn := g.xFmt
	if fmtFn == nil {
		layout := g.layout
		if layout == "" {
			layout = autoTimeLayout(maxU - minU)
		}
		fmtFn = func(v float64) string { return time.Unix(int64(v), 0).Format(layout) }
	}

	gutter := 1
	for _, b := range g.bars {
		if n := runeLen(b.Name); n+1 > gutter && n*2 < inner.W {
			gutter = n + 1
		}
	}

	gridW := inner.X2() - inner.X - gutter + 1
	mapCol := func(u float64) int {
		f := (u - minU) / (maxU - minU)
		return inner.X + gutter + int(f*float64(gridW-1)+0.5)
	}

	dim := S(DimGray)
	tickRow := inner.Y2()
	barCh := '█'
	if !uni {
		barCh = '#'
	}
	ci := 0
	for i, b := range g.bars {
		y := inner.Y + i
		if y >= tickRow {
			break
		}
		st := S(b.Color)
		if b.Color.IsZero() {
			st = S(rc.Palette[ci%len(rc.Palette)])
			ci++
		}
		if gutter > 1 {
			cv.TextRight(inner.X+gutter-2, y, ellipTrunc(b.Name, gutter-1, uni), S(Default))
		}
		s, e := float64(b.Start.Unix()), float64(b.End.Unix())
		if e < s {
			s, e = e, s
		}
		x0, x1 := mapCol(s), mapCol(e)
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		for x := x0; x <= x1 && x <= inner.X2(); x++ {
			cv.Set(x, y, barCh, st)
		}
	}

	for _, tk := range niceTicks(minU, maxU, g.tickN) {
		col := mapCol(tk.Value)
		if col < inner.X+gutter || col > inner.X2() {
			continue
		}
		ch := '┬'
		if !uni {
			ch = '+'
		}
		cv.Set(col, tickRow, ch, dim)
		lbl := fmtFn(tk.Value)
		writeLabel(cv, tickRow, col, lbl, S(Default), inner, uni)
	}
}
