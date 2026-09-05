package tuichart

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func renderD(d Drawable, opts ...Option) string {
	g := New(append([]Option{WithNoColor(), WithUnicode(true), WithWidth(60)}, opts...)...)
	g.Add(d)
	return g.Render()
}

func firstLine(s string, n int) string {
	lines := splitLines(s)
	if n < len(lines) {
		lines = lines[:n]
	}
	return strings.Join(lines, "|")
}

func runeCol(s string, substr string) int {
	i := strings.Index(s, substr)
	if i < 0 {
		return -1
	}
	return runeLen(s[:i])
}

func TestTitleAlignDiagram(t *testing.T) {
	p := NewPlot().Title("TITLE_X")
	p.Add(NewLineVals("s", []float64{1, 2}))
	p.SetTitleAlign(AlignRight)
	right := renderD(p)
	if idx := runeCol(right, "TITLE_X"); idx < 30 {
		t.Errorf("right-aligned title at col %d", idx)
	}
	p.SetTitleAlign(AlignLeft)
	left := renderD(p)
	if i := runeCol(left, "TITLE_X"); i > 6 || i < 0 {
		t.Errorf("left-aligned title at col %d", i)
	}
	p.ResetTitleAlign()
	if p.titleAlign != AlignLeft {
		t.Error("ResetTitleAlign failed")
	}
}

func TestTitleAlignChartContainer(t *testing.T) {
	g := New(WithNoColor(), WithWidth(50)).Title("CONTAINER")
	g.Add(NewPlot())
	center := g.Render()
	if runeCol(center, "CONTAINER") < 15 {
		t.Error("default container title should be centered")
	}
	g.TitleAlign(AlignRight)
	if i := runeCol(g.Render(), "CONTAINER"); i < 40 {
		t.Errorf("right container title at %d", i)
	}
	g.TitleAlign(AlignLeft)
	if i := runeCol(g.Render(), "CONTAINER"); i != 0 {
		t.Errorf("left container title at %d", i)
	}
}

func testBar() *BarChart {
	return NewBarValues([]string{"a", "b", "c"}, []float64{3, 6, 2}).Title("bars")
}

func TestOrientationBars(t *testing.T) {
	b := testBar()
	vert := renderD(b)
	b.SetOrientation(OrientHorizontal)
	horiz := renderD(b)
	if vert == horiz {
		t.Fatal("orientation had no effect")
	}
	if !strings.Contains(horiz, "a ") && !strings.Contains(horiz, " a ") {
		t.Errorf("horizontal bars missing category gutter: %q", firstLine(horiz, 3))
	}
	b.SetOrientation(OrientVertical)
	if back := renderD(b); back != vert {
		t.Error("OrientVertical did not restore vertical layout")
	}
	b.ResetOrientation()
	b.Horizontal(true)
	if legacy := renderD(b); legacy != horiz {
		t.Error("legacy Horizontal(true) diverges from OrientHorizontal")
	}
}

func TestHistogramOrientation(t *testing.T) {
	data := []float64{1, 2, 2, 3, 3, 3, 4, 4, 5}
	h := NewHistogram(data).Bins(4).Title("hist")
	vert := renderD(h)
	h.Orientation(OrientHorizontal)
	horiz := renderD(h)
	if vert == horiz {
		t.Error("histogram orientation ignored")
	}
}

func TestPlotSwapAxes(t *testing.T) {
	p := NewPlot().Title("swap")
	p.Add(NewLineVals("s", []float64{1, 5, 3}))
	p.XLabel("xx").YLabel("yy")
	norm := renderD(p)
	p.SetOrientation(OrientHorizontal)
	swapped := renderD(p)
	if norm == swapped {
		t.Fatal("plot swap had no effect")
	}
	p.SetYRange(-100, 100) // becomes the x range after swap
	out := renderD(p)
	if strings.Contains(out, "-100") == false && strings.Contains(out, "100") == false {
		t.Errorf("swapped x-range ticks missing: %q", out)
	}
	p.ResetOrientation()
	p.ResetScale()
	if renderD(p) != norm {
		t.Error("ResetOrientation did not restore plot")
	}
}

func timelineFixture() *Timeline {
	base := time.Date(2026, 8, 23, 9, 0, 0, 0, time.UTC)
	return NewTimeline().Title("tl").
		Event(base.Add(-2*time.Hour), "alpha").
		Event(base.Add(-30*time.Minute), "beta").
		Event(base, "gamma")
}

func TestTimelineRendersEvents(t *testing.T) {
	tl := timelineFixture().Format("15:04")
	out := renderD(tl)
	for _, want := range []string{"alpha", "beta", "gamma", "◆"} {
		if !strings.Contains(out, want) {
			t.Errorf("timeline missing %q", want)
		}
	}
}

