package tuichart

import (
	"math"
	"strings"
	"testing"
)

func monoPlot() *Plot {
	p := NewPlot()
	p.Add(NewLineVals("a", []float64{1, 3, 2, 5, 4}))
	return p
}

func renderMono(d Drawable, w int) string {
	g := New(WithWidth(w), WithNoColor(), WithUnicode(true))
	g.Add(d)
	return g.Render()
}

func renderASCII(d Drawable, w int) string {
	g := New(WithWidth(w), WithNoColor(), WithUnicode(false))
	g.Add(d)
	return g.Render()
}

func seriesPlot(ss ...any) *Plot {
	p := NewPlot()
	p.Add(ss...)
	return p
}

func TestPlotRenderDeterministic(t *testing.T) {
	a := renderMono(monoPlot(), 50)
	b := renderMono(monoPlot(), 50)
	if a != b {
		t.Error("render not deterministic")
	}
	if !strings.Contains(a, "┌") || !strings.Contains(a, "└") {
		t.Error("missing frame border")
	}
	if !strings.Contains(a, "a") {
		t.Error("legend missing series name")
	}
}

func TestPlotMaybeSwapHorizontal(t *testing.T) {
	p := NewPlot()
	p.Add(NewLineVals("line", []float64{1, 5, 3}),
		NewScatterVals("scat", []float64{2, 4, 1}))
	norm := renderMono(p, 40)
	p.Orientation(OrientHorizontal)
	horiz := renderMono(p, 40)
	if norm == horiz {
		t.Error("horizontal plot swap had no effect")
	}
	if !strings.Contains(horiz, "line") || !strings.Contains(horiz, "scat") {
		t.Error("horizontal plot legend missing series")
	}
}

func TestPlotMaybeSwapNoData(t *testing.T) {
	p := NewPlot()
	p.Add(NewLine("line"), NewScatter("scat"))
	p.Orientation(OrientHorizontal)
	out := renderMono(p, 40)
	if strings.Contains(out, "NaN") || strings.Contains(out, "+Inf") {
		t.Errorf("empty-series swap produced bad output:\n%s", out)
	}
}

type fakeSeries struct{}

func (fakeSeries) seriesName() string                  { return "fake" }
func (fakeSeries) bounds(db *dataBounds)               { db.add(0, 0); db.add(1, 1) }
func (fakeSeries) hasColor() bool                      { return true }
func (fakeSeries) setColor(Color)                      {}
func (fakeSeries) colorOf() Color                      { return Red }
func (fakeSeries) draw(cv *Canvas, fr frame, st Style) {}
func (fakeSeries) legendEntry(Style, bool) LegendEntry {
	return LegendEntry{Label: "fake", Glyph: "x"}
}

// maybeSwap returns non-Line/Scatter series untouched even when swapping.
func TestPlotMaybeSwapPassthrough(t *testing.T) {
	s := fakeSeries{}
	if got := maybeSwap(s, true); got != s {
		t.Error("non-Line/Scatter series should pass through untouched")
	}
	line := NewLine("l", Point{X: 1, Y: 2})
	if got := maybeSwap(line, false); got != line {
		t.Error("swap=false should return series as-is")
	}
}

func TestPlotYRangeClip(t *testing.T) {
	p := monoPlot()
	p.SetYRange(0, 100)
	out := renderMono(p, 40)
	if strings.Contains(out, "NaN") || strings.Contains(out, "+Inf") {
		t.Error("clipping produced non-finite output markers")
	}
}

func TestPlotResetScale(t *testing.T) {
	p := monoPlot()
	auto := renderMono(p, 40)
	p.SetYRange(-1000, 1000)
	fixed := renderMono(p, 40)
	if auto == fixed {
		t.Fatal("SetYRange had no effect")
	}
	p.ResetScale()
	back := renderMono(p, 40)
	if back != auto {
		t.Error("ResetScale did not restore autoscale")
	}
}

