package tuichart

import (
	"math"
	"strings"
	"testing"
)

func TestSparklineUnicodeBlocks(t *testing.T) {
	cv := NewCanvas(20, 1)
	s := NewSpark(1, 2, 3, 2, 1)
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	out := cv.Plain()
	if !strings.ContainsAny(out, "▁▂▃▄▅▆▇█") {
		t.Errorf("unicode sparkline got %q", out)
	}
}

func TestSparklineASCIIFallback(t *testing.T) {
	cv := NewCanvas(20, 1)
	s := NewSpark(1, 5, 9, 4, 0)
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: false}), cv)
	out := strings.TrimRight(cv.Plain(), "\n")
	if strings.ContainsAny(out, "▁▂▃▄▅▆▇█") {
		t.Errorf("ascii fallback leaked blocks: %q", out)
	}
	if len([]rune(out)) != len(s.vals) {
		t.Errorf("cell count %d != %d", len([]rune(out)), len(s.vals))
	}
}

func TestSparklineEmpty(t *testing.T) {
	cv := NewCanvas(10, 3)
	NewSpark().Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
}

func TestSparkString(t *testing.T) {
	out := Spark([]float64{0, 1, 2, 1, 0})
	if runeLen(out) != 5 {
		t.Errorf("Spark len %d", runeLen(out))
	}
}

func TestHeatmapMonoRamp(t *testing.T) {
	grid := [][]float64{{0, 25}, {50, 100}}
	h := NewHeat(grid).RowLabels("r")
	cv := NewCanvas(30, 6)
	h.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	out := cv.Plain()
	distinct := map[rune]bool{}
	for _, r := range out {
		if strings.ContainsRune(string(heatRampASCII), r) {
			distinct[r] = true
		}
	}
	if len(distinct) < 2 {
		t.Errorf("mono heatmap shows <2 ramp levels: %q", out)
	}
	if !strings.Contains(out, "r") {
		t.Error("row label missing")
	}
}

func TestPieLegendAndSlices(t *testing.T) {
	p := NewPie().Slice("alpha", 50).Slice("beta", 30).Slice("gamma", 20)
	g := New(WithWidth(60), WithNoColor(), WithUnicode(true))
	g.Add(p)
	out := g.Render()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(out, name) {
			t.Errorf("legend missing %q", name)
			continue
		}
		i := strings.Index(out, name)
		if i+1 >= len(out) || !strings.Contains(out[i:i+len(name)+4], "%") {
			t.Errorf("%s has no percent label", name)
		}
	}
	chars := []rune(pieASCIIChars)
	found := map[rune]int{}
	for _, r := range out {
		for _, c := range chars {
			if r == c {
				found[c]++
			}
		}
	}
	if len(found) < 2 {
		t.Error("mono pie slices indistinguishable")
	}
}

func TestBarNegativeValues(t *testing.T) {
	b := NewBarValues([]string{"neg", "pos"}, []float64{-5, 8})
	g := New(WithWidth(40), WithNoColor())
	g.Add(b)
	out := g.Render()
	if out == "" {
		t.Fatal("empty render")
	}
}

// ── sparkline.go ────────────────────────────────────────────────────────────

func TestSparklineSetValuesReturnsReceiver(t *testing.T) {
	s := NewSpark(1, 2, 3)
	ret := s.SetValues([]float64{9, 9, 9, 9})
	if ret != s {
		t.Error("SetValues did not return receiver")
	}
	if len(s.vals) != 4 {
		t.Fatalf("vals len = %d, want 4", len(s.vals))
	}
	for _, v := range s.vals {
		if v != 9 {
			t.Errorf("SetValues did not replace value, got %v", v)
		}
	}
}

func TestSparklineSetValuesReplacesRender(t *testing.T) {
	s := NewSpark(0, 4, 8)
	before := renderD(s, WithWidth(20))
	s.SetValues([]float64{1, 2})
	out := renderD(s, WithWidth(20))
	if out == before {
		t.Error("SetValues did not change render")
	}
	if len(s.vals) != 2 {
		t.Errorf("vals len = %d, want 2 (replacement, not append)", len(s.vals))
	}
}

func TestSparklineSetValuesAllSame(t *testing.T) {
	cv := NewCanvas(10, 1)
	s := NewSpark(1, 2, 3)
	s.SetValues([]float64{5, 5, 5})
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	for _, r := range strings.TrimRight(cv.Plain(), "\n") {
		if r != ' ' {
			t.Errorf("all-same after SetValues got rune %q, want space", r)
		}
	}
}

func TestSparklineValuesReturnsSelf(t *testing.T) {
	s := NewSpark(1, 2)
	ret := s.Values(3, 4)
	if ret != s {
		t.Fatal("Values did not return receiver")
	}
	if len(s.vals) != 4 {
		t.Errorf("vals len = %d, want 4", len(s.vals))
	}
}

func TestSparklineValuesEmpty(t *testing.T) {
	s := NewSpark()
	s.Values()
	if len(s.vals) != 0 {
		t.Errorf("Values() with no args changed len to %d", len(s.vals))
	}
}