func TestTimelineTimeFormats(t *testing.T) {
	tl := timelineFixture()
	dayOnly := renderD(tl.Format("2006-01-02"))
	if !strings.Contains(dayOnly, "2026-08-23") {
		t.Errorf("custom layout ignored: %q", dayOnly)
	}
	auto := renderD(timelineFixture()) // span ~2h -> HH:MM preset
	if !strings.Contains(auto, ":") {
		t.Errorf("auto time format missing clock labels: %q", auto)
	}
	tl2 := timelineFixture()
	tl2.SetXFormatter(func(v float64) string { return "T" })
	if out := renderD(tl2); !strings.Contains(out, "T") {
		t.Error("SetXFormatter override ignored by timeline")
	}
}

func TestTimelineASCIIFallback(t *testing.T) {
	tl := timelineFixture()
	out := renderD(tl, WithUnicode(false))
	if strings.ContainsAny(out, "◆─┬") {
		t.Errorf("ascii fallback leaked unicode: %q", out)
	}
	if !strings.Contains(out, "*") {
		t.Error("ascii marker missing")
	}
}

func gaugeByStyle(s GaugeStyle) *Gauge {
	return NewGauge(6, 10).Max(10).Style(s).ShowPercent(false).Title("")
}

func TestGaugeColorReturnsReceiver(t *testing.T) {
	g := NewGauge(5, 10)
	ret := g.Color(Red)
	if ret != g {
		t.Error("Color did not return receiver")
	}
	if g.color != Red {
		t.Errorf("color = %v, want Red", g.color)
	}
}

func TestGaugeColorRendering(t *testing.T) {
	def := renderD(NewGauge(5, 10).ShowPercent(false))
	red := renderD(NewGauge(5, 10).ShowPercent(false).Color(Red), WithColor256())
	if def == red {
		t.Error("Color setter had no effect on render")
	}
	if !strings.Contains(red, "█") {
		t.Errorf("colored gauge missing fill:\n%s", red)
	}
}

func TestGaugeStyleDistinct(t *testing.T) {
	styled := map[GaugeStyle]string{
		GaugeBlocks:   renderD(gaugeByStyle(GaugeBlocks)),
		GaugeASCII:    renderD(gaugeByStyle(GaugeASCII)),
		GaugeBrackets: renderD(gaugeByStyle(GaugeBrackets)),
		GaugeArrow:    renderD(gaugeByStyle(GaugeArrow)),
		GaugeSegments: renderD(gaugeByStyle(GaugeSegments)),
	}
	seen := map[string]bool{}
	for st, out := range styled {
		if seen[out] {
			t.Errorf("style %d duplicates another", st)
		}
		seen[out] = true
	}
	if !strings.Contains(styled[GaugeBrackets], "[") {
		t.Error("brackets style missing [")
	}
	if !strings.Contains(styled[GaugeASCII], "#") {
		t.Error("ascii style missing #")
	}
	if !strings.Contains(styled[GaugeSegments], "▰") {
		t.Error("segments style missing ▰")
	}
}

func TestGaugePercentAndClamp(t *testing.T) {
	out := renderD(NewGauge(0.62, 1))
	if !strings.Contains(out, "62%") {
		t.Errorf("percent missing: %q", out)
	}
	over := renderD(NewGauge(999, 1))
	if !strings.Contains(over, "100%") {
		t.Errorf("clamp to 100%% failed: %q", over)
	}
	neg := renderD(NewGauge(-5, 1))
	if !strings.Contains(neg, "0%") {
		t.Errorf("negative clamp failed: %q", neg)
	}
	noPct := renderD(NewGauge(5, 10).ShowPercent(false).Label("fuel"))
	if !strings.Contains(noPct, "fuel") || strings.Contains(noPct, "%") {
		t.Errorf("label mode wrong: %q", noPct)
	}
}

func TestGaugeHeightHint(t *testing.T) {
	if h := NewGauge(1, 2).HeightHint(40); h != 1 {
		t.Errorf("plain hint=%d", h)
	}
	if h := NewGauge(1, 2).Title("t").HeightHint(40); h != 3 {
		t.Errorf("titled hint=%d", h)
	}
	g := NewGauge(1, 2)
	g.SetSize(5)
	if g.HeightHint(40) != 5 {
		t.Error("SetSize override ignored")
	}
}

func TestResetClearsNewOptions(t *testing.T) {
	p := NewPlot()
	p.SetTitleAlign(AlignRight)
	p.SetOrientation(OrientHorizontal)
	p.Reset()
	if p.titleAlign != AlignLeft || p.orient != OrientAuto {
		t.Error("Reset did not clear alignment/orientation")
	}
}

func TestFormattersApplied(t *testing.T) {
	p := NewPlot().Add(NewLineVals("l", []float64{1, 2, 3}))
	p.SetYFormatter(func(v float64) string { return fmt.Sprintf("%d%%", int(v)) })
	out := renderD(p)
	if !strings.Contains(out, "3%") {
		t.Errorf("y formatter ignored on plot:\n%s", out)
	}

	h := NewHistogram([]float64{1, 2, 2, 3, 3, 3, 9}).Bins(4)
	h.SetXFormatter(func(v float64) string { return fmt.Sprintf("%.0fms", v) })
	out = renderD(h, WithWidth(70))
	if !strings.Contains(out, "ms") {
		t.Errorf("x formatter ignored by histogram:\n%s", out)
	}
}
