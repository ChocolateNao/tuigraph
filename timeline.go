package tuichart

import (
	"math"
	"sort"
	"strings"
	"time"
)

// EventSide controls where an event's label sits relative to the axis.
type EventSide uint8

const (
	// SideAuto alternates above/below by event index (default).
	SideAuto EventSide = iota
	// SideAbove pins the label above the line.
	SideAbove
	// SideBelow pins the label below the line.
	SideBelow
)

// TimelineEvent is a single moment on a Timeline.
type TimelineEvent struct {
	At    time.Time
	Label string
	// Detail is optional secondary info drawn next to the label (below it
	// for labels above the line, above it for labels below the line) in
	// the chart's detail color.
	Detail string
	// Side overrides the default alternating placement.
	Side EventSide
}

// Timeline renders dated events along a horizontal time axis with labels
// alternating above and below the line. Tick labels are formatted through
// Go time layouts: set one explicitly with Format, override completely via
// SetXFormatter, or leave both unset for automatic presets based on the
// covered span.
type Timeline struct {
	layout string
	events []TimelineEvent
	chartBase
	color       Color
	detailColor Color
}

// NewTimeline creates an empty Timeline.
func NewTimeline() *Timeline {
	return &Timeline{chartBase: newChartBase()}
}

// Title sets the chart title.
func (t *Timeline) Title(s string) *Timeline { t.SetTitle(s); return t }

// Event appends an event.
func (t *Timeline) Event(at time.Time, label string) *Timeline {
	t.events = append(t.events, TimelineEvent{At: at, Label: label})
	return t
}

// Events appends several events at once.
func (t *Timeline) Events(ev ...TimelineEvent) *Timeline {
	t.events = append(t.events, ev...)
	return t
}

// Format sets the Go time layout used for axis ticks (e.g. "2006-01-02",
// time.RFC3339, "15:04"). An empty string restores automatic presets.
func (t *Timeline) Format(layout string) *Timeline {
	t.layout = layout
	return t
}

// Color fixes the event marker color; by default palette colors cycle.
func (t *Timeline) Color(c Color) *Timeline { t.color = c; return t }

// DetailColor sets the color of event detail lines; default dim gray.
func (t *Timeline) DetailColor(c Color) *Timeline { t.detailColor = c; return t }

// SetDetailColor sets the color of event detail lines; default dim gray.
func (t *Timeline) SetDetailColor(c Color) { t.detailColor = c }

// ResetDetailColor restores the default detail line color.
func (t *Timeline) ResetDetailColor() { t.detailColor = Color{} }

// HeightHint returns the suggested height in rows for the given width.
func (t *Timeline) HeightHint(width int) int {
	if h := t.chartBase.HeightHint(width); h > 0 {
		return h
	}
	h := 7
	lineW := maxInt(width-6, 8)
	for _, e := range t.events {
		if e.Detail == "" {
			continue
		}
		h = 9 // label + single detail line on each side
		if len(wrapText(e.Detail, lineW)) > 1 {
			h = 11 // room for a wrapped second line
			break
		}
	}
	return h
}

