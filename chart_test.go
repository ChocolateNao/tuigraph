package tuichart

import (
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