func TestPlotLogYWithNonPositive(t *testing.T) {
	p := NewPlot()
	p.LogY(true)
	p.Add(NewLineVals("log", []float64{-1, 0, 1, 10, 100}))
	out := renderMono(p, 40)
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestPlotEmpty(t *testing.T) {
	out := renderMono(NewPlot(), 30)
	if !strings.Contains(out, "(no data)") {
		t.Errorf("empty plot: %q", out)
	}
}

func TestPlotFunction(t *testing.T) {
	fn := NewFunction(math.Sin).Domain(-math.Pi, math.Pi).Name("sin(x)")
	out := renderMono(fn, 60)
	if !strings.Contains(out, "sin(x)") {
		t.Error("function legend missing")
	}
	if !strings.Contains(out, "-1") || !strings.Contains(out, "1") {
		t.Error("expected -1/1 ticks for sin domain")
	}
}

func TestScatterMarkers(t *testing.T) {
	p := NewPlot()
	sc := NewScatterVals("pts", []float64{1, 2, 3}).Marker('x')
	p.Add(sc)
	out := renderMono(p, 30)
	if !strings.Contains(out, "x") {
		t.Error("explicit marker not rendered")
	}
}

func TestZipEqualLength(t *testing.T) {
	pts := Zip([]float64{1, 2, 3}, []float64{10, 20, 30})
	if len(pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(pts))
	}
	for i, want := range []struct{ x, y float64 }{{1, 10}, {2, 20}, {3, 30}} {
		if pts[i].X != want.x || pts[i].Y != want.y {
			t.Errorf("pts[%d] = %v, want %v", i, pts[i], want)
		}
	}
}

func TestZipUnequalLength(t *testing.T) {
	pts := Zip([]float64{1, 2, 3, 4}, []float64{10, 20})
	if len(pts) != 2 {
		t.Fatalf("expected 2 points (truncated), got %d", len(pts))
	}
	if pts[0].X != 1 || pts[0].Y != 10 {
		t.Errorf("pts[0] = %v", pts[0])
	}
	if pts[1].X != 2 || pts[1].Y != 20 {
		t.Errorf("pts[1] = %v", pts[1])
	}
}

func TestZipEmptyInputs(t *testing.T) {
	pts := Zip(nil, nil)
	if len(pts) != 0 {
		t.Fatalf("expected 0 points, got %d", len(pts))
	}
	pts = Zip([]float64{1, 2}, nil)
	if len(pts) != 0 {
		t.Fatalf("expected 0 points from nil ys, got %d", len(pts))
	}
	pts = Zip(nil, []float64{1, 2})
	if len(pts) != 0 {
		t.Fatalf("expected 0 points from nil xs, got %d", len(pts))
	}
	pts = Zip([]float64{}, []float64{1})
	if len(pts) != 0 {
		t.Fatalf("expected 0 points from empty xs, got %d", len(pts))
	}
}