// Draw renders the time axis, event markers, and labels into the canvas.
func (t *Timeline) Draw(rc *Ctx, cv *Canvas) {
	inner := t.frameTitle(cv, rc.Info.Unicode)
	if len(t.events) == 0 || inner.W < 8 || inner.H < 3 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}

	evs := append([]TimelineEvent(nil), t.events...)
	sort.Slice(evs, func(i, j int) bool { return evs[i].At.Before(evs[j].At) })

	minU := float64(evs[0].At.Unix())
	maxU := minU
	for _, e := range evs[1:] {
		u := float64(e.At.Unix())
		minU = math.Min(minU, u)
		maxU = math.Max(maxU, u)
	}
	pad := (maxU - minU) * 0.04
	if pad <= 0 {
		pad = 30
	}
	minU -= pad
	maxU += pad

	fmtFn := t.xFmt
	if fmtFn == nil {
		layout := t.layout
		if layout == "" {
			layout = autoTimeLayout(maxU - minU)
		}
		fmtFn = func(v float64) string { return time.Unix(int64(v), 0).Format(layout) }
	}

	markCh, lineCh := '◆', '─'
	if !rc.Info.Unicode {
		markCh, lineCh = '*', '-'
	}
	axisRow := inner.Y + inner.H/2

	dim := S(DimGray)
	cv.HLine(axisRow, inner.X, inner.X2(), lineCh, dim)

	mapCol := func(u float64) int {
		f := (u - minU) / (maxU - minU)
		return inner.X + int(f*float64(inner.W-1)+0.5)
	}

	for _, tk := range niceTicks(minU, maxU, t.tickN) {
		col := mapCol(tk.Value)
		if col < inner.X || col > inner.X2() {
			continue
		}
		ch := '┬'
		if !rc.Info.Unicode {
			ch = '+'
		}
		cv.Set(col, axisRow, ch, dim)
		lbl := ellipTrunc(fmtFn(tk.Value), maxInt(inner.W/3, 4), rc.Info.Unicode)
		start := col - runeLen(lbl)/2
		if start < inner.X {
			start = inner.X
		}
		// Prefer the bottom row for tick labels so they never collide
		// with event labels below the axis.
		row := axisRow + 1
		if inner.H >= 5 {
			row = inner.Y2()
		}
		if row > inner.Y2() || start+runeLen(lbl)-1 > inner.X2() {
			continue
		}
		clearAndWrite(cv, row, start, lbl, S(Default))
	}

	detailSt := S(t.detailColor)
	if t.detailColor.IsZero() {
		detailSt = S(DimGray)
	}

	occ := newCellOcc()

	for _, tk := range niceTicks(minU, maxU, t.tickN) {
		col := mapCol(tk.Value)
		if col < inner.X || col > inner.X2() {
			continue
		}
		ch := '┬'
		if !rc.Info.Unicode {
			ch = '+'
		}
		cv.Set(col, axisRow, ch, dim)
		lbl := ellipTrunc(fmtFn(tk.Value), maxInt(inner.W/3, 4), rc.Info.Unicode)
		start := col - runeLen(lbl)/2
		if start < inner.X {
			start = inner.X
		}
		// Prefer the bottom row for tick labels so they never collide
		// with event labels below the axis.
		row := axisRow + 1
		if inner.H >= 5 {
			row = inner.Y2()
		}
		if row > inner.Y2() || start+runeLen(lbl)-1 > inner.X2() {
			continue
		}
		clearAndWrite(cv, row, start, lbl, S(Default))
		occ.mark(row, start, start+runeLen(lbl)-1)
	}

	for i, e := range evs {
		col := mapCol(float64(e.At.Unix()))
		st := S(t.color)
		if t.color.IsZero() {
			st = S(rc.Palette[i%len(rc.Palette)])
		}
		cv.Set(col, axisRow, markCh, st.Bolder())

		if e.Label == "" && e.Detail == "" || inner.H < 4 {
			continue
		}
		var up bool
		switch e.Side {
		case SideAbove:
			up = true
		case SideBelow:
			up = false
		default:
			up = i%2 == 0
		}
		labelRow := axisRow - 1
		if !up {
			labelRow = axisRow + 1
		}

		lineW := maxInt(minInt(inner.W-4, detailWrapWidth), 8)
		dlines := wrapText(e.Detail, lineW)
		if len(dlines) > maxDetailLines {
			rest := strings.Join(dlines[maxDetailLines-1:], " ")
			dlines = append(dlines[:maxDetailLines-1], ellipTrunc(rest, lineW, rc.Info.Unicode))
		}
		t.placeEventBlock(cv, occ, inner, rc.Info.Unicode, col, labelRow, up,
			e.Label, dlines, st, detailSt)
	}
}

// detailWrapWidth caps wrapped detail lines so one event's text cannot
// hog the whole row and starve its neighbours.
const detailWrapWidth = 30

// cellOcc tracks which canvas cells are already used by text so later
// blocks can dodge instead of overwriting earlier ones.
type cellOcc struct {
	m map[int]map[int]bool
}

func newCellOcc() *cellOcc { return &cellOcc{m: make(map[int]map[int]bool)} }

func (o *cellOcc) free(row, x0, x1 int) bool {
	cells, ok := o.m[row]
	if !ok {
		return true
	}
	for x := x0; x <= x1; x++ {
		if cells[x] {
			return false
		}
	}
	return true
}

func (o *cellOcc) mark(row, x0, x1 int) {
	cells := o.m[row]
	if cells == nil {
		cells = make(map[int]bool)
		o.m[row] = cells
	}
	for x := x0; x <= x1; x++ {
		cells[x] = true
	}
}

