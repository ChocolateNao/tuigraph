package tuichart

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func smallPlot() *Plot {
	p := NewPlot()
	p.Add(NewLineVals("s", []float64{1, 2, 3}))
	return p
}

func TestChartStackedRows(t *testing.T) {
	g := New(WithWidth(40), WithNoColor())
	g.Add(smallPlot())
	one := g.Render()
	g.Clear()
	g.Add(smallPlot())
	g.Add(smallPlot())
	two := g.Render()
	h1 := strings.Count(one, "\n")
	h2 := strings.Count(two, "\n")
	if h2 < h1*2-2 {
		t.Errorf("stacked height %d, want ~2x %d", h2, h1)
	}
}

func TestChartRowSideBySide(t *testing.T) {
	g := New(WithWidth(80), WithNoColor(), WithUnicode(false))
	g.Row(smallPlot(), smallPlot())
	out := g.Render()
	lines := splitLines(out)
	w := runeLen(lines[0])
	if w > 82 {
		t.Errorf("row width %d exceeds 80", w)
	}
	// both plots visible side by side: two left borders on a mid line
	mid := lines[len(lines)/2]
	if strings.Count(mid, "│")+strings.Count(mid, "|") < 3 {
		t.Errorf("expected 3+ frame edges in middle row: %q", mid)
	}
}

func TestChartWidthOption(t *testing.T) {
	for _, w := range []int{30, 60, 100} {
		g := New(WithWidth(w), WithNoColor())
		g.Add(smallPlot())
		lines := splitLines(g.Render())
		if got := runeLen(lines[0]); got != w {
			t.Errorf("width %d: got %d", w, got)
		}
	}
}

func TestChartClearLenReset(t *testing.T) {
	g := New()
	if g.Len() != 0 {
		t.Error("new chart not empty")
	}
	g.Add(smallPlot())
	g.Row(smallPlot(), smallPlot())
	if g.Len() != 3 {
		t.Errorf("Len=%d want 3 (diagrams, not rows)", g.Len())
	}
	g.Reset()
	if g.Len() != 0 {
		t.Error("Reset did not clear")
	}
}

func TestChartTitleAndString(t *testing.T) {
	g := New(WithWidth(40), WithNoColor()).Title("mytitle")
	g.Add(smallPlot())
	if !strings.Contains(g.String(), "mytitle") {
		t.Error("chart title missing")
	}
}

func TestUnicodeOverrideOption(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithUnicode(true))
	g.Add(smallPlot())
	out := g.Render()
	if !strings.Contains(out, "┌") {
		t.Error("WithUnicode(true) ignored")
	}
	g2 := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g2.Add(smallPlot())
	if strings.Contains(g2.Render(), "┌") {
		t.Error("WithUnicode(false) ignored")
	}
}

// --- chart.go option edge cases ---

func TestWithGap(t *testing.T) {
	// Gap of 0: no blank rows between diagrams.
	g0 := New(WithWidth(40), WithNoColor(), WithUnicode(false), WithGap(0))
	g0.Add(smallPlot())
	g0.Add(smallPlot())
	out0 := g0.Render()

	// Gap of 3: three blank rows between diagrams.
	g3 := New(WithWidth(40), WithNoColor(), WithUnicode(false), WithGap(3))
	g3.Add(smallPlot())
	g3.Add(smallPlot())
	out3 := g3.Render()

	lines0 := strings.Count(out0, "\n")
	lines3 := strings.Count(out3, "\n")
	if lines3 <= lines0 {
		t.Errorf("gap=3 output (%d lines) should be longer than gap=0 (%d lines)", lines3, lines0)
	}
}

func TestWithGapNegative(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithGap(-5))
	g.Add(smallPlot())
	g.Add(smallPlot())
	// Must not panic; negative treated as 0 by maxInt.
	out := g.Render()
	if out == "" {
		t.Error("empty output with negative gap")
	}
}

func TestWithPalette(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithPalette(Red, Blue, Green))
	g.Add(smallPlot())
	out := g.Render()
	if out == "" {
		t.Error("empty output with custom palette")
	}
}

func TestWithPaletteEmpty(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithPalette())
	g.Add(smallPlot())
	out := g.Render()
	if out == "" {
		t.Error("empty output with empty palette")
	}
}

