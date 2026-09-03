package tuichart

import (
	"math"
	"strconv"
	"strings"
)

type Level int

const (
	LevelNone Level = iota
	Level16
	Level256
	LevelTrue
)

type colorKind uint8

const (
	colorNone colorKind = iota
	colorIndexed
	colorRGB
)

type Color struct {
	kind colorKind
	idx  uint8
	r    uint8
	g    uint8
	b    uint8
}

var (
	Default = Color{}
	Black   = Indexed(0)
	Maroon  = Indexed(1)
	Green   = Indexed(2)
	Olive   = Indexed(3)
	Navy    = Indexed(4)
	Purple  = Indexed(5)
	Teal    = Indexed(6)
	Silver  = Indexed(7)
	Gray    = Indexed(8)
	Red     = Indexed(9)
	Lime    = Indexed(10)
	Yellow  = Indexed(11)
	Blue    = Indexed(12)
	Fuchsia = Indexed(13)
	Cyan    = Indexed(14)
	White   = Indexed(15)
)

func Indexed(i int) Color {
	if i < 0 {
		i = 0
	}
	if i > 255 {
		i = 255
	}
	return Color{kind: colorIndexed, idx: uint8(i)}
}

func RGB(r, g, b uint8) Color {
	return Color{kind: colorRGB, r: r, g: g, b: b}
}

func (c Color) IsZero() bool { return c.kind == colorNone }

// RGB resolves the color to concrete red/green/blue components so host
// frameworks (tcell, tview, custom renderers) can map it onto their own
// color types. ok is false for Default.
func (c Color) RGB() (r, g, b int, ok bool) {
	switch c.kind {
	case colorRGB:
		return int(c.r), int(c.g), int(c.b), true
	case colorIndexed:
		r, g, b := idxToRGB(int(c.idx))
		return r, g, b, true
	default:
		return 0, 0, 0, false
	}
}

var defaultPalette = []Color{
	Indexed(39), Indexed(205), Indexed(214), Indexed(114),
	Indexed(141), Indexed(203), Indexed(75), Indexed(156),
	Indexed(179), Indexed(44),
}

type Style struct {
	Fg   Color
	Bg   Color
	Bold bool
}

func S(fg Color) Style { return Style{Fg: fg} }

func (s Style) On(bg Color) Style     { s.Bg = bg; return s }
func (s Style) WithFg(fg Color) Style { s.Fg = fg; return s }
func (s Style) Bolder() Style         { s.Bold = true; return s }

func (s Style) isZero() bool {
	return !s.Bold && s.Fg.IsZero() && s.Bg.IsZero()
}

func (s Style) eq(o Style) bool {
	return s.Bold == o.Bold && s.Fg == o.Fg && s.Bg == o.Bg
}

func (l Level) seq(st Style) string {
	if l <= LevelNone || st.isZero() {
		return ""
	}
	var parts []string
	if st.Bold {
		parts = append(parts, "1")
	}
	if !st.Fg.IsZero() {
		parts = append(parts, colorCodes(st.Fg, false, l)...)
	}
	if !st.Bg.IsZero() {
		parts = append(parts, colorCodes(st.Bg, true, l)...)
	}
	if len(parts) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(parts, ";") + "m"
}

const ansiReset = "\x1b[0m"

func colorCodes(c Color, bg bool, l Level) []string {
	base, bright, ext := 30, 90, "38"
	if bg {
		base, bright, ext = 40, 100, "48"
	}
	switch c.kind {
	case colorIndexed:
		i := int(c.idx)
		if i < 16 {
			if i < 8 {
				return []string{strconv.Itoa(base + i)}
			}
			return []string{strconv.Itoa(bright + i - 8)}
		}
		if l >= Level256 {
			if l >= LevelTrue {
				r, g, b := idxToRGB(i)
				return []string{ext, "2", strconv.Itoa(r), strconv.Itoa(g), strconv.Itoa(b)}
			}
			return []string{ext, "5", strconv.Itoa(i)}
		}
		r, g, b := idxToRGB(i)
		return basic16(nearest16([3]int{r, g, b}), base, bright)
	case colorRGB:
		switch {
		case l >= LevelTrue:
			return []string{
				ext,
				"2",
				strconv.Itoa(int(c.r)),
				strconv.Itoa(int(c.g)),
				strconv.Itoa(int(c.b)),
			}
		case l == Level256:
			return []string{ext, "5", strconv.Itoa(rgbTo256(c.r, c.g, c.b))}
		default:
			return basic16(nearest16([3]int{int(c.r), int(c.g), int(c.b)}), base, bright)
		}
	}
	return nil
}

func basic16(n int, base, bright int) []string {
	if n < 8 {
		return []string{strconv.Itoa(base + n)}
	}
	return []string{strconv.Itoa(bright + n - 8)}
}

var ansi16rgb = [16][3]int{
	{0x00, 0x00, 0x00},
	{0xcd, 0x00, 0x00},
	{0x00, 0xcd, 0x00},
	{0xcd, 0xcd, 0x00},
	{0x00, 0x00, 0xee},
	{0xcd, 0x00, 0xcd},
	{0x00, 0xcd, 0xcd},
	{0xe5, 0xe5, 0xe5},
	{0x7f, 0x7f, 0x7f},
	{0xff, 0x00, 0x00},
	{0x00, 0xff, 0x00},
	{0xff, 0xff, 0x00},
	{0x5c, 0x5c, 0xff},
	{0xff, 0x00, 0xff},
	{0x00, 0xff, 0xff},
	{0xff, 0xff, 0xff},
}

func nearest16(rgb [3]int) int {
	best, bestD := 0, 1<<30
	for i, c := range ansi16rgb {
		dr, dg, db := rgb[0]-c[0], rgb[1]-c[1], rgb[2]-c[2]
		d := 2*dr*dr + 4*dg*dg + 3*db*db
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

var cubeLevels = [6]int{0, 95, 135, 175, 215, 255}

func idxToRGB(i int) (int, int, int) {
	if i < 16 {
		c := ansi16rgb[i]
		return c[0], c[1], c[2]
	}
	if i < 232 {
		n := i - 16
		return cubeLevels[n/36], cubeLevels[(n/6)%6], cubeLevels[n%6]
	}
	v := 8 + (i-232)*10
	return v, v, v
}

func rgbTo256(r, g, b uint8) int {
	R, G, B := int(r), int(g), int(b)
	if R == G && G == B {
		if R < 8 {
			return 16
		}
		if R > 238 {
			return 231
		}
		return 232 + (R-8)/10
	}
	q := func(v int) int {
		best, bd := 0, 1<<30
		for i, lv := range cubeLevels {
			if d := abs(v - lv); d < bd {
				best, bd = i, d
			}
		}
		return best
	}
	return 16 + 36*q(R) + 6*q(G) + q(B)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func toUint8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func mix(a, b Color, t float64) Color {
	if a.kind != colorRGB {
		r1, g1, b1 := idxToRGB(int(a.idx))
		a = RGB(toUint8(r1), toUint8(g1), toUint8(b1))
	}
	if b.kind != colorRGB {
		r2, g2, b2 := idxToRGB(int(b.idx))
		b = RGB(toUint8(r2), toUint8(g2), toUint8(b2))
	}
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t))
	}
	return RGB(lerp(a.r, b.r), lerp(a.g, b.g), lerp(a.b, b.b))
}