// freeRun finds the widest free interval on row, scoring runs by their
// overlap with the wanted [lo, hi] window; returns the best run's bounds.
func (o *cellOcc) freeRun(row, lo, hi, x0, x1 int) (rs, re, score int) {
	best := -(1 << 30)
	runStart := -1
	for x := x0; x <= x1+1; x++ {
		blocked := x > x1 || !o.free(row, x, x)
		if !blocked && runStart < 0 {
			runStart = x
		}
		if blocked && runStart >= 0 {
			s := minInt(hi, x-1) - maxInt(lo, runStart) + 1
			if s > best {
				best = s
				rs, re = runStart, x-1
			}
			runStart = -1
		}
	}
	score = best
	return
}

// placeEventBlock draws one event's label plus wrapped detail lines so the
// block reads naturally top-to-bottom and never overlaps other text: the
// whole block first tries shifting sideways together, then each row falls
// back independently (details are dropped rather than garbled).
//
// Above the axis the detail stack is bottom-aligned toward the label with
// the first line on top, so wrapped text reads like normal prose.
type blockLine struct {
	text   string
	row    int
	detail bool
}

// anchorSpan picks where text of width w sits relative to its event at
// col: centered under the marker when that fits, but pinned to the marker
// near the frame edges — starting at it on the left side, ending on it on
// the right — so labels never drift away from their point. Returns the
// start column and how many cells are visible (may be < w near an edge).
func anchorSpan(col, w, x0, x1 int) (int, int) {
	if w <= 0 {
		return col, 0
	}
	ideal := col - w/2
	if ideal >= x0 && ideal+w-1 <= x1 {
		return ideal, w // centered, no drift
	}
	if ideal < x0 { // hugging the left edge: start at the marker
		st := maxInt(x0, col-1)
		if st+w-1 > x1 {
			return st, x1 - st + 1
		}
		return st, w
	}
	// hugging the right edge: end at the marker
	en := minInt(col, x1)
	st := en - w + 1
	if st < x0 {
		return x0, en - x0 + 1 // keep the head, run up to the marker
	}
	return st, w
}

// placeEventBlock draws one event's label plus its wrapped detail lines so
// the block reads naturally top-to-bottom and never overlaps other text:
// the whole block first tries shifting sideways as one unit, then each row
// falls back independently (details are dropped rather than garbled).
//
// Above the axis the detail stack sits between the frame top and the
// label with the first line on top, so wrapped text reads like prose.
func (t *Timeline) placeEventBlock(cv *Canvas, occ *cellOcc, inner Rect, uni bool,
	col, labelRow int, up bool, label string, dlines []string, labSt, detSt Style,
) {
	n := len(dlines)
	// Trim lines that would fall outside the frame, keeping the start of
	// the sentence.
	if up {
		if room := labelRow - inner.Y; n > room {
			n = maxInt(room, 0)
			dlines = dlines[:n]
		}
	} else {
		if room := inner.Y2() - labelRow; n > room {
			n = maxInt(room, 0)
			dlines = dlines[:n]
		}
	}

	var block []blockLine
	if up {
		// Bottom-aligned stack hugging the label, first line on top so the
		// text reads naturally downward.
		for j := 0; j < n; j++ {
			block = append(block, blockLine{row: labelRow - n + j, text: dlines[j], detail: true})
		}
	} else {
		for j := 0; j < n; j++ {
			block = append(block, blockLine{row: labelRow + 1 + j, text: dlines[j], detail: true})
		}
	}
	if label != "" {
		block = append(block, blockLine{row: labelRow, text: label})
	}
	if len(block) == 0 {
		return
	}

	styleOf := func(b blockLine) Style {
		if b.detail {
			return detSt
		}
		return labSt
	}
	widths := make([]int, len(block))
	for k, b := range block {
		widths[k] = runeLen(b.text)
	}

	// tryPlace lays the whole block out with each row anchored via
	// anchorSpan at column col+shift. When allowTrunc is set, rows that no
	// longer fit the frame are ellipsized instead of rejected.
	tryPlace := func(shift int, allowTrunc bool) bool {
		spans := make([][2]int, len(block))
		for k, b := range block {
			w := widths[k]
			if w == 0 {
				continue
			}
			st, n := anchorSpan(col+shift, w, inner.X, inner.X2())
			if n <= 0 || (!allowTrunc && n < w) {
				return false
			}
			en := st + n - 1
			if !occ.free(b.row, st, en) {
				return false
			}
			spans[k] = [2]int{st, en}
		}
		for k, b := range block {
			w := widths[k]
			if w == 0 {
				continue
			}
			st := spans[k][0]
			txt := b.text
			if span := spans[k][1] - st + 1; span < w {
				txt = ellipTrunc(txt, span, uni)
			}
			clearAndWrite(cv, b.row, st, txt, styleOf(b))
			occ.mark(b.row, st-1, spans[k][1]+1) // pad: keep a gap
		}
		return true
	}

	if tryPlace(0, false) {
		return
	}
	// Collision: slide the block sideways while staying coherent.
	half := inner.W/2 + 1
	for dx := 1; dx <= half; dx++ {
		if tryPlace(dx, false) || tryPlace(-dx, false) {
			return
		}
	}
	// Last joint chance: keep rows anchored at the point even if that
	// means truncating near the frame edge.
	if tryPlace(0, true) {
		return
	}

	// Fallback: place each row independently near the event column.
	for _, b := range block {
		w := runeLen(b.text)
		if w == 0 {
			continue
		}
		rs, re, _ := occ.freeRun(b.row, col-w/2, col+w/2, inner.X, inner.X2())
		avail := re - rs + 1
		if avail <= 0 || (b.detail && avail < 6) {
			continue // no usable room on this row; drop rather than garble
		}
		txt := b.text
		if avail < w {
			txt = ellipTrunc(txt, avail, uni)
		}
		st := col - runeLen(txt)/2
		st = maxInt(st, rs)
		st = minInt(st, re-runeLen(txt)+1)
		clearAndWrite(cv, b.row, st, txt, styleOf(b))
		occ.mark(b.row, st-1, st+runeLen(txt))
	}
}