func TestLineColor(t *testing.T) {
	l := NewLineVals("c", []float64{1, 2}).Color(Red)
	if !l.hasColor() {
		t.Error("Color setter did not apply")
	}
	if l.colorOf() != Red {
		t.Error("colorOf mismatch")
	}
	l.setColor(Blue)
	if l.colorOf() != Red {
		t.Error("setColor overwrote existing color")
	}
	out := renderASCII(seriesPlot(l), 40)
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestLineMarker(t *testing.T) {
	l := NewLineVals("m", []float64{1, 3, 2}).Marker('*')
	out := renderASCII(seriesPlot(l), 40)
	if !strings.Contains(out, "*") {
		t.Error("marker glyph not found in render")
	}
}

func TestLineDashed(t *testing.T) {
	l := NewLineVals("d", []float64{1, 2, 3}).Dashed(true)
	if !l.dashed {
		t.Error("Dashed(true) did not set flag")
	}
	l.Dashed(false)
	if l.dashed {
		t.Error("Dashed(false) did not clear flag")
	}
	out := renderASCII(seriesPlot(l), 40)
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestLineSetName(t *testing.T) {
	l := NewLineVals("old", []float64{1, 2}).SetName("new")
	if l.seriesName() != "new" {
		t.Errorf("seriesName = %q, want %q", l.seriesName(), "new")
	}
	out := renderASCII(seriesPlot(l), 40)
	if !strings.Contains(out, "new") {
		t.Error("renamed series not in legend")
	}
	if strings.Contains(out, "old") {
		t.Error("old name still present")
	}
}

func TestLinePoints(t *testing.T) {
	l := NewLine("p")
	l.Points(Point{X: 0, Y: 1}, Point{X: 1, Y: 2})
	if len(l.pts) != 2 {
		t.Fatalf("expected 2 points, got %d", len(l.pts))
	}
	if l.pts[1].X != 1 || l.pts[1].Y != 2 {
		t.Errorf("unexpected point: %v", l.pts[1])
	}
}

func TestLineValues(t *testing.T) {
	l := NewLine("v")
	l.Values(10, 20, 30)
	if len(l.pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(l.pts))
	}
	for i, want := range []float64{10, 20, 30} {
		if l.pts[i].Y != want {
			t.Errorf("pts[%d].Y = %v, want %v", i, l.pts[i].Y, want)
		}
		if l.pts[i].X != float64(i) {
			t.Errorf("pts[%d].X = %v, want %v", i, l.pts[i].X, float64(i))
		}
	}
}

func TestNewScatter(t *testing.T) {
	sc := NewScatter("s", Point{X: 5, Y: 10}, Point{X: 6, Y: 20})
	if sc.seriesName() != "s" {
		t.Errorf("name = %q", sc.seriesName())
	}
	if len(sc.pts) != 2 {
		t.Fatalf("expected 2 points, got %d", len(sc.pts))
	}
	if sc.pts[0].X != 5 || sc.pts[0].Y != 10 {
		t.Errorf("point 0 = %v", sc.pts[0])
	}
}

func TestNewScatterVals(t *testing.T) {
	sc := NewScatterVals("sv", []float64{4, 8, 12})
	if sc.seriesName() != "sv" {
		t.Errorf("name = %q", sc.seriesName())
	}
	if len(sc.pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(sc.pts))
	}
	for i, want := range []float64{4, 8, 12} {
		if sc.pts[i].Y != want || sc.pts[i].X != float64(i) {
			t.Errorf("pts[%d] = %v, want X=%d Y=%v", i, sc.pts[i], i, want)
		}
	}
}

