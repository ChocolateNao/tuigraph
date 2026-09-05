package tuichart

import (
	"math"
	"strings"
	"testing"
)

func TestFixedScaleBothRanges(t *testing.T) {
	p := NewPlot().Add(NewLineVals("a", []float64{1, 5, 3}))
	p.SetXRange(0, 100)
	p.SetYRange(0, 200)
	out := renderD(p, WithWidth(40))
	if !strings.Contains(out, "100") || !strings.Contains(out, "200") {
		t.Errorf("fixed-range ticks missing:\n%s", out)
	}
}

func TestFixedScaleDirect(t *testing.T) {
	s := fixedScale(Linear, 0, 10)
	if s.Min != 0 || s.Max != 10 {
		t.Errorf("fixedScale(0,10) = %+v", s)
	}
	if s.Kind != Linear {
		t.Errorf("kind = %v", s.Kind)
	}
}

func TestFixedScaleDegenerate(t *testing.T) {
	s := fixedScale(Linear, 5, 5)
	if s.Min >= s.Max {
		t.Errorf("degenerate equal range (%v, %v)", s.Min, s.Max)
	}
	z := fixedScale(Linear, 0, 0)
	if z.Min >= z.Max || z.Min != -1 || z.Max != 1 {
		t.Errorf("zero degenerate (%v, %v), want (-1, 1)", z.Min, z.Max)
	}
}

func TestNiceTicksMonotonicInRange(t *testing.T) {
	cases := [][3]float64{
		{0, 100, 5},
		{-50, 50, 6},
		{0.1, 0.9, 4},
		{1e6, 2e6, 5},
	}
	for _, c := range cases {
		ticks := niceTicks(c[0], c[1], int(c[2]))
		if len(ticks) < 2 {
			t.Fatalf("ticks(%v,%v): got %d", c[0], c[1], len(ticks))
		}
		for i, tk := range ticks {
			if tk.Value < c[0]-1e-9 || tk.Value > c[1]+1e-9 {
				t.Errorf("tick %v out of range [%v,%v]", tk.Value, c[0], c[1])
			}
			if i > 0 && tk.Value <= ticks[i-1].Value {
				t.Errorf("ticks not increasing: %v then %v", ticks[i-1].Value, tk.Value)
			}
		}
	}
}

func TestDegenerateScale(t *testing.T) {
	s := newScale(Linear, 5, 5, 0)
	tks := s.Ticks(5)
	if len(tks) == 0 {
		t.Fatal("expected ticks for degenerate range")
	}
}

func TestLogScalePositiveDomain(t *testing.T) {
	s := newScale(Logarithmic, -5, 1000, defaultPad)
	if s.Min <= 0 {
		t.Fatalf("log scale min must be positive, got %v", s.Min)
	}
	for _, tk := range s.Ticks(5) {
		if tk.Value <= 0 {
			t.Errorf("log tick %v not positive", tk.Value)
		}
	}
	m := s.Map(s.Min)
	if math.IsNaN(m) || m < 0 || m > 1 {
		t.Errorf("Map(min) = %v", m)
	}
}

func TestLogScaleMapping(t *testing.T) {
	s := fixedScale(Logarithmic, 1, 100)
	if v := s.Map(1); math.Abs(v) > 1e-9 {
		t.Errorf("Map(1)=%v want 0", v)
	}
	if v := s.Map(10); math.Abs(v-0.5) > 1e-9 {
		t.Errorf("Map(10)=%v want 0.5", v)
	}
	if v := s.Map(100); math.Abs(v-1) > 1e-9 {
		t.Errorf("Map(100)=%v want 1", v)
	}
}

func TestFmtTick(t *testing.T) {
	cases := []struct {
		want string
		v    float64
		step float64
	}{
		{"20", 20, 20},
		{"5", 5, 0},
		{"2.5", 2.5, 2.5},
		{"10", 10, -30},
	}
	for _, c := range cases {
		if got := fmtTick(c.v, c.step); got != c.want {
			t.Errorf("fmtTick(%v,%v)=%q want %q", c.v, c.step, got, c.want)
		}
	}
}

func TestFormatValue(t *testing.T) {
	if FormatValue(0) != "0" {
		t.Error("zero")
	}
	if FormatValue(3.5) != "3.5" {
		t.Errorf("got %q", FormatValue(3.5))
	}
	if FormatValue(2e7) != "2e7" {
		t.Errorf("big: %q", FormatValue(2e7))
	}
}

func TestHistogramCounts(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	h := NewHistogram(data).Bins(4)
	counts, edges := h.Counts()
	sum := 0
	for _, c := range counts {
		sum += c
	}
	if sum != len(data) {
		t.Errorf("counts sum %d != %d", sum, len(data))
	}
	if len(edges) != h.bins+1 {
		t.Errorf("edges %d != bins+1 %d", len(edges), h.bins+1)
	}
}

// ── histogram.go ────────────────────────────────────────────────────────────

func TestHistogramColorReturnsSelf(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 3})
	ret := h.Color(Red)
	if ret != h {
		t.Fatal("Color did not return receiver")
	}
	if h.color != Red {
		t.Error("color not set")
	}
}

func TestHistogramNameReturnsSelf(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 3})
	ret := h.Name("freq")
	if ret != h {
		t.Fatal("Name did not return receiver")
	}
	if h.name != "freq" {
		t.Errorf("name = %q", h.name)
	}
}

func TestHistogramShowValuesReturnsSelf(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 3})
	ret := h.ShowValues(true)
	if ret != h {
		t.Fatal("ShowValues did not return receiver")
	}
	if !h.showVals {
		t.Error("ShowValues(true) did not enable showVals")
	}
}

func TestHistogramEmptyData(t *testing.T) {
	h := NewHistogram(nil)
	counts, edges := h.Counts()
	if counts != nil || edges != nil {
		t.Error("empty histogram should return nil counts/edges")
	}
}

func TestHistogramSingleValue(t *testing.T) {
	h := NewHistogram([]float64{5})
	counts, edges := h.Counts()
	if len(counts) != 1 || len(edges) != 2 {
		t.Errorf("single-value: counts=%v edges=%v", counts, edges)
	}
	if counts[0] != 1 {
		t.Errorf("count = %d, want 1", counts[0])
	}
}

func TestHistogramNaNData(t *testing.T) {
	h := NewHistogram([]float64{1, math.NaN(), 2, 3})
	counts, _ := h.Counts()
	total := 0
	for _, c := range counts {
		total += c
	}
	// NaN should be skipped; only 3 valid values counted
	if total != 3 {
		t.Errorf("total count = %d, want 3 (NaN skipped)", total)
	}
}

func TestHistogramBinsClamp(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 3})
	h.Bins(0)
	if h.bins != 1 {
		t.Errorf("Bins(0) -> %d, want 1", h.bins)
	}
	h.Bins(100)
	if h.bins != 64 {
		t.Errorf("Bins(100) -> %d, want 64", h.bins)
	}
}

func TestHistogramRenderNoColor(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 2, 3, 3, 3}).Bins(3).Title("h")
	out := renderD(h, WithWidth(40))
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestHistogramTitleReturnsSelf(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 3})
	ret := h.Title("myhist")
	if ret != h {
		t.Fatal("Title did not return receiver")
	}
}

func TestHistogramOrientationReturnsSelf(t *testing.T) {
	h := NewHistogram([]float64{1, 2, 3})
	ret := h.Orientation(OrientHorizontal)
	if ret != h {
		t.Fatal("Orientation did not return receiver")
	}
}