// maxDetailLines caps how many rows a wrapped detail may occupy.
const maxDetailLines = 2

// wrapText breaks s into lines of at most w cells, preferring spaces;
// words longer than w are hard-split.
func wrapText(s string, w int) []string {
	w = maxInt(w, 4)
	var lines []string
	for _, para := range strings.Split(s, "\n") {
		cur := ""
		flush := func() {
			if cur != "" {
				lines = append(lines, cur)
				cur = ""
			}
		}
		for _, word := range strings.Split(para, " ") {
			for runeLen(word) > w {
				flush()
				lines = append(lines, truncStr(word, w))
				word = string([]rune(word)[w:])
			}
			if cur == "" {
				cur = word
				continue
			}
			if runeLen(cur)+1+runeLen(word) <= w {
				cur += " " + word
			} else {
				lines = append(lines, cur)
				cur = word
			}
		}
		flush()
	}
	return lines
}

// writeLabel draws a centered label at the given column, clamped to the
// frame: it shifts sideways near the edges before falling back to
// truncation, so full text survives whenever it fits somewhere on the row.
func writeLabel(cv *Canvas, row, col int, text string, st Style, inner Rect, uni bool) {
	if row < inner.Y || row > inner.Y2() {
		return
	}
	lbl := ellipTrunc(text, inner.W-2, uni)
	start := col - runeLen(lbl)/2
	if start < inner.X {
		start = inner.X
	}
	if end := start + runeLen(lbl) - 1; end > inner.X2() {
		// Prefer shifting left over truncating.
		start = inner.X2() - runeLen(lbl) + 1
		if start < inner.X {
			start = inner.X
			lbl = ellipTrunc(text, inner.X2()-inner.X+1, uni)
		}
	}
	if runeLen(lbl) == 0 {
		return
	}
	clearAndWrite(cv, row, start, lbl, st)
}

// clearAndWrite overwrites the target segment so labels never mix with
// leftover glyphs.
func clearAndWrite(cv *Canvas, y, x int, s string, st Style) {
	for i := 0; i < runeLen(s); i++ {
		cv.Set(x+i, y, ' ', Style{})
	}
	cv.Text(x, y, s, st)
}

func autoTimeLayout(spanSec float64) string {
	switch {
	case spanSec < 90:
		return "15:04:05"
	case spanSec < 24*3600:
		return "15:04"
	case spanSec < 30*24*3600:
		return "Jan 02 15:04"
	default:
		return "2006-01-02"
	}
}
