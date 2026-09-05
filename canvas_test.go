package tuichart

import (
	"strings"
	"testing"
)

func TestCanvasBounds(t *testing.T) {
	cv := NewCanvas(5, 3)
	cv.Set(-1, 0, 'x', Style{})
	cv.Set(5, 0, 'x', Style{})
	cv.Set(0, -1, 'x', Style{})
	cv.Set(0, 3, 'x', Style{})
	if got := cv.Plain(); strings.ContainsAny(got, "x") {
		t.Errorf("out-of-bounds writes leaked: %q", got)
	}
}

func TestCanvasTextAndRender(t *testing.T) {
	cv := NewCanvas(6, 2)
	cv.Text(1, 0, "hello", Style{})
	want := " hello\n\n"
	if got := cv.Plain(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestCanvasSub(t *testing.T) {
	root := NewCanvas(10, 5)
	sub := root.Sub(Rect{X: 2, Y: 1, W: 4, H: 2})
	sub.Set(0, 0, 'A', Style{})
	sub.Set(3, 1, 'B', Style{})
	if root.At(2, 1).ch != 'A' || root.At(5, 2).ch != 'B' {
		t.Fatal("sub view offset wrong")
	}
	sub.Set(10, 0, 'X', Style{})
	if strings.Contains(root.Plain(), "X") {
		t.Error("sub write escaped root bounds")
	}
}

func TestCanvasBlitSkipsBlank(t *testing.T) {
	dst := NewCanvas(4, 2)
	dst.Text(0, 0, "abcd", Style{})
	src := NewCanvas(2, 1)
	src.Set(1, 0, 'Z', Style{})
	dst.Blit(src, 1, 0)
	if dst.At(1, 0).ch != 'b' {
		t.Errorf("blank src cell overwrote dst: %q", dst.Plain())
	}
	if dst.At(2, 0).ch != 'Z' {
		t.Error("non-blank cell not blitted")
	}
}

func TestCanvasBorderUnicode(t *testing.T) {
	cv := NewCanvas(4, 3)
	cv.Border(Style{}, true)
	lines := splitLines(cv.Plain())
	if !strings.HasPrefix(lines[0], "┌") || !strings.HasSuffix(lines[0], "┐") {
		t.Errorf("unicode border wrong: %q", lines[0])
	}
}

func TestTruncStr(t *testing.T) {
	if s := truncStr("abcdef", 4); runeLen(s) != 4 {
		t.Errorf("truncStr len %d", runeLen(s))
	}
	if truncStr("ab", 5) != "ab" {
		t.Error("no-op trunc failed")
	}
}

// ── canvas.go ───────────────────────────────────────────────────────────────

func TestRectContains(t *testing.T) {
	r := Rect{X: 2, Y: 3, W: 5, H: 4}
	tests := []struct {
		x, y int
		want bool
	}{
		{2, 3, true},  // top-left corner (inside)
		{6, 6, true},  // bottom-right inner
		{7, 3, false}, // just outside right edge (X+W-1=6, so x=7 is outside)
		{2, 7, false}, // just outside bottom edge (Y+H-1=6, so y=7 is outside)
		{1, 3, false}, // left of rect
		{2, 2, false}, // above rect
		{0, 0, false}, // far outside
	}
	for _, tc := range tests {
		if got := r.Contains(tc.x, tc.y); got != tc.want {
			t.Errorf("Rect{%v}.Contains(%d,%d) = %v, want %v", r, tc.x, tc.y, got, tc.want)
		}
	}
}

func TestRectContainsZeroSize(t *testing.T) {
	r := Rect{X: 5, Y: 5, W: 0, H: 0}
	if r.Contains(5, 5) {
		t.Error("zero-size rect should not contain any point")
	}
}

func TestRectX2Y2(t *testing.T) {
	r := Rect{X: 1, Y: 2, W: 3, H: 4}
	if r.X2() != 3 {
		t.Errorf("X2 = %d, want 3", r.X2())
	}
	if r.Y2() != 5 {
		t.Errorf("Y2 = %d, want 5", r.Y2())
	}
}

func TestRectClipContained(t *testing.T) {
	r := (Rect{X: 1, Y: 1, W: 5, H: 5}).clip(Rect{X: 0, Y: 0, W: 10, H: 10})
	if r != (Rect{X: 1, Y: 1, W: 5, H: 5}) {
		t.Errorf("contained clip = %+v", r)
	}
}

func TestRectClipOffLeftTop(t *testing.T) {
	r := (Rect{X: -2, Y: -2, W: 5, H: 5}).clip(Rect{X: 0, Y: 0, W: 10, H: 10})
	want := Rect{X: 0, Y: 0, W: 3, H: 3}
	if r != want {
		t.Errorf("left/top clip = %+v, want %+v", r, want)
	}
}

func TestRectClipOverflow(t *testing.T) {
	r := (Rect{X: 8, Y: 8, W: 5, H: 5}).clip(Rect{X: 0, Y: 0, W: 10, H: 10})
	if r.W != 2 || r.H != 2 {
		t.Errorf("overflow clip = %+v, want W=2 H=2", r)
	}
}

func TestRectClipOutsideZero(t *testing.T) {
	r := (Rect{X: 20, Y: 20, W: 5, H: 5}).clip(Rect{X: 0, Y: 0, W: 10, H: 10})
	if r.W != 0 || r.H != 0 {
		t.Errorf("fully outside clip = %+v, want zero size", r)
	}
}

func TestFillRectInside(t *testing.T) {
	cv := NewCanvas(6, 4)
	cv.FillRect(Rect{X: 1, Y: 1, W: 3, H: 2}, '#', S(Red))
	for y := 0; y < 4; y++ {
		for x := 0; x < 6; x++ {
			ch := cv.At(x, y).ch
			inR := x >= 1 && x < 4 && y >= 1 && y < 3
			if inR && ch != '#' {
				t.Errorf("(%d,%d) expected '#', got %q", x, y, ch)
			}
			if !inR && ch != ' ' {
				t.Errorf("(%d,%d) expected ' ', got %q", x, y, ch)
			}
		}
	}
}

func TestFillRectClipped(t *testing.T) {
	cv := NewCanvas(4, 3)
	// Rect extends beyond canvas bounds — Set clips silently
	cv.FillRect(Rect{X: 2, Y: 1, W: 10, H: 10}, 'X', Style{})
	if cv.At(3, 2).ch != 'X' {
		t.Error("clipped fill missing expected cell")
	}
	// Should not panic
}

func TestFillRectEmpty(t *testing.T) {
	cv := NewCanvas(4, 3)
	cv.FillRect(Rect{X: 10, Y: 10, W: 2, H: 2}, 'Z', Style{})
	// Entirely outside — no cells changed
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			if cv.At(x, y).ch != ' ' {
				t.Errorf("unexpected fill at (%d,%d)", x, y)
			}
		}
	}
}