func TestScatterColor(t *testing.T) {
	sc := NewScatterVals("sc", []float64{1, 2}).Color(Cyan)
	if !sc.hasColor() {
		t.Error("Color setter did not apply")
	}
	if sc.colorOf() != Cyan {
		t.Error("colorOf mismatch")
	}
	// setColor should not overwrite
	sc.setColor(Purple)
	if sc.colorOf() != Cyan {
		t.Error("setColor overwrote existing color")
	}
	out := renderASCII(seriesPlot(sc), 40)
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestScatterMarker(t *testing.T) {
	sc := NewScatterVals("sm", []float64{1, 2, 3}).Marker('+')
	out := renderASCII(seriesPlot(sc), 40)
	if !strings.Contains(out, "+") {
		t.Error("marker glyph not found in render")
	}
}

func TestScatterSetName(t *testing.T) {
	sc := NewScatterVals("old", []float64{1, 2}).SetName("fresh")
	if sc.seriesName() != "fresh" {
		t.Errorf("seriesName = %q", sc.seriesName())
	}
	out := renderASCII(seriesPlot(sc), 40)
	if !strings.Contains(out, "fresh") {
		t.Error("renamed scatter not in legend")
	}
}

func TestScatterPoints(t *testing.T) {
	sc := NewScatter("sp")
	sc.Points(Point{X: 10, Y: 20}, Point{X: 30, Y: 40})
	if len(sc.pts) != 2 {
		t.Fatalf("expected 2 points, got %d", len(sc.pts))
	}
	if sc.pts[1].X != 30 || sc.pts[1].Y != 40 {
		t.Errorf("unexpected point: %v", sc.pts[1])
	}
}

func TestPlotLogX(t *testing.T) {
	p := NewPlot()
	p.LogX(true)
	p.Add(NewLineVals("lx", []float64{1, 10, 100}))
	out := renderASCII(p, 40)
	if out == "" {
		t.Fatal("empty render with LogX")
	}
	p.LogX(false)
	out2 := renderASCII(p, 40)
	if out2 == "" {
		t.Fatal("empty render after LogX(false)")
	}
}

func TestPlotLogYToggle(t *testing.T) {
	p := NewPlot()
	p.LogY(true)
	p.Add(NewLineVals("ly", []float64{1, 10, 100}))
	on := renderASCII(p, 40)
	p.LogY(false)
	off := renderASCII(p, 40)
	if on == off {
		t.Error("LogY toggle had no visible effect")
	}
}

func TestPlotOrientationHorizontal(t *testing.T) {
	p := NewPlot()
	p.Add(NewLineVals("h", []float64{1, 3, 2, 5, 4}))
	v := renderASCII(p, 40)
	p.Orientation(OrientHorizontal)
	h := renderASCII(p, 40)
	if v == h {
		t.Error("OrientHorizontal had no visible effect")
	}
	p.Orientation(OrientVertical)
	back := renderASCII(p, 40)
	if back != v {
		t.Error("reverting orientation did not restore original render")
	}
}

func TestPlotGrid(t *testing.T) {
	p := NewPlot()
	p.Add(NewLineVals("g", []float64{1, 2, 3}))
	on := renderASCII(p, 40)
	p.Grid(false)
	off := renderASCII(p, 40)
	if on == off {
		t.Error("Grid toggle had no visible effect")
	}
}

func TestPlotLegend(t *testing.T) {
	p := NewPlot()
	p.Add(NewLineVals("lg", []float64{1, 2, 3}))
	// Legend returns the receiver for chaining
	got := p.Legend(false)
	if got != p {
		t.Error("Legend did not return the receiver")
	}
	// Line series legend entry is shown by default
	le := NewLineVals("lg", nil).legendEntry(S(Gray), false)
	if le.Glyph != "---" {
		t.Errorf("ascii line glyph = %q, want %q", le.Glyph, "---")
	}
	// Scatter legend entry
	se := NewScatterVals("sg", nil).legendEntry(S(Gray), false)
	if se.Glyph != "oo" {
		t.Errorf("ascii scatter glyph = %q, want %q", se.Glyph, "oo")
	}
	// Render must not panic with legend toggled off
	out := renderASCII(p, 40)
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestPlotHeightHint(t *testing.T) {
	p := NewPlot()
	// Default: no height set, should return clamped proportion of width
	h := p.HeightHint(40)
	if h < 9 || h > 24 {
		t.Errorf("default HeightHint(40) = %d, want 9..24", h)
	}
	// Small width => minimum hint
	h = p.HeightHint(10)
	if h != 9 {
		t.Errorf("HeightHint(10) = %d, want 9 (min)", h)
	}
	// Very large width => capped at 24
	h = p.HeightHint(200)
	if h != 24 {
		t.Errorf("HeightHint(200) = %d, want 24 (max)", h)
	}
	// Height explicitly set => base returns it
	p.Height(15)
	h = p.HeightHint(40)
	if h != 15 {
		t.Errorf("HeightHint with Height(15) = %d, want 15", h)
	}
}

func TestPlotAddInvalidType(t *testing.T) {
	p := NewPlot()
	p.Add("not a series", 42)
	if len(p.order) != 0 {
		t.Errorf("expected 0 series after invalid Add, got %d", len(p.order))
	}
	out := renderASCII(p, 30)
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestScatterDefaultMarker(t *testing.T) {
	sc := NewScatterVals("dm", []float64{1, 2, 3})
	out := renderASCII(seriesPlot(sc), 40)
	if !strings.Contains(out, "o") {
		t.Error("default ASCII scatter marker 'o' not found")
	}
}

func TestLineEmptyPoints(t *testing.T) {
	l := NewLine("empty")
	out := renderASCII(seriesPlot(l), 40)
	if out == "" {
		t.Fatal("empty line plot failed to render")
	}
}

func TestScatterEmptyPoints(t *testing.T) {
	sc := NewScatter("empty")
	out := renderASCII(seriesPlot(sc), 40)
	if out == "" {
		t.Fatal("empty scatter plot failed to render")
	}
}

func TestSeqBasic(t *testing.T) {
	pts := Seq(10, 20, 30)
	if len(pts) != 3 {
		t.Fatalf("expected 3, got %d", len(pts))
	}
	for i, want := range []float64{10, 20, 30} {
		if pts[i].X != float64(i) || pts[i].Y != want {
			t.Errorf("pts[%d] = %v, want {X:%d Y:%v}", i, pts[i], i, want)
		}
	}
}

func TestSeqEmpty(t *testing.T) {
	pts := Seq()
	if len(pts) != 0 {
		t.Fatalf("expected 0, got %d", len(pts))
	}
}
