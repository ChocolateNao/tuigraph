package graph

// Ctx carries rendering context passed to Drawables.
type Ctx struct {
	Info    Info
	Palette []Color
	next    int
}

func newCtx(info Info) *Ctx {
	pal := defaultPalette
	return &Ctx{Info: info, Palette: pal}
}

// NewRenderCtx builds a Ctx manually, e.g. when rendering a Drawable
// directly into a Canvas without going through a Chart.
func NewRenderCtx(info Info) *Ctx { return newCtx(info) }

func (rc *Ctx) Next() Color {
	c := rc.Palette[rc.next%len(rc.Palette)]
	rc.next++
	return c
}

// LegendEntry is one row of a chart legend.
type LegendEntry struct {
	Label string
	Style Style
	Glyph string
}

// Drawable is implemented by everything that can be placed in a Chart.
// Implement it to add custom diagram types; use Canvas primitives inside
// Draw to paint into the area you are given.
type Drawable interface {
	Draw(rc *Ctx, cv *Canvas)
	HeightHint(width int) int
}

func drawLegendInside(cv *Canvas, r Rect, entries []LegendEntry, unicode bool) {
	if len(entries) == 0 || r.W < 12 || r.H < 2 {
		return
	}
	const sep = "  "
	x := r.X2()
	y := r.Y
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		glyph := e.Glyph
		if glyph == "" {
			glyph = "██"
			if !unicode {
				glyph = "##"
			}
		}
		w := runeLen(glyph) + 1 + runeLen(e.Label)
		if x-w < r.X {
			x = r.X2()
			y++
			if y > r.Y2() {
				return
			}
		}
		cv.TextRight(x, y, e.Label, e.Style)
		x -= runeLen(e.Label) + 1
		cv.TextRight(x, y, glyph, e.Style)
		x -= runeLen(glyph) + len(sep)
	}
}
