package tuichart

import (
	"fmt"
	"io"
	"strings"
)

// Options configures a Chart at construction time.
type Options struct {
	palette         []Color
	levelOverride   Level
	width           int
	gap             int
	diagramHeight   int
	hasLevel        bool
	unicodeOverride int8
}

// Option mutates Chart options.
type Option func(*Options)

// WithWidth overrides the detected terminal width with a fixed column count.
func WithWidth(w int) Option {
	return func(o *Options) {
		if w > 0 {
			o.width = w
		}
	}
}

// WithGap sets the number of blank rows between diagram rows.
func WithGap(rows int) Option { return func(o *Options) { o.gap = rows } }

// WithPalette replaces the default color palette with the provided colors.
func WithPalette(cs ...Color) Option { return func(o *Options) { o.palette = cs } }

// WithDiagramHeight sets the default height in rows for diagrams that do not
// provide their own.
func WithDiagramHeight(rows int) Option {
	return func(o *Options) {
		if rows > 2 {
			o.diagramHeight = rows
		}
	}
}

// WithUnicode overrides automatic unicode detection: true forces braille
// and box-drawing glyphs, false forces pure ASCII.
func WithUnicode(on bool) Option {
	return func(o *Options) {
		if on {
			o.unicodeOverride = 1
		} else {
			o.unicodeOverride = 0
		}
	}
}

// WithProfile overrides automatic color-level detection with a fixed Level.
func WithProfile(l Level) Option {
	return func(o *Options) { o.levelOverride, o.hasLevel = l, true }
}

// WithNoColor forces monochrome output with no ANSI escape sequences.
func WithNoColor() Option { return WithProfile(LevelNone) }

// WithColor16 forces 16-color ANSI output.
func WithColor16() Option { return WithProfile(Level16) }

// WithColor256 forces 256-color ANSI output.
func WithColor256() Option { return WithProfile(Level256) }

// WithTrueColor forces 24-bit truecolor ANSI output.
func WithTrueColor() Option { return WithProfile(LevelTrue) }

func (o *Options) apply(info Info) Info {
	if o.hasLevel {
		info.Level = o.levelOverride
	}
	if o.unicodeOverride >= 0 {
		info.Unicode = o.unicodeOverride == 1
	}
	return info
}

type rowEntry struct{ d Drawable }

// Chart renders multiple diagrams into one string, stacked vertically or
// arranged side by side with Row.
type Chart struct {
	title      string
	rows       [][]rowEntry
	opts       Options
	titleAlign Align
}

// New creates an empty chart container. Options configure rendering behavior;
// they override automatic terminal detection.
func New(opts ...Option) *Chart {
	c := &Chart{opts: Options{unicodeOverride: -1}, titleAlign: AlignCenter}
	for _, opt := range opts {
		opt(&c.opts)
	}
	return c
}

// Title sets the container title rendered above all diagrams.
func (c *Chart) Title(t string) *Chart { c.title = t; return c }

// TitleAlign sets where the container title sits: AlignLeft, AlignCenter
// (default) or AlignRight.
func (c *Chart) TitleAlign(a Align) *Chart { c.titleAlign = a; return c }

// Add appends a diagram on its own row.
func (c *Chart) Add(d Drawable) *Chart {
	return c.Row(d)
}

// Row places diagrams side by side, splitting available width evenly.
func (c *Chart) Row(ds ...Drawable) *Chart {
	entry := make([]rowEntry, len(ds))
	for i, d := range ds {
		entry[i] = rowEntry{d: d}
	}
	c.rows = append(c.rows, entry)
	return c
}

// Clear removes all diagrams and the title from the chart.
func (c *Chart) Clear() *Chart { c.rows = nil; return c }

// Len returns the total number of diagrams across all rows.
func (c *Chart) Len() int {
	n := 0
	for _, r := range c.rows {
		n += len(r)
	}
	return n
}

// Reset removes all diagrams but keeps options.
func (c *Chart) Reset() *Chart { c.Clear(); return c }

func defaultDiagramHeight(width int) int {
	h := width / 3
	if h > 20 {
		h = 20
	}
	if h < 9 {
		h = 9
	}
	return h
}

