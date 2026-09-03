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
