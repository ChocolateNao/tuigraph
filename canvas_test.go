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
