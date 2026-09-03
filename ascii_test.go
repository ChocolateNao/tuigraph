package tuichart

import (
	"strings"
	"testing"
	"time"
)

// stripANSI removes CSI escape sequences so glyph content can be inspected.
func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			// skip ESC [ ... final byte
			for j := i + 1; j < len(s); j++ {
				c := s[j]
				if c >= 0x40 && c <= 0x7e {
					i = j
					break
				}
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func assertASCII(t *testing.T, name, out string) {
	t.Helper()
	clean := stripANSI(out)
	for _, r := range clean {
		if r != '\n' && (r < 0x20 || r > 0x7e) {
			t.Errorf("%s: non-ASCII rune %q (U+%04X) leaked in unicode=false mode", name, r, r)
			return
		}
	}
}

// every diagram must degrade to pure printable ASCII when Unicode is off,
// so any dumb terminal renders something readable.
func TestASCIIPurityAllDiagrams(t *testing.T) {
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	longCats := []string{"verylongcategoryname", "another-long-one", "c"}

	diagrams := map[string]Drawable{
		"plot": NewPlot().Title("line plot").Add(
			NewLineVals("sine", []float64{1, 3, 2, 5, 4}),
			NewScatterVals("pts", []float64{2, 4, 3}),
		),
		"function": NewFunction(func(x float64) float64 { return x / 10 }).Name("f(x)"),
		"bar-v":    NewBarValues(longCats, []float64{3, 6, 2}).Title("bars with a very long title"),
		"bar-h":    NewBarValues(longCats, []float64{3, 6, 2}).Orientation(OrientHorizontal),
		"bar-stacked": NewBar(longCats,
			BarSeries{Values: []float64{1, 2, 3}},
			BarSeries{Values: []float64{2, 1, 1}}).Stacked(true),
		"histogram": NewHistogram([]float64{1, 2, 2, 3, 3, 3, 9}).Bins(5).Title("hist"),
		"pie":       NewPie().Slice("alpha-slice-name", 5).Slice("beta", 3).Slice("gamma", 2),
		"donut":     NewPie().Slice("a", 1).Slice("b", 1).Donut(true),
		"heatmap": NewHeat([][]float64{{1, 50}, {99, 25}}).
			RowLabels("row-one").ColLabels("col-one-label", "col-two-label"),
		"sparkline": NewSpark(1, 5, 3, 8, 2).Title("spark"),
		"timeline": timelineFixture().Format("15:04:05").
			Event(now.Add(time.Minute), "an-extraordinarily-long-event-label"),
		"gantt": NewGantt().Bar("research", now, now.Add(24*time.Hour)).
			Bar("build-out-the-thing", now.Add(12*time.Hour), now.Add(48*time.Hour)),
		"timeseries": func() Drawable {
			ts := NewTimeSeries().Title("series")
			ln := ts.Line("degC")
			for i := 0; i < 10; i++ {
				ln.Add(now.Add(time.Duration(i)*time.Hour), float64(i%5))
			}
			return ts
		}(),
		"candlestick": NewCandlestick().Format("Jan 02").
			Candle(now, 100, 110, 95, 108).
			Candle(now.Add(24*time.Hour), 108, 112, 90, 96),
		"gauge-blocks":   NewGauge(7, 10).Title("gauge with a long title"),
		"gauge-ascii":    gaugeByStyle(GaugeASCII),
		"gauge-brackets": gaugeByStyle(GaugeBrackets),
		"gauge-arrow":    gaugeByStyle(GaugeArrow),
		"gauge-segments": gaugeByStyle(GaugeSegments),
	}

	for name, d := range diagrams {
		out := renderD(d, WithUnicode(false), WithWidth(48))
		assertASCII(t, name, out)
	}

	// chart container with long title and side-by-side row
	g := New(WithNoColor(), WithUnicode(false), WithWidth(60)).
		Title("container title that is definitely too long for the width")
	g.Row(
		NewSpark(1, 2, 3).Title("left spark title"),
		gaugeByStyle(GaugeBlocks),
	)
	assertASCII(t, "chart-row", g.Render(60))
}

func TestEllipTruncModes(t *testing.T) {
	const s = "abcdefgh"
	if got := ellipTrunc(s, 5, true); got != "abcd…" {
		t.Errorf("uni = %q", got)
	}
	if got := ellipTrunc(s, 5, false); got != "ab..." {
		t.Errorf("ascii = %q", got)
	}
	if got := ellipTrunc(s, 2, false); got != "ab" {
		t.Errorf("tiny ascii = %q", got)
	}
	if got := ellipTrunc("ab", 5, false); got != "ab" {
		t.Errorf("no-op = %q", got)
	}
}

func TestLocaleUnicodeDetection(t *testing.T) {
	cases := []struct {
		lcAll, lcType, lang string
		want                bool
	}{
		{"en_US.UTF-8", "", "", true},
		{"C", "", "", false},
		{"C.UTF-8", "", "", true},
		{"", "de_DE.utf8", "", true},
		{"", "POSIX", "en_US.UTF-8", false},
		{"", "", "C", false},
		{"", "", "", true}, // no locale info: assume modern terminal
		{"ru_RU.koi8r", "", "", false},
	}
	for _, c := range cases {
		t.Setenv("LC_ALL", c.lcAll)
		t.Setenv("LC_CTYPE", c.lcType)
		t.Setenv("LANG", c.lang)
		if got := localeSupportsUnicode(); got != c.want {
			t.Errorf("LC_ALL=%q LC_CTYPE=%q LANG=%q -> %v want %v",
				c.lcAll, c.lcType, c.lang, got, c.want)
		}
	}
}

func TestASCIIFallbackStillInformative(t *testing.T) {
	// slope glyphs for lines, # for bars, ramp chars for heat: shape info survives
	out := renderD(
		NewPlot().Add(NewLineVals("l", []float64{1, 5, 2, 7})),
		WithUnicode(false),
		WithWidth(40),
	)
	if !strings.ContainsAny(out, "\\/|-+~") && !strings.ContainsAny(out, "xo*") {
		// some ascii line rendering artifact must exist
		t.Errorf("ascii plot has no line glyphs at all")
	}
	hm := renderD(NewHeat([][]float64{{0, 50}, {100, 25}}), WithUnicode(false), WithWidth(20))
	if !strings.ContainsAny(hm, string(heatRampASCII)) {
		t.Error("ascii heatmap lost ramp encoding")
	}
}

func TestUnicodeUserTextPassesThrough(t *testing.T) {
	// library structure degrades to ASCII, but user-provided label text is
	// rendered verbatim in either mode — transliteration is out of scope.
	out := renderD(NewGauge(5, 10).Title("ünïcode"), WithUnicode(false), WithWidth(30))
	if !strings.Contains(out, "ünïcode") {
		t.Errorf("user text lost: %q", out)
	}
	assertASCII(t, "structure", stripANSI(strings.ReplaceAll(out, "ünïcode", "unicode")))
}