func TestCanvasString(t *testing.T) {
	cv := NewCanvas(3, 1)
	cv.Set(0, 0, 'A', Style{})
	s := cv.String()
	if s == "" {
		t.Fatal("String() returned empty")
	}
}

func TestCanvasRenderAllBlank(t *testing.T) {
	cv := NewCanvas(4, 2)
	got := cv.Plain()
	// Two blank lines
	if got != "\n\n" {
		t.Errorf("blank canvas render = %q, want %q", got, "\n\n")
	}
}

func TestCanvasNewClamp(t *testing.T) {
	cv := NewCanvas(0, 0)
	if cv.Width() != 1 || cv.Height() != 1 {
		t.Errorf("zero dimensions clamped to %dx%d", cv.Width(), cv.Height())
	}
}

func TestCanvasCellAtBounds(t *testing.T) {
	cv := NewCanvas(2, 2)
	cl := cv.CellAt(-1, -1)
	if cl.Ch != ' ' {
		t.Error("out-of-bounds CellAt should return blank")
	}
	cl = cv.CellAt(0, 0)
	if cl.Ch != ' ' {
		t.Error("fresh canvas cell should be blank")
	}
}

func TestCanvasEachCell(t *testing.T) {
	cv := NewCanvas(3, 2)
	count := 0
	cv.EachCell(func(x, y int, cl Cell) {
		count++
	})
	if count != 6 {
		t.Errorf("EachCell visited %d cells, want 6", count)
	}
}

func TestCanvasTextRight(t *testing.T) {
	cv := NewCanvas(10, 1)
	n := cv.TextRight(9, 0, "hi", Style{})
	if n != 2 {
		t.Errorf("TextRight wrote %d chars", n)
	}
	if cv.At(8, 0).ch != 'h' {
		t.Errorf("expected 'h' at col 8, got %q", cv.At(8, 0).ch)
	}
}

func TestCanvasTextRightTruncate(t *testing.T) {
	cv := NewCanvas(3, 1)
	n := cv.TextRight(2, 0, "hello", Style{})
	if n != 3 {
		t.Errorf("truncated TextRight wrote %d chars", n)
	}
}

func TestCanvasHLineVLine(t *testing.T) {
	cv := NewCanvas(5, 5)
	cv.HLine(2, 1, 3, '-', Style{})
	for x := 1; x <= 3; x++ {
		if cv.At(x, 2).ch != '-' {
			t.Errorf("HLine missing at (%d,2)", x)
		}
	}
	cv.VLine(2, 1, 3, '|', Style{})
	for y := 1; y <= 3; y++ {
		if cv.At(2, y).ch != '|' {
			t.Errorf("VLine missing at (2,%d)", y)
		}
	}
}

func TestCanvasHLineReversed(t *testing.T) {
	cv := NewCanvas(5, 1)
	cv.HLine(0, 4, 1, 'R', Style{})
	for x := 1; x <= 4; x++ {
		if cv.At(x, 0).ch != 'R' {
			t.Errorf("reversed HLine missing at %d", x)
		}
	}
}

func TestCanvasBorderASCII(t *testing.T) {
	cv := NewCanvas(4, 3)
	cv.Border(Style{}, false)
	lines := splitLines(cv.Plain())
	if !strings.HasPrefix(lines[0], "+") || !strings.HasSuffix(lines[0], "+") {
		t.Errorf("ascii border wrong: %q", lines[0])
	}
}

func TestCanvasSubNegativeDims(t *testing.T) {
	cv := NewCanvas(4, 4)
	sub := cv.Sub(Rect{X: 1, Y: 1, W: -1, H: -1})
	if sub.Width() != 0 || sub.Height() != 0 {
		t.Errorf("negative dims clamped to %dx%d", sub.Width(), sub.Height())
	}
}

func TestCanvasClear(t *testing.T) {
	cv := NewCanvas(3, 1)
	cv.Set(1, 0, 'X', S(Red))
	cv.Clear()
	if cv.At(1, 0).ch != ' ' {
		t.Error("Clear did not reset cells")
	}
}