// layout resolves the effective width and renders every diagram row into
// a slice of finished line strings (ANSI-styled per the resolved profile).
func (c *Chart) layout(width int) []string {
	w := 0
	if width > 0 {
		w = width
	}
	info := c.opts.apply(Detect())
	rc := newCtx(info)
	if len(c.opts.palette) > 0 {
		rc.Palette = c.opts.palette
	}
	if w <= 0 {
		w = info.W
		if c.opts.width > 0 {
			w = c.opts.width
		}
	}
	if w < 10 {
		w = 10
	}
	const sepW = 2

	var lines []string
	if c.title != "" {
		lines = append(lines, alignStyled(c.title, w, c.titleAlign, info.Unicode))
		lines = append(lines, "")
	}

	for ri, row := range c.rows {
		if ri > 0 || len(lines) > 0 && c.title != "" {
			for g := 0; g < maxInt(c.opts.gap, 1); g++ {
				lines = append(lines, "")
			}
		}
		k := len(row)
		seg := (w - (k-1)*sepW) / k
		if seg < 8 {
			seg = 8
		}
		heights := make([]int, k)
		canvases := make([]*Canvas, k)
		maxLines := 0
		for i, e := range row {
			h := e.d.HeightHint(seg)
			if h <= 0 {
				h = defaultDiagramHeight(w)
			}
			if c.opts.diagramHeight > 0 && e.d.HeightHint(seg) == 0 {
				h = c.opts.diagramHeight
			}
			if h < 3 {
				h = 3
			}
			heights[i] = h
			cv := NewCanvas(seg, h)
			e.d.Draw(rc, cv)
			canvases[i] = cv
		}
		rendered := make([][]string, k)
		for i, cv := range canvases {
			s := cv.Render(info.Level)
			s = trimTrailingNewlines(s)
			rendered[i] = splitLines(s)
			if len(rendered[i]) > maxLines {
				maxLines = len(rendered[i])
			}
		}
		for ln := 0; ln < maxLines; ln++ {
			var sb []byte
			for i := 0; i < k; i++ {
				if i > 0 {
					sb = append(sb, ' ', ' ')
				}
				if ln < len(rendered[i]) {
					sb = append(sb, rendered[i][ln]...)
				} else {
					sb = append(sb, padSpaces(seg+2)...)
				}
			}
			lines = append(lines, stringsTrimRight(string(sb)))
		}
	}
	return lines
}

// Render lays out every added diagram. With no argument the detected
// terminal width is used.
func (c *Chart) Render(width ...int) string {
	w := 0
	if len(width) > 0 {
		w = width[0]
	}
	lines := c.layout(w)
	var out []byte
	for i, l := range lines {
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, l...)
	}
	out = append(out, '\n')
	return string(out)
}

// RenderLines is the embedding-friendly form of Render: it returns one
// string per terminal row, ANSI-styled according to the resolved profile,
// without trailing newlines. Feed the result to frameworks that manage
// their own line buffers (Bubble Tea views, tview TextView, ...).
func (c *Chart) RenderLines(width ...int) []string {
	w := 0
	if len(width) > 0 {
		w = width[0]
	}
	return c.layout(w)
}

// String renders at the detected terminal width.
func (c *Chart) String() string { return c.Render(0) }

// RenderTo writes the rendered chart to any io.Writer — stdout, a file, a
// network connection, an http.ResponseWriter — instead of returning a
// string. The width follows the variadic convention of Render: detected
// terminal width when omitted, explicit otherwise. Styling is decided by
// the chart's own options (WithNoColor / WithProfile), so the same call
// serves ANSI terminals and plain-text sinks like log files.
//
// The chart is rendered fully in memory and then written; a short write
// or any writer error aborts and is returned wrapped with context.
// Rendering itself never fails.
func (c *Chart) RenderTo(w io.Writer, width ...int) error {
	out := c.Render(width...)
	n, err := w.Write([]byte(out))
	if err == nil && n != len(out) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return fmt.Errorf("tuichart: rendering to writer failed: %w", err)
	}
	return nil
}

// WriteTo implements io.WriterTo so charts compose with io.Copy and any
// other writer-based plumbing. It renders at the detected terminal width;
// use RenderTo for an explicit width or wrapped errors.
func (c *Chart) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(c.Render()))
	return int64(n), err
}

// Reader returns an io.Reader over the rendered chart (detected terminal
// width), so charts plug into reader-based APIs: io.Copy, http response
// bodies, multipart writers, and friends. The content is fully rendered
// up front; reading never fails.
func (c *Chart) Reader() io.Reader {
	return strings.NewReader(c.Render())
}

func alignStyled(s string, w int, a Align, uni bool) string {
	n := runeLen(s)
	if n >= w {
		return ellipTrunc(s, w, uni)
	}
	switch a {
	case AlignRight:
		return repeatStr(" ", w-n) + s
	case AlignLeft:
		return s
	default:
		pad := (w - n) / 2
		if pad < 0 {
			pad = 0
		}
		return repeatStr(" ", pad) + s
	}
}

func padSpaces(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return b
}

func repeatStr(s string, n int) string {
	return string(padSpaces(n))[:0] + mulStr(s, n)
}

func mulStr(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

func trimTrailingNewlines(s string) string {
	for len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func stringsTrimRight(s string) string {
	end := len(s)
	for end > 0 && (s[end-1] == ' ') {
		end--
	}
	return s[:end]
}