func TestWithColor16(t *testing.T) {
	g := New(WithWidth(40), WithColor16(), WithUnicode(false))
	g.Add(smallPlot())
	out := g.Render()
	if out == "" {
		t.Error("empty output with WithColor16")
	}
	// WithColor16 should produce ANSI escape sequences.
	if !strings.Contains(out, "\x1b[") {
		t.Error("WithColor16: expected ANSI escapes in output")
	}
}

func TestWithTrueColor(t *testing.T) {
	g := New(WithWidth(40), WithTrueColor(), WithUnicode(false))
	g.Add(smallPlot())
	out := g.Render()
	if out == "" {
		t.Error("empty output with WithTrueColor")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Error("WithTrueColor: expected ANSI escapes in output")
	}
}

func TestWriteTo(t *testing.T) {
	var buf bytes.Buffer
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	n, err := g.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n <= 0 {
		t.Errorf("WriteTo wrote %d bytes", n)
	}
	if buf.Len() == 0 {
		t.Error("WriteTo: buffer empty")
	}
	if int(n) != buf.Len() {
		t.Errorf("WriteTo: n=%d but buf.Len()=%d", n, buf.Len())
	}
}

func TestWriteToShortWrite(t *testing.T) {
	sw := &chartShortWriter{max: 2}
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	_, err := g.WriteTo(sw)
	if err == nil {
		t.Error("WriteTo to short writer should return error")
	}
}

func TestRenderTo(t *testing.T) {
	var buf bytes.Buffer
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	err := g.RenderTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("RenderTo: buffer empty")
	}
}

func TestRenderToShortWrite(t *testing.T) {
	sw := &chartShortWriter{max: 2}
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	err := g.RenderTo(sw)
	if err == nil {
		t.Error("RenderTo to short writer should return error")
	}
}

func TestRenderToWithWidth(t *testing.T) {
	var buf bytes.Buffer
	g := New(WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	err := g.RenderTo(&buf, 50)
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(buf.String())
	if len(lines) == 0 {
		t.Error("RenderTo with width: no output")
	}
}

func TestDefaultDiagramHeight(t *testing.T) {
	tests := []struct {
		w, want int
	}{
		{0, 9},   // clamped to min 9
		{5, 9},   // below min
		{27, 9},  // 27/3 = 9 exactly
		{30, 10}, // 30/3 = 10
		{60, 20}, // 60/3 = 20 = max
		{90, 20}, // 90/3 = 30, clamped to 20
		{1000, 20},
	}
	for _, tt := range tests {
		got := defaultDiagramHeight(tt.w)
		if got != tt.want {
			t.Errorf("defaultDiagramHeight(%d) = %d, want %d", tt.w, got, tt.want)
		}
	}
}

func TestChartRenderNoDiagrams(t *testing.T) {
	g := New(WithWidth(40), WithNoColor())
	out := g.Render()
	// Empty chart should still produce at least a newline.
	if len(out) == 0 {
		t.Error("empty chart render produced no output")
	}
}

func TestChartRenderLines(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	lines := g.RenderLines(40)
	if len(lines) == 0 {
		t.Error("RenderLines returned empty slice")
	}
	for i, l := range lines {
		if len(l) > 0 && l[len(l)-1] == '\n' {
			t.Errorf("RenderLines line %d has trailing newline", i)
		}
	}
}

func TestChartReader(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(smallPlot())
	r := g.Reader()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Error("Reader returned empty content")
	}
}

// --- chartbase setter/resetter no-panic tests ---

func TestChartBaseResetTitle(t *testing.T) {
	p := smallPlot()
	p.SetTitle("hello")
	if p.title != "hello" {
		t.Fatal("SetTitle did not store title")
	}
	p.ResetTitle()
	if p.title != "" {
		t.Error("ResetTitle did not clear title")
	}
}

func TestChartBaseResetLabels(t *testing.T) {
	p := smallPlot()
	p.SetXLabel("x")
	p.SetYLabel("y")
	if p.xLabel != "x" || p.yLabel != "y" {
		t.Fatal("SetLabels did not store values")
	}
	p.ResetLabels()
	if p.xLabel != "" || p.yLabel != "" {
		t.Error("ResetLabels did not clear labels")
	}
}

