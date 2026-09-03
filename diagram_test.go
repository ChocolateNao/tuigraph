package tuichart

import (
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