func TestSparklineColorReturnsSelf(t *testing.T) {
	s := NewSpark(1, 2, 3)
	ret := s.Color(Red)
	if ret != s {
		t.Fatal("Color did not return receiver")
	}
	if s.color != Red {
		t.Error("color not set")
	}
}

func TestSparklineAllSameValues(t *testing.T) {
	cv := NewCanvas(10, 1)
	s := NewSpark(5, 5, 5, 5)
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	// All same values → spaces (hi == lo path)
	out := cv.Plain()
	for _, r := range strings.TrimRight(out, "\n") {
		if r != ' ' {
			t.Errorf("all-same values got rune %q, want space", r)
		}
	}
}

func TestSparklineWithNaN(t *testing.T) {
	cv := NewCanvas(10, 1)
	s := NewSpark(1, math.NaN(), 3, math.NaN(), 5)
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	out := cv.Plain()
	// NaN entries should be spaces, non-NaN should have block chars
	if strings.TrimSpace(out) == "" {
		t.Error("all-NaN output")
	}
}

func TestSparklineWithInf(t *testing.T) {
	cv := NewCanvas(10, 1)
	s := NewSpark(math.Inf(1), 2, 3)
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	if cv.Plain() == "" {
		t.Error("empty render with Inf input")
	}
}

func TestSparklineTitle(t *testing.T) {
	s := NewSpark(1, 2, 3).Title("sp")
	if s.HeightHint(40) != 2 {
		t.Error("titled sparkline hint should be 2")
	}
}

func TestSparklineHeightHintNoTitle(t *testing.T) {
	s := NewSpark(1, 2, 3)
	if s.HeightHint(40) != 1 {
		t.Error("untitled sparkline hint should be 1")
	}
}

func TestSparklineRenderNoColor(t *testing.T) {
	s := NewSpark(1, 2, 3, 4, 5).Color(Red)
	out := renderD(s, WithWidth(20))
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestSparklineMoreValuesThanWidth(t *testing.T) {
	vals := make([]float64, 100)
	for i := range vals {
		vals[i] = float64(i)
	}
	s := NewSpark(vals...)
	cv := NewCanvas(5, 1)
	s.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: false}), cv)
	// Only first 5 runes should appear (canvas width)
	runeCount := len(strings.TrimRight(cv.Plain(), "\n"))
	if runeCount > 5 {
		t.Errorf("output %d runes exceeds canvas width 5", runeCount)
	}
}

func TestHeatmapTitleReturnsReceiver(t *testing.T) {
	h := NewHeat([][]float64{{1}})
	ret := h.Title("T")
	if ret != h {
		t.Error("Title did not return receiver")
	}
}

func TestHeatmapColorsReturnsReceiver(t *testing.T) {
	h := NewHeat([][]float64{{1}})
	ret := h.Colors(Red, Blue)
	if ret != h {
		t.Error("Colors did not return receiver")
	}
}

func TestHeatmapOneByOne(t *testing.T) {
	h := NewHeat([][]float64{{42}})
	out := renderD(h, WithWidth(40))
	if strings.Contains(out, "(no data)") {
		t.Errorf("1x1 heatmap hit no-data guard:\n%s", out)
	}
}

func TestHeatmapNonSquareWide(t *testing.T) {
	h := NewHeat([][]float64{{1, 2, 3, 4}})
	out := renderD(h, WithWidth(50))
	if strings.Contains(out, "(no data)") {
		t.Errorf("1x4 heatmap empty:\n%s", out)
	}
}

func TestHeatmapNonSquareTall(t *testing.T) {
	h := NewHeat([][]float64{{1}, {2}, {3}, {4}})
	out := renderD(h, WithWidth(30))
	if strings.Contains(out, "(no data)") {
		t.Errorf("4x1 heatmap empty:\n%s", out)
	}
}

func TestHeatmapEmptyGrid(t *testing.T) {
	out := renderD(NewHeat(nil), WithWidth(40))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("nil grid should show no data:\n%s", out)
	}
}

func TestHeatmapAllNaN(t *testing.T) {
	out := renderD(NewHeat([][]float64{{NaN(), NaN()}}), WithWidth(40))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("all-NaN grid should show no data:\n%s", out)
	}
}

func TestHeatmapConstantValue(t *testing.T) {
	h := NewHeat([][]float64{{5, 5}, {5, 5}})
	out := renderD(h, WithWidth(40))
	if len(out) == 0 {
		t.Error("constant-value heatmap empty")
	}
}

func TestHeatmapWithValues(t *testing.T) {
	h := NewHeat([][]float64{{1, 2}, {3, 4}}).ShowValues(true)
	out := renderD(h, WithWidth(50))
	hasDigit := false
	for _, r := range out {
		if r >= '0' && r <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		t.Errorf("ShowValues produced no digits:\n%s", out)
	}
}