func TestChartBaseSetXTicks(t *testing.T) {
	p := smallPlot()
	ticks := []Tick{{Label: "a", Value: 1}, {Label: "b", Value: 2}}
	p.SetXTicks(ticks)
	if len(p.xTicks) != 2 {
		t.Errorf("SetXTicks: got %d ticks, want 2", len(p.xTicks))
	}
	// nil ticks
	p.SetXTicks(nil)
	if p.xTicks != nil {
		t.Error("SetXTicks(nil) did not set nil")
	}
	// empty slice
	p.SetXTicks([]Tick{})
	if len(p.xTicks) != 0 {
		t.Error("SetXTicks([]) should produce empty slice")
	}
}

func TestChartBaseSetYTicks(t *testing.T) {
	p := smallPlot()
	ticks := []Tick{{Label: "0", Value: 0}, {Label: "5", Value: 5}}
	p.SetYTicks(ticks)
	if len(p.yTicks) != 2 {
		t.Errorf("SetYTicks: got %d ticks, want 2", len(p.yTicks))
	}
}

func TestChartBaseResetTicks(t *testing.T) {
	p := smallPlot()
	p.SetXTicks([]Tick{{Label: "a", Value: 1}})
	p.SetYTicks([]Tick{{Label: "b", Value: 2}})
	p.ResetTicks()
	if p.xTicks != nil || p.yTicks != nil {
		t.Error("ResetTicks did not clear ticks")
	}
}

func TestChartBaseResetFormatters(t *testing.T) {
	p := smallPlot()
	p.SetXFormatter(func(f float64) string { return "xfmt" })
	p.SetYFormatter(func(f float64) string { return "yfmt" })
	if p.xFmt == nil || p.yFmt == nil {
		t.Fatal("formatters not set")
	}
	p.ResetFormatters()
	if p.xFmt != nil || p.yFmt != nil {
		t.Error("ResetFormatters did not clear formatters")
	}
}

func TestChartBaseSetScaleAndReset(t *testing.T) {
	p := smallPlot()
	p.SetScale(0, 10, -5, 5)
	if p.x0 != 0 || p.x1 != 10 || p.y0 != -5 || p.y1 != 5 {
		t.Error("SetScale values mismatch")
	}
	if !p.xSet || !p.ySet {
		t.Error("SetScale did not set xSet/ySet")
	}
	p.ResetScale()
	if p.xSet || p.ySet {
		t.Error("ResetScale did not clear xSet/ySet")
	}
}

func TestChartBaseSetXRange(t *testing.T) {
	p := smallPlot()
	p.SetXRange(1, 100)
	if p.x0 != 1 || p.x1 != 100 || !p.xSet {
		t.Error("SetXRange values mismatch")
	}
}

func TestChartBaseResetSize(t *testing.T) {
	p := smallPlot()
	p.SetSize(15)
	if p.HeightHint(0) != 15 {
		t.Error("SetSize not reflected in HeightHint")
	}
	p.ResetSize()
	if p.height != 0 {
		t.Error("ResetSize did not clear height")
	}
}

func TestChartBaseSetGrid(t *testing.T) {
	p := smallPlot()
	p.SetGrid(false)
	if p.grid {
		t.Error("SetGrid(false) did not disable grid")
	}
	p.SetGrid(true)
	if !p.grid {
		t.Error("SetGrid(true) did not enable grid")
	}
}

func TestChartBaseSetBorder(t *testing.T) {
	p := smallPlot()
	p.SetBorder(false)
	if p.frame {
		t.Error("SetBorder(false) did not disable border")
	}
	p.SetBorder(true)
	if !p.frame {
		t.Error("SetBorder(true) did not enable border")
	}
}

func TestChartBaseSetLegend(t *testing.T) {
	p := smallPlot()
	p.SetLegend(false)
	if p.legend {
		t.Error("SetLegend(false) did not disable legend")
	}
	p.SetLegend(true)
	if !p.legend {
		t.Error("SetLegend(true) did not enable legend")
	}
}

