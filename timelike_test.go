package tuichart

import (
	"strings"
	"testing"
	"time"
)

func base() time.Time {
	return time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
}

func TestTimelineDetailAndSide(t *testing.T) {
	t0 := base()
	tl := NewTimeline().Format("15:04").
		Event(t0, "deploy").
		Events(
			TimelineEvent{At: t0.Add(30 * time.Minute), Label: "spike", Detail: "cpu 98%"},
			TimelineEvent{At: t0.Add(time.Hour), Label: "rollback", Side: SideAbove},
			TimelineEvent{At: t0.Add(90 * time.Minute), Label: "stable", Detail: "ok"},
		)
	out := renderD(tl, WithWidth(70))
	if !strings.Contains(out, "cpu 98%") || !strings.Contains(out, "ok") {
		t.Errorf("detail lines missing:\n%s", out)
	}
	if tl.HeightHint(70) != 9 {
		t.Errorf("HeightHint with details = %d, want 9", tl.HeightHint(70))
	}
}

func TestTimelineSidePlacement(t *testing.T) {
	t0 := base()
	tl := NewTimeline().Format("15:04").Events(
		TimelineEvent{At: t0.Add(15 * time.Minute), Label: "up-ev", Side: SideAbove},
		TimelineEvent{At: t0.Add(45 * time.Minute), Label: "down-ev", Side: SideBelow},
	)
	lines := splitLines(renderD(tl, WithWidth(60)))
	// find rows relative to the axis line row
	axis := -1
	for i, l := range lines {
		if strings.Contains(l, "─") && strings.ContainsAny(l, "◆*") {
			axis = i
			break
		}
	}
	if axis < 0 {
		for i, l := range lines {
			if strings.Contains(l, "*") {
				axis = i
				break
			}
		}
	}
	if axis < 1 || axis >= len(lines)-2 {
		t.Fatalf("odd layout:\n%s", strings.Join(lines, "\n"))
	}
	aboveRow := strings.Join(lines[:axis], "\n")
	belowRows := strings.Join(lines[axis+1:], "\n")
	if !strings.Contains(aboveRow, "up-ev") {
		t.Errorf("SideAbove label not above line:\n%s", strings.Join(lines, "\n"))
	}
	if strings.Contains(belowRows, "up-ev") {
		t.Error("SideAbove label leaked below")
	}
	if !strings.Contains(belowRows, "down-ev") {
		t.Errorf("SideBelow label not below line:\n%s", strings.Join(lines, "\n"))
	}
}

func TestGanttRenders(t *testing.T) {
	t0 := base()
	gt := NewGantt().Title("plan").Format("Jan 02").
		Bar("design", t0, t0.Add(48*time.Hour)).
		Duration("build", t0.Add(24*time.Hour), 72*time.Hour).
		Task(GanttBar{Name: "qa", Start: t0.Add(72 * time.Hour), End: t0.Add(120 * time.Hour)})
	out := renderD(gt, WithUnicode(false), WithWidth(70))
	rows := splitLines(out)
	found := map[string]bool{}
	for _, l := range rows {
		if strings.Contains(l, "design") {
			found["design"] = true
		}
		if strings.Contains(l, "build") {
			found["build"] = true
		}
		if strings.Contains(l, "qa") {
			found["qa"] = true
		}
	}
	for _, k := range []string{"design", "build", "qa"} {
		if !found[k] {
			t.Errorf("gantt row %q missing:\n%s", k, out)
		}
	}
	if !strings.Contains(out, "#") {
		t.Error("no bars rendered")
	}
	wantH := len(gt.bars) + 3
	if wantH < 7 {
		wantH = 7
	}
	if gt.HeightHint(70) != wantH {
		t.Errorf("HeightHint = %d, want %d", gt.HeightHint(70), wantH)
	}
}

func TestTimeSeriesRenders(t *testing.T) {
	t0 := base()
	ts := NewTimeSeries().Title("temp")
	temp := ts.Line("degC")
	for i := 0; i <= 24; i++ {
		temp.Add(t0.Add(time.Duration(i)*time.Hour), float64(15+i%7))
	}
	ts.Line("target").Add(t0, 18).Add(t0.Add(24*time.Hour), 18).Color(Indexed(3))

	out := renderD(ts, WithWidth(70))
	if !strings.Contains(out, "temp") == false && !strings.Contains(out, "degC") {
		t.Errorf("series missing:\n%s", out)
	}
	// auto layout for ~24h span is HH:MM
	if !strings.Contains(out, ":") {
		t.Errorf("no time-formatted ticks:\n%s", out)
	}

	ts.Format("Mon")
	out = renderD(ts, WithWidth(70))
	if !strings.Contains(out, "Sun") && !strings.Contains(out, "Mon") {
		// span is one day starting Sunday -> Sun appears at least once
		t.Errorf("custom layout ignored:\n%s", out)
	}
}

