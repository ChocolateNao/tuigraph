package tuichart

import (
	"math"
	"strconv"
	"strings"
)

// Kind identifies the interpolation method of a Scale.
type Kind uint8

const (
	// Linear is a uniform scale from Min to Max.
	Linear Kind = iota
	// Logarithmic is a base-10 logarithmic scale.
	Logarithmic
)

// Scale maps a numeric range [Min, Max] to [0, 1] using the chosen Kind.
type Scale struct {
	Kind Kind
	Min  float64
	Max  float64
}

func newScale(kind Kind, lo, hi float64, pad float64) Scale {
	if math.IsNaN(lo) || math.IsInf(lo, 0) {
		lo = 0
	}
	if math.IsNaN(hi) || math.IsInf(hi, 0) {
		hi = 1
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	if kind == Logarithmic {
		smallest := smallestPositive([]float64{lo, hi})
		if !math.IsNaN(smallest) && smallest > 0 {
			if lo <= 0 {
				lo = smallest / 10
			}
			l0, l1 := math.Log10(lo), math.Log10(hi)
			d := (l1 - l0) * pad
			return Scale{Kind: kind, Min: math.Pow(10, l0-d), Max: math.Pow(10, l1+d)}
		}
		kind = Linear
	}
	if hi-lo < 1e-12 {
		center := (hi + lo) / 2
		if center == 0 {
			lo, hi = -1, 1
		} else {
			lo, hi = center*0.9, center*1.1
		}
	}
	d := (hi - lo) * pad
	return Scale{Kind: kind, Min: lo - d, Max: hi + d}
}

func fixedScale(kind Kind, lo, hi float64) Scale {
	if lo >= hi {
		if lo == 0 {
			lo, hi = -1, 1
		} else {
			hi = lo + math.Abs(lo)*0.01 + 1e-9
		}
	}
	return Scale{Kind: kind, Min: lo, Max: hi}
}

// Map normalises v to the range [0, 1] according to the scale kind.
func (s Scale) Map(v float64) float64 {
	if s.Kind == Logarithmic {
		min := s.Min
		if min <= 0 {
			min = 1e-9
		}
		max := s.Max
		if max <= min {
			max = min * 10
		}
		lv := math.Log10(v)
		lmin, lmax := math.Log10(min), math.Log10(max)
		return (lv - lmin) / (lmax - lmin)
	}
	return (v - s.Min) / (s.Max - s.Min)
}

// Ticks returns evenly-spaced tick marks suitable for the scale's range.
// target is the approximate number of ticks desired.
func (s Scale) Ticks(target int) []Tick {
	if target < 2 {
		target = 2
	}
	if s.Kind == Logarithmic {
		return logTicks(s.Min, s.Max, target)
	}
	return niceTicks(s.Min, s.Max, target)
}

// Tick represents a single axis tick with a numeric value and its label.
type Tick struct {
	Label string
	Value float64
}

func niceTicks(lo, hi float64, target int) []Tick {
	if lo > hi {
		lo, hi = hi, lo
	}
	step := niceStep(hi-lo, target)
	start := math.Ceil(lo/step-1e-9) * step
	var out []Tick
	for k := int(math.Round(start / step)); ; k++ {
		v := float64(k) * step
		if v > hi+step*1e-6 {
			break
		}
		if v < lo-step*1e-6 {
			continue
		}
		out = append(out, Tick{Value: v, Label: fmtTick(v, step)})
		if len(out) > 64 {
			break
		}
	}
	return out
}

func logTicks(lo, hi float64, target int) []Tick {
	if lo <= 0 {
		lo = 1e-9
	}
	if hi <= lo {
		hi = lo * 10
	}
	l0 := math.Floor(math.Log10(lo))
	l1 := math.Ceil(math.Log10(hi))
	var out []Tick
	for e := int(l0); e <= int(l1); e++ {
		mults := []float64{1, 5}
		if math.Log10(hi)-math.Log10(lo) <= 1.2 {
			mults = []float64{1, 2, 3, 5, 7}
		}
		for _, m := range mults {
			v := m * math.Pow(10, float64(e))
			if v >= lo*0.999 && v <= hi*1.001 {
				out = append(out, Tick{Value: v, Label: fmtSci(v)})
			}
		}
		if len(out) > 64 {
			break
		}
	}
	return out
}

func niceStep(span float64, target int) float64 {
	if span <= 0 || math.IsNaN(span) {
		return 1
	}
	raw := span / float64(target)
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	f := raw / mag
	switch {
	case f <= 1:
		f = 1
	case f <= 2:
		f = 2
	case f <= 5:
		f = 5
	default:
		f = 10
	}
	return f * mag
}

func fmtTick(v, step float64) string {
	dec := 0
	if f := math.Abs(step); f > 0 && f < 10 {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if i := strings.IndexByte(s, '.'); i >= 0 {
			dec = len(s) - i - 1
			if dec > 8 {
				dec = 8
			}
		}
	}
	s := strconv.FormatFloat(v, 'f', dec, 64)
	if dec > 0 {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s
}

func fmtSci(v float64) string {
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'g', 4, 64)
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		mant := strings.TrimRight(strings.TrimRight(s[:i], "0"), ".")
		exp := strings.TrimLeft(s[i+1:], "+")
		exp = strings.TrimLeft(exp, "0")
		if exp == "" {
			exp = "0"
		}
		return mant + "e" + exp
	}
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

// FormatValue formats v as a human-readable string, using scientific
// notation for very large or very small values.
func FormatValue(v float64) string {
	a := math.Abs(v)
	if a >= 1e6 || (a > 0 && a < 1e-4) {
		return fmtSci(v)
	}
	s := strconv.FormatFloat(v, 'f', 4, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s
}

func smallestPositive(vals []float64) float64 {
	best := math.NaN()
	for _, v := range vals {
		if v > 0 && (math.IsNaN(best) || v < best) {
			best = v
		}
	}
	return best
}