func TestChartBaseSetTickCount(t *testing.T) {
	p := smallPlot()
	p.SetTickCount(10)
	if p.tickN != 10 {
		t.Errorf("SetTickCount(10): got %d", p.tickN)
	}
	// Minimum is 2.
	p.SetTickCount(1)
	if p.tickN != 2 {
		t.Errorf("SetTickCount(1): got %d, want 2", p.tickN)
	}
	p.SetTickCount(0)
	if p.tickN != 2 {
		t.Errorf("SetTickCount(0): got %d, want 2", p.tickN)
	}
	p.SetTickCount(-5)
	if p.tickN != 2 {
		t.Errorf("SetTickCount(-5): got %d, want 2", p.tickN)
	}
}

func TestChartBaseResetShowValues(t *testing.T) {
	p := smallPlot()
	p.SetShowValues(true)
	if !p.showVals {
		t.Error("SetShowValues(true) did not enable")
	}
	p.ResetShowValues()
	if p.showVals {
		t.Error("ResetShowValues did not disable")
	}
}

func TestChartBaseResetCellWidth(t *testing.T) {
	p := smallPlot()
	p.SetCellWidth(5)
	if p.cellW != 5 {
		t.Errorf("SetCellWidth(5): got %d", p.cellW)
	}
	p.ResetCellWidth()
	if p.cellW != 0 {
		t.Error("ResetCellWidth did not reset to 0")
	}
}

func TestChartBaseResetOrientation(t *testing.T) {
	p := smallPlot()
	p.SetOrientation(OrientHorizontal)
	if p.orient != OrientHorizontal {
		t.Error("SetOrientation did not store value")
	}
	p.ResetOrientation()
	if p.orient != OrientAuto {
		t.Error("ResetOrientation did not restore OrientAuto")
	}
}

func TestChartBaseSetTitleAlign(t *testing.T) {
	p := smallPlot()
	p.SetTitleAlign(AlignRight)
	if p.titleAlign != AlignRight {
		t.Error("SetTitleAlign did not store value")
	}
	p.SetTitleAlign(AlignCenter)
	if p.titleAlign != AlignCenter {
		t.Error("SetTitleAlign(AlignCenter) failed")
	}
}

func TestChartBaseResetTitleAlign(t *testing.T) {
	p := smallPlot()
	p.SetTitleAlign(AlignRight)
	p.ResetTitleAlign()
	if p.titleAlign != AlignLeft {
		t.Error("ResetTitleAlign did not restore AlignLeft")
	}
}

// --- chartbase rendering with base setters visible ---

func TestPlotTitleAlignVisible(t *testing.T) {
	for _, align := range []Align{AlignLeft, AlignCenter, AlignRight} {
		p := smallPlot()
		p.SetTitle("T")
		p.SetTitleAlign(align)
		p.SetBorder(false)
		g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
		g.Add(p)
		out := g.Render()
		if !strings.Contains(out, "T") {
			t.Errorf("align=%d: title 'T' not in output", align)
		}
	}
}

func TestPlotGridToggle(t *testing.T) {
	p := smallPlot()
	p.SetGrid(false)
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(p)
	out := g.Render()
	if out == "" {
		t.Error("empty output with grid off")
	}
}

func TestPlotBorderOff(t *testing.T) {
	p := smallPlot()
	p.SetBorder(false)
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(p)
	out := g.Render()
	if strings.Contains(out, "┌") || strings.Contains(out, "│") {
		t.Error("border symbols present with SetBorder(false)")
	}
}

func TestPlotLegendOff(t *testing.T) {
	p := smallPlot()
	p.SetLegend(false)
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(p)
	out := g.Render()
	if out == "" {
		t.Error("empty output with legend off")
	}
}

func TestPlotTicksOverride(t *testing.T) {
	p := smallPlot()
	p.SetXTicks([]Tick{{Label: "A", Value: 1}, {Label: "B", Value: 2}})
	p.SetYTicks([]Tick{{Label: "0", Value: 0}, {Label: "5", Value: 5}})
	g := New(WithWidth(40), WithNoColor(), WithUnicode(false))
	g.Add(p)
	out := g.Render()
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Error("custom X ticks not visible")
	}
}

// chartShortWriter is a writer that fails after max bytes.
type chartShortWriter struct {
	max int
	n   int
}

func (sw *chartShortWriter) Write(p []byte) (int, error) {
	remain := sw.max - sw.n
	if remain <= 0 {
		return 0, io.ErrShortWrite
	}
	if len(p) > remain {
		sw.n += remain
		return remain, io.ErrShortWrite
	}
	sw.n += len(p)
	return len(p), nil
}
