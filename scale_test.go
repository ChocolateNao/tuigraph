package tuichart

import (
	"math"
	"testing"
)

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
