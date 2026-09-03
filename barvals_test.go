package tuichart

import (
	"strings"
	"testing"
)

// Bars rooted at 0 must occupy the zero axis row: no eighth-block
// slivers (▁-▆), and at least one full block on that row.
func TestBarBaselineFlush(t *testing.T) {
	out := renderD(NewBarValues([]string{"a", "b", "c", "d", "e"},
		[]float64{1, 2, 3, 5, 8}), WithUnicode(true), WithWidth(50))
	partial := "▁▂▃▄▅▆"
	foundBlock := false
	for _, line := range splitLines(out) {
		if !strings.Contains(line, "┤") || !isZeroAxisRow(line) {
			continue
		}
		if strings.ContainsAny(line, partial) {
			t.Errorf("ragged baseline slivers on zero axis row: %q", line)
		}
		if strings.Contains(line, "█") {
			foundBlock = true
		}
	}
	if !foundBlock {
		t.Errorf("zero axis row not occupied by bars:\n%s", out)
	}
}

func TestNegativeBarsFlushTop(t *testing.T) {
	out := renderD(NewBarValues([]string{"a", "b"}, []float64{-3, -7}),
		WithUnicode(true), WithWidth(40))
	partial := "▁▂▃▄▅▆"
	for _, line := range splitLines(out) {
		if strings.Contains(line, "┤") && isZeroAxisRow(line) &&
			strings.ContainsAny(line, partial) {
			t.Errorf("negative bars ragged at zero axis: %q", line)
		}
	}
}

func TestBarShowValues(t *testing.T) {
	bc := NewBarValues([]string{"a", "b"}, []float64{12, 7}).ShowValues(true)
	out := renderD(bc)
	if !strings.Contains(out, "12") || !strings.Contains(out, "7") {
		t.Errorf("value labels missing:\n%s", out)
	}
	bc.ResetShowValues()
	out = renderD(bc)
	if strings.Contains(out, "12") {
		t.Errorf("labels rendered after reset:\n%s", out)
	}
}

func TestStackedShowValuesTotalsOnly(t *testing.T) {
	bc := NewBar([]string{"x"},
		BarSeries{Name: "a", Values: []float64{1}},
		BarSeries{Name: "b", Values: []float64{2}}).Stacked(true).ShowValues(true)
	out := renderD(bc, WithUnicode(false))
	if !strings.Contains(out, "3") { // total labeled, not parts
		t.Errorf("stacked total label missing:\n%s", out)
	}
}

func TestHeatmapCellWidthAndValues(t *testing.T) {
	hm := NewHeat([][]float64{{10, 20}, {30, 40}}).CellWidth(6).ShowValues(true)
	out := renderD(hm, WithWidth(40))
	if !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Errorf("cell values missing:\n%s", out)
	}
	// fixed width: grid spans gutter + 2*6 columns, not stretched to frame
	lines := splitLines(out)
	found := false
	for _, ln := range lines {
		if strings.Contains(ln, "@") || strings.Contains(ln, "%") || strings.Contains(ln, "*") {
			trimmed := strings.TrimRight(ln, "│ ")
			if n := runeLen(strings.TrimRight(trimmed, " ")); n > 30 {
				t.Errorf("cells stretched despite CellWidth(6): %q", ln)
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("no heatmap cells rendered")
	}

	hm.ResetCellWidth()
	hm.ResetShowValues()
	out = renderD(hm, WithWidth(40))
	if strings.Contains(out, "20") || strings.Contains(out, "30") {
		t.Errorf("values rendered after reset:\n%s", out)
	}
}

func TestHeatmapValuesSkippedWhenNarrow(t *testing.T) {
	// 1234.5 is an interior value: it can only come from a cell label,
	// never from the min..max legend under the grid.
	hm := NewHeat([][]float64{{10, 1234.5, 70}}).CellWidth(3).ShowValues(true)
	out := renderD(hm, WithWidth(24))
	if strings.Contains(out, "1234.5") {
		t.Errorf("label drawn although cell too narrow:\n%s", out)
	}
	hm.CellWidth(8)
	out = renderD(hm, WithWidth(30))
	if !strings.Contains(out, "1234.5") {
		t.Errorf("label missing once cell wide enough:\n%s", out)
	}
}

func TestPieShowValues(t *testing.T) {
	pc := NewPie().Slice("home", 42).Slice("usr", 25).ShowValues(true)
	out := renderD(pc, WithWidth(70))
	if !strings.Contains(out, "home 42") || !strings.Contains(out, "usr 25") {
		t.Errorf("legend values missing:\n%s", out)
	}
	if !strings.Contains(out, "(") {
		t.Error("percentage lost when values shown")
	}
}

func isZeroAxisRow(line string) bool {
	i := strings.Index(line, "┤")
	if i < 0 {
		return false
	}
	start := i
	for start > 0 && line[start-1] >= '0' && line[start-1] <= '9' {
		start--
	}
	return strings.TrimRight(line[start:i], " ") == "0"
}

func TestHeatmapColLabelsVisible(t *testing.T) {
	hm := NewHeat([][]float64{{10, 20}, {30, 40}}).
		ColLabels("left-col", "right-col").
		RowLabels("r1", "r2")
	out := renderD(hm, WithWidth(44))
	lines := splitLines(out)
	if !strings.Contains(lines[1], "left-col") {
		t.Errorf("column labels missing on top inner row:\n%s", out)
	}
	// labels must not be overwritten by the grid: the label text survives
	// on its line even though cells start right below
	idx := strings.Index(out, "left-col")
	if idx < 0 {
		t.Fatal("no labels rendered")
	}
	if !strings.Contains(out[idx:], "█") || !strings.Contains(out[idx:], "-") {
		// grid content appears after the label line
		if !strings.Contains(strings.Join(lines[2:], "\n"), "-") {
			t.Errorf("grid missing below labels:\n%s", out)
		}
	}
}

// Bars whose value is smaller than one cell row must render as a single
// partial segment on the axis row, not a full block.
func TestShortBarPartialSegment(t *testing.T) {
	out := renderD(NewBarValues([]string{"t1", "t2", "big"}, []float64{1, 2, 90}),
		WithUnicode(true), WithWidth(44))
	var zeroRow string
	for _, line := range splitLines(out) {
		if strings.Contains(line, "┤") && isZeroAxisRow(line) {
			zeroRow = line
			break
		}
	}
	if zeroRow == "" {
		t.Fatal("no zero axis row found")
	}
	if !strings.Contains(zeroRow, "▁") {
		t.Errorf("short bars not rendered as partial segments: %q", zeroRow)
	}
	if strings.Count(zeroRow, "█") == 0 {
		t.Errorf("tall bar lost its axis-row block: %q", zeroRow)
	}
}

func TestNegativeShortBarPartialSegment(t *testing.T) {
	out := renderD(NewBarValues([]string{"s", "l"}, []float64{-0.4, -8}),
		WithUnicode(true), WithWidth(36))
	found := false
	for _, line := range splitLines(out) {
		if strings.Contains(line, "┤") && isZeroAxisRow(line) {
			if !strings.Contains(line, "▂") {
				t.Errorf("short negative bar not partial on axis row: %q", line)
			}
			if strings.Count(line, "█") == 0 {
				t.Errorf("tall negative bar missing block on axis row: %q", line)
			}
			found = true
		}
	}
	if !found {
		t.Error("no zero axis row found")
	}
}
