package tuichart

import "strings"

// Rect represents a rectangular region on the canvas.
type Rect struct{ X, Y, W, H int }

// X2 returns the rightmost X coordinate (X + W - 1).
func (r Rect) X2() int { return r.X + r.W - 1 }

// Y2 returns the bottommost Y coordinate (Y + H - 1).
func (r Rect) Y2() int { return r.Y + r.H - 1 }

// Contains reports whether the point (x, y) lies inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type cell struct {
	ch   rune
	fg   Color
	bg   Color
	bold bool
}

var blankCell = cell{ch: ' '}

func (c cell) blank() bool {
	return c.ch == ' ' && c.fg.IsZero() && c.bg.IsZero() && !c.bold
}

// Canvas is a 2-D cell buffer used for drawing text and graphics.
type Canvas struct {
	buf    []cell
	w      int
	h      int
	stride int
	ox     int
	oy     int
}

// NewCanvas creates a blank canvas of the given width and height.
func NewCanvas(w, h int) *Canvas {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	c := &Canvas{w: w, h: h, stride: w, buf: make([]cell, w*h)}
	c.Clear()
	return c
}

// Clear resets every cell in the canvas to a blank space.
func (c *Canvas) Clear() {
	for i := range c.buf {
		c.buf[i] = blankCell
	}
}

// Width returns the canvas width in cells.
func (c *Canvas) Width() int { return c.w }

// Height returns the canvas height in cells.
func (c *Canvas) Height() int { return c.h }

// Rect returns the bounding rectangle of the canvas.
func (c *Canvas) Rect() Rect { return Rect{X: 0, Y: 0, W: c.w, H: c.h} }

// Set writes a single cell at position (x, y) with the given rune and style.
func (c *Canvas) Set(x, y int, ch rune, st Style) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	px, py := c.ox+x, c.oy+y
	if px < 0 || py < 0 || px >= c.stride || py*c.stride+px >= len(c.buf) {
		return
	}
	i := py*c.stride + px
	cell := &c.buf[i]
	cell.ch = ch
	cell.fg = st.Fg
	cell.bg = st.Bg
	cell.bold = st.Bold
}

// At returns the internal cell at position (x, y), or a blank cell if
// out of bounds.
func (c *Canvas) At(x, y int) cell {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return blankCell
	}
	return c.buf[(c.oy+y)*c.stride+c.ox+x]
}

// Cell is the public snapshot of a canvas cell, for embedding charts into
// external screen buffers (tcell, tview primitives, custom renderers).
type Cell struct {
	Ch   rune
	Fg   Color
	Bg   Color
	Bold bool
}

// CellAt returns the cell at canvas coordinates; out-of-bounds reads yield
// a blank cell.
func (c *Canvas) CellAt(x, y int) Cell {
	cl := c.At(x, y)
	return Cell{Ch: cl.ch, Fg: cl.fg, Bg: cl.bg, Bold: cl.bold}
}

// EachCell visits every cell in row-major order. Coordinates are relative
// to this canvas view (Sub views included).
func (c *Canvas) EachCell(fn func(x, y int, cl Cell)) {
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			fn(x, y, c.CellAt(x, y))
		}
	}
}

// Text draws the string s starting at position (x, y) and returns the
// number of cells written.
func (c *Canvas) Text(x, y int, s string, st Style) int {
	cx := x
	for _, r := range s {
		if cx >= c.w {
			break
		}
		c.Set(cx, y, r, st)
		cx++
	}
	return cx - x
}

// TextCenter draws the string s horizontally centered on column cx at row y.
func (c *Canvas) TextCenter(cx, y int, s string, st Style) int {
	n := runeLen(s)
	x := cx - n/2
	if x < 0 {
		x = 0
	}
	return c.Text(x, y, s, st)
}

// TextRight draws the string s right-aligned so that it ends at column x2.
func (c *Canvas) TextRight(x2, y int, s string, st Style) int {
	x := x2 - runeLen(s) + 1
	if x < 0 {
		s = truncStr(s, x2+1)
		x = 0
	}
	return c.Text(x, y, s, st)
}

