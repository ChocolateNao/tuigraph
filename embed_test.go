package tuichart

import (
	"strings"
	"testing"
)

func sampleChart(opts ...Option) *Chart {
	g := New(opts...)
	p := NewPlot()
	p.Add(NewLineVals("s", []float64{1, 3, 2, 5}))
	g.Add(p)
	return g
}

func TestRenderLinesMatchesRender(t *testing.T) {
	g := sampleChart(WithNoColor(), WithUnicode(true))
	joined := strings.Join(g.RenderLines(60), "\n") + "\n"
	if joined != g.Render(60) {
		t.Error("RenderLines and Render disagree")
	}
	for i, l := range g.RenderLines(40) {
		if strings.Contains(l, "\n") {
			t.Errorf("line %d contains newline", i)
		}
	}
}

func TestEachCellCoversGrid(t *testing.T) {
	cv := NewCanvas(5, 3)
	n := 0
	cv.EachCell(func(x, y int, cl Cell) {
		if x < 0 || x >= 5 || y < 0 || y >= 3 {
			t.Errorf("out of range %d,%d", x, y)
		}
		if cl.Ch != ' ' || !cl.Fg.IsZero() {
			t.Errorf("expected blank, got %+v", cl)
		}
		n++
	})
	if n != 15 {
		t.Errorf("visited %d cells", n)
	}
}

func TestCellAtCarriesStyle(t *testing.T) {
	cv := NewCanvas(10, 1)
	cv.Text(0, 0, "hi", S(RGB(1, 2, 3)).On(RGB(4, 5, 6)).Bolder())
	cl := cv.CellAt(0, 0)
	if cl.Ch != 'h' || !cl.Bold {
		t.Fatalf("cell = %+v", cl)
	}
	r, g, b, ok := cl.Fg.RGB()
	if !ok || r != 1 || g != 2 || b != 3 {
		t.Errorf("fg = %d,%d,%d ok=%v", r, g, b, ok)
	}
}

func TestColorRGBIndexed(t *testing.T) {
	if _, _, _, ok := Default.RGB(); ok {
		t.Error("Default should not resolve")
	}
	r, g, b, _ := Red.RGB()
	if r != 255 || g != 0 || b != 0 {
		t.Errorf("Red -> %d,%d,%d", r, g, b)
	}
	r, g, b, _ = Indexed(16).RGB()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("idx16 -> %d,%d,%d", r, g, b)
	}
	r, _, _, _ = Indexed(231).RGB()
	if r != 255 {
		t.Errorf("idx231 r=%d", r)
	}
	r, _, _, _ = Indexed(255).RGB()
	if r != 238 {
		t.Errorf("idx255 r=%d", r)
	}
}
