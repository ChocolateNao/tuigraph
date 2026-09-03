package tuichart

import (
	"strings"
	"testing"
	"time"
)

func candleFixture() *Candlestick {
	t0 := time.Date(2026, 8, 20, 9, 30, 0, 0, time.UTC)
	return NewCandlestick().Title("ACME").Format("Jan 02").
		Candle(t0, 100, 110, 95, 108). // up
		Candle(t0.Add(24*time.Hour), 108, 112, 90, 96).
		Candle(t0.Add(48*time.Hour), 96, 106, 94, 104).
		Candle(t0.Add(72*time.Hour), 104, 105, 88, 89)
}

func TestCandlestickRendersOHLC(t *testing.T) {
	out := renderD(candleFixture(), WithUnicode(true), WithWidth(60))
	if got := splitLines(out)[0]; runeLen(got) != 60 {
		t.Fatalf("width = %d, want 60", runeLen(got))
	}
	lines := splitLines(out)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"│", "┬", "ACME"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in output:\n%s", want, joined)
		}
	}
	if !strings.ContainsAny(joined, "\u2588#") { // block or mono body glyph
		t.Errorf("no candle bodies in output:\n%s", joined)
	}
	// some row must mix wick chars with body glyphs
	foundWick := false
	for _, l := range lines {
		if strings.Contains(l, "│") && strings.ContainsAny(l, "\u2588#%") {
			foundWick = true
		}
	}
	if !foundWick {
		t.Error("no row combines wicks and bodies")
	}
	// Y axis shows price levels (three digits expected somewhere)
	if !strings.Contains(joined, "100") && !strings.Contains(joined, "95") {
		t.Error("no price tick labels")
	}
}

func TestCandlestickColorsConfigurable(t *testing.T) {
	base := renderD(candleFixture(), WithWidth(56), WithUnicode(true), WithColor256())
	rec := NewCandlestick().Title("ACME").Format("Jan 02").UpColor(Blue).DownColor(Purple).
		Candles(candleFixture().candles...)
	altered := renderD(rec, WithWidth(56), WithUnicode(true), WithColor256())
	if base == altered {
		t.Error("UpColor/DownColor had no effect on rendering")
	}
	if c := candleFixture(); c.up != Indexed(2) || c.down != Indexed(1) {
		t.Error("default colors are not green/red")
	}
	c := candleFixture()
	c.ResetColors()
	c.UpColor(Red)
	if c.up != Red || c.down != Indexed(1) {
		t.Error("ResetColors should restore defaults only when called")
	}
}

func TestCandlestickMonoGlyphs(t *testing.T) {
	out := renderD(candleFixture(), WithNoColor(), WithUnicode(true))
	joined := strings.Join(splitLines(out), "\n")
	if !strings.Contains(joined, "#") || !strings.Contains(joined, "%") {
		t.Errorf("mono mode must distinguish up(#) and down(%%):\n%s", joined)
	}
	if strings.Contains(joined, "█") {
		t.Error("colorless mode must not use colored block glyphs")
	}
}

func TestCandlestickASCIIFallback(t *testing.T) {
	out := renderD(candlestickASCII(), WithNoColor(), WithUnicode(false), WithWidth(52))
	assertASCII(t, "candlestick", out)
	if !strings.Contains(out, "|") {
		t.Error("ascii wick missing")
	}
}

func candlestickASCII() *Candlestick {
	t0 := time.Date(2026, 8, 20, 9, 30, 0, 0, time.UTC)
	cs := NewCandlestick().Format("Jan 02").
		Candle(t0, 100, 110, 95, 108).
		Candle(t0.Add(24*time.Hour), 108, 112, 90, 96).
		Candle(t0.Add(48*time.Hour), 96, 106, 94, 104)
	return cs
}

func TestCandlestickDoji(t *testing.T) {
	t0 := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	cs := NewCandlestick().Format("15:04").
		Candle(t0, 50, 55, 45, 50). // perfect doji
		Candle(t0.Add(time.Hour), 50, 52, 48, 51)
	out := renderD(cs, WithUnicode(true), WithWidth(40))
	if !strings.Contains(out, "─") {
		t.Errorf("doji dash missing:\n%s", out)
	}
}

func TestCandlestickEmptyAndFormat(t *testing.T) {
	if out := renderD(
		NewCandlestick(),
		WithUnicode(true),
		WithWidth(40),
	); !strings.Contains(
		out,
		"(no data)",
	) {
		t.Errorf("empty guard missing:\n%s", out)
	}
	t0 := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	cs := NewCandlestick().Format("Jan 02").Candle(t0, 1, 2, 0.5, 1.5)
	if out := renderD(cs, WithWidth(44)); !strings.Contains(out, "Mar 05") {
		t.Errorf("Format layout ignored:\n%s", out)
	}
	h := NewCandlestick().HeightHint(70)
	if h < 8 {
		t.Errorf("HeightHint = %d", h)
	}
}