// FillRect fills the rectangle r with the given rune and style.
func (c *Canvas) FillRect(r Rect, ch rune, st Style) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			c.Set(x, y, ch, st)
		}
	}
}

// HLine draws a horizontal line from (x1, y) to (x2, y).
func (c *Canvas) HLine(y, x1, x2 int, ch rune, st Style) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		c.Set(x, y, ch, st)
	}
}

// VLine draws a vertical line from (x, y1) to (x, y2).
func (c *Canvas) VLine(x, y1, y2 int, ch rune, st Style) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		c.Set(x, y, ch, st)
	}
}

func boxRunes(unicode bool) (lt, rt, lb, rb, hz, vt rune) {
	if unicode {
		return '┌', '┐', '└', '┘', '─', '│'
	}
	return '+', '+', '+', '+', '-', '|'
}

// Border draws a single-line box around the entire canvas.
func (c *Canvas) Border(st Style, unicode bool) {
	lt, rt, lb, rb, hz, vt := boxRunes(unicode)
	x2, y2 := c.w-1, c.h-1
	c.Set(0, 0, lt, st)
	c.Set(x2, 0, rt, st)
	c.Set(0, y2, lb, st)
	c.Set(x2, y2, rb, st)
	c.HLine(0, 1, x2-1, hz, st)
	c.HLine(y2, 1, x2-1, hz, st)
	c.VLine(0, 1, y2-1, vt, st)
	c.VLine(x2, 1, y2-1, vt, st)
}

// Sub returns a view of the canvas clipped to rectangle r, sharing the
// same underlying buffer.
func (c *Canvas) Sub(r Rect) *Canvas {
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return &Canvas{
		w: r.W, h: r.H,
		stride: c.stride,
		buf:    c.buf,
		ox:     c.ox + r.X,
		oy:     c.oy + r.Y,
	}
}

// Blit copies non-blank cells from src onto this canvas at offset (ox, oy).
func (c *Canvas) Blit(src *Canvas, ox, oy int) {
	for y := 0; y < src.h; y++ {
		for x := 0; x < src.w; x++ {
			sc := src.At(x, y)
			if sc.blank() {
				continue
			}
			c.Set(ox+x, oy+y, sc.ch, Style{Fg: sc.fg, Bg: sc.bg, Bold: sc.bold})
		}
	}
}

// String renders the canvas honoring the current color profile.
func (c *Canvas) String() string {
	return c.Render(Detect().Level)
}

// Plain renders the canvas without any ANSI escapes.
func (c *Canvas) Plain() string {
	return c.Render(LevelNone)
}

// Render produces the canvas as a string using ANSI escape sequences
// appropriate for the given color level.
func (c *Canvas) Render(lvl Level) string {
	var b strings.Builder
	for y := 0; y < c.h; y++ {
		last := -1
		for x := c.w - 1; x >= 0; x-- {
			if !c.At(x, y).blank() {
				last = x
				break
			}
		}
		if last < 0 {
			b.WriteString("\n")
			continue
		}
		cur := Style{}
		active := false
		for x := 0; x <= last; x++ {
			cl := c.At(x, y)
			st := Style{Fg: cl.fg, Bg: cl.bg, Bold: cl.bold}
			if !st.eq(cur) {
				if active {
					b.WriteString(ansiReset)
					active = false
				}
				seq := lvl.seq(st)
				if seq != "" {
					b.WriteString(seq)
					active = true
					cur = st
				} else {
					cur = st
				}
			}
			b.WriteRune(cl.ch)
		}
		if active {
			b.WriteString(ansiReset)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func runeLen(s string) int { return len([]rune(s)) }

func truncStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		if n <= 0 {
			return ""
		}
		return "…"
	}
	return string(r[:n-1]) + "…"
}

// ellipTrunc truncates to n cells using "…" in unicode mode and plain
// "..." otherwise, so non-unicode terminals never receive multi-byte
// glyphs they cannot render.
func ellipTrunc(s string, n int, uni bool) string {
	if uni {
		return truncStr(s, n)
	}
	return truncASCII(s, n)
}

func truncASCII(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:maxInt(n, 0)]
	}
	return s[:n-3] + "..."
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