func TestHeatmapCustomColors(t *testing.T) {
	base := renderD(NewHeat([][]float64{{0, 100}}), WithColor256(), WithWidth(40))
	custom := renderD(
		NewHeat([][]float64{{0, 100}}).Colors(White, Black),
		WithColor256(),
		WithWidth(40),
	)
	if base == custom {
		t.Error("Colors setter had no effect")
	}
}

func TestHeatmapHeightHintColLabels(t *testing.T) {
	h := NewHeat([][]float64{{1, 2}}).ColLabels("a", "b")
	hh := h.HeightHint(40)
	if hh < 5 {
		t.Errorf("HeightHint with colLabels = %d, want >=5", hh)
	}
}

func TestHeatmapFluentChain(t *testing.T) {
	h := NewHeat([][]float64{{1}})
	ret := h.Title("t").Colors(Red, Blue).RowLabels("r").ColLabels("c").
		CellWidth(3).ShowValues(true)
	if ret != h {
		t.Error("fluent chain broke")
	}
}

func TestHeatmapTinySize(t *testing.T) {
	h := NewHeat([][]float64{{1, 2}, {3, 4}})
	out := renderD(h, WithWidth(12))
	if len(out) == 0 {
		t.Error("tiny heatmap empty")
	}
}

func TestHeatmapRowLabelsOnlySome(t *testing.T) {
	h := NewHeat([][]float64{{1, 2}, {3, 4}, {5, 6}}).
		RowLabels("r0", "r1")
	out := renderD(h, WithWidth(50))
	if strings.Contains(out, "(no data)") {
		t.Errorf("partial labels heatmap empty:\n%s", out)
	}
}

func TestPieSliceColorReturnsReceiver(t *testing.T) {
	p := NewPie().Slice("a", 1)
	ret := p.SliceColor(Red)
	if ret != p {
		t.Error("SliceColor did not return receiver")
	}
}

func TestPieTitleReturnsReceiver(t *testing.T) {
	p := NewPie()
	ret := p.Title("T")
	if ret != p {
		t.Error("Title did not return receiver")
	}
}

func TestPieSliceColorEmpty(t *testing.T) {
	p := NewPie()
	ret := p.SliceColor(Red)
	if ret != p {
		t.Error("SliceColor on empty pie did not return receiver")
	}
}

func TestPieSingleSlice(t *testing.T) {
	p := NewPie().Slice("only", 100)
	out := renderD(p, WithNoColor(), WithUnicode(false), WithWidth(40))
	if !strings.Contains(out, "only") {
		t.Errorf("single slice missing:\n%s", out)
	}
	if !strings.Contains(out, "100%") {
		t.Errorf("single slice should be 100%%:\n%s", out)
	}
}

func TestPieManySlices(t *testing.T) {
	p := NewPie()
	for i := 0; i < 10; i++ {
		name := string(rune('A' + i))
		p.Slice(name, float64(i+1))
	}
	out := renderD(p, WithNoColor(), WithUnicode(false), WithWidth(70))
	for i := 0; i < 10; i++ {
		name := string(rune('A' + i))
		if !strings.Contains(out, name) {
			t.Errorf("slice %q missing:\n%s", name, out)
		}
	}
}

func TestPieShowValuesEdge(t *testing.T) {
	p := NewPie().Slice("x", 42).Slice("y", 58).ShowValues(true)
	out := renderD(p, WithNoColor(), WithUnicode(false), WithWidth(60))
	if !strings.Contains(out, "42") {
		t.Errorf("ShowValues missing raw value:\n%s", out)
	}
}

func TestPieDonut(t *testing.T) {
	p := NewPie().Slice("a", 50).Slice("b", 50).Donut(true)
	out := renderD(p, WithNoColor(), WithUnicode(true), WithWidth(50))
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("donut legend missing:\n%s", out)
	}
}

func TestPieZeroValues(t *testing.T) {
	p := NewPie().Slice("zero", 0).Slice("also", 0)
	out := renderD(p, WithWidth(40))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("all-zero pie should show no data:\n%s", out)
	}
}

func TestPieNegativeValues(t *testing.T) {
	p := NewPie().Slice("neg", -10).Slice("pos", 10)
	out := renderD(p, WithNoColor(), WithUnicode(false), WithWidth(50))
	if !strings.Contains(out, "pos") {
		t.Errorf("positive slice missing:\n%s", out)
	}
}

func TestPieHeightHintBounds(t *testing.T) {
	p := NewPie()
	h := p.HeightHint(20)
	if h < 8 {
		t.Errorf("HeightHint too low: %d", h)
	}
	h2 := p.HeightHint(200)
	if h2 > 22 {
		t.Errorf("HeightHint too high: %d", h2)
	}
}

func TestPieFluentChain(t *testing.T) {
	p := NewPie()
	ret := p.Title("t").Slice("a", 1).SliceColor(Red).Slice("b", 2).
		Donut(true).ShowValues(true)
	if ret != p {
		t.Error("fluent chain broke")
	}
}

func TestPieTinySize(t *testing.T) {
	p := NewPie().Slice("a", 1).Slice("b", 1)
	out := renderD(p, WithWidth(12))
	if len(out) == 0 {
		t.Error("tiny pie empty")
	}
}