func TestTimeSeriesEmpty(t *testing.T) {
	out := renderD(NewTimeSeries(), WithWidth(40))
	if !strings.Contains(out, "(no data)") {
		t.Errorf("empty series guard missing:\n%s", out)
	}
}

func TestWrapText(t *testing.T) {
	got := wrapText("the quick brown fox jumps over", 10)
	if len(got) != 3 || got[0] != "the quick" || got[1] != "brown fox" || got[2] != "jumps over" {
		t.Errorf("wrap = %#v", got)
	}
	got = wrapText("supercalifragilistic", 6)
	if len(got) != 4 || anyLineTooLong(got, 6) {
		t.Errorf("hard split = %#v", got)
	}
}

func anyLineTooLong(lines []string, w int) bool {
	for _, l := range lines {
		if runeLen(l) > w {
			return true
		}
	}
	return false
}

func TestTimelineDetailWraps(t *testing.T) {
	t0 := base()
	long := "latency spike traced to the upstream payment gateway pool exhaustion"
	tl := NewTimeline().Format("15:04").Events(
		TimelineEvent{At: t0.Add(30 * time.Minute), Label: "spike", Detail: long},
	)
	out := renderD(tl, WithWidth(60), WithDiagramHeight(12))
	lines := splitLines(out)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "latency") || !strings.Contains(joined, "gateway") {
		t.Errorf("wrapped detail lost words:\n%s", joined)
	}
	for _, l := range lines {
		if runeLen(l) > 62 { // frame width + slack
			t.Errorf("row too long (%d): %q", runeLen(l), l)
		}
	}
	// very big string must be truncated to two detail rows
	huge := strings.Repeat("word ", 60)
	tl2 := NewTimeline().Format("15:04").
		Events(TimelineEvent{At: t0.Add(30 * time.Minute), Label: "x", Detail: huge})
	out2 := renderD(tl2, WithWidth(50))
	n := strings.Count(out2, "word")
	if n == 0 || n > 40 {
		t.Errorf("unbounded detail rendering: %d occurrences", n)
	}
	if tl.HeightHint(60) < 9 {
		t.Error("HeightHint should grow for details")
	}
	if tl2.HeightHint(50) < 11 && len(wrapText(huge, 44)) > 1 {
		t.Errorf("HeightHint = %d, want >=11 for wrapping detail", tl2.HeightHint(50))
	}
}

func TestTimelineWrapReadsTopDown(t *testing.T) {
	t0 := base()
	tl := NewTimeline().Format("15:04").Events(
		TimelineEvent{
			At: t0.Add(30 * time.Minute), Label: "cpu-peak",
			Detail: "latency spike traced to upstream gateway pool", Side: SideAbove,
		},
	)
	lines := splitLines(renderD(tl, WithWidth(44), WithDiagramHeight(11)))
	rowOf := func(s string) int {
		for i, l := range lines {
			if strings.Contains(l, s) {
				return i
			}
		}
		return -1
	}
	first, second, lbl := rowOf("latency"), rowOf("upstream"), rowOf("cpu-peak")
	if first < 0 || second < 0 || lbl < 0 {
		t.Fatalf("missing lines:\n%s", strings.Join(lines, "\n"))
	}
	if first >= second || second >= lbl {
		t.Errorf(
			"wrap order wrong: first=%d second=%d label=%d\n%s",
			first,
			second,
			lbl,
			strings.Join(lines, "\n"),
		)
	}
}

func TestTimelineClusterNoOverlap(t *testing.T) {
	t0 := base()
	tl := NewTimeline().Format("15:04").Events(
		TimelineEvent{At: t0, Label: "build-started"},
		TimelineEvent{
			At:     t0.Add(6 * time.Minute),
			Label:  "tests-green",
			Detail: "142 tests passed",
		},
		TimelineEvent{At: t0.Add(14 * time.Minute), Label: "canary", Detail: "error rate nominal"},
		TimelineEvent{
			At:     t0.Add(22 * time.Minute),
			Label:  "rollback-requested",
			Detail: "cpu throttled hard",
		},
		TimelineEvent{At: t0.Add(30 * time.Minute), Label: "fixed"},
		TimelineEvent{
			At:     t0.Add(38 * time.Minute),
			Label:  "deploy-done",
			Detail: "v2.4.1 live everywhere",
		},
	)
	for _, w := range []int{60, 90, 140} {
		out := renderD(tl, WithWidth(w), WithDiagramHeight(10))
		joined := strings.Join(splitLines(out), "\n")
		for _, lbl := range []string{
			"build-started", "tests-green", "canary",
			"rollback-requested", "fixed", "deploy-done",
		} {
			if !strings.Contains(joined, lbl) {
				t.Errorf("width %d: label %q clobbered or missing:\n%s", w, lbl, joined)
			}
		}
		// details may be edge-truncated but must never vanish entirely
		for _, d := range []string{"142", "error rate nominal", "cpu throttled hard", "v2.4.1"} {
			if !strings.Contains(joined, d) {
				t.Errorf("width %d: detail fragment %q missing:\n%s", w, d, joined)
			}
		}
	}
}
