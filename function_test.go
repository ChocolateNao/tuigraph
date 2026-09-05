package tuichart

import (
	"math"
	"testing"
)

func TestFunctionTitleReturnsSelf(t *testing.T) {
	f := NewFunction(math.Sin)
	ret := f.Title("sin")
	if ret != f {
		t.Fatal("Title did not return receiver")
	}
	if f.title != "sin" {
		t.Errorf("title = %q", f.title)
	}
}

func TestFunctionResetDomain(t *testing.T) {
	f := NewFunction(math.Sin).Domain(-5, 5)
	if f.lo != -5 || f.hi != 5 || !f.domainSet {
		t.Fatal("Domain not applied")
	}
	ret := f.ResetDomain()
	if ret != f {
		t.Fatal("ResetDomain did not return receiver")
	}
	if f.lo != -10 || f.hi != 10 || f.domainSet {
		t.Errorf("ResetDomain failed: lo=%v hi=%v domainSet=%v", f.lo, f.hi, f.domainSet)
	}
}

func TestFunctionSamples(t *testing.T) {
	f := NewFunction(math.Sin)
	f.Samples(200)
	if f.samples != 200 {
		t.Errorf("samples = %d", f.samples)
	}
	// n <= 8 is rejected
	f.Samples(5)
	if f.samples != 200 {
		t.Errorf("low samples accepted: %d", f.samples)
	}
	f.Samples(0)
	if f.samples != 200 {
		t.Errorf("zero samples accepted: %d", f.samples)
	}
	f.Samples(9)
	if f.samples != 9 {
		t.Errorf("boundary 9 not accepted: %d", f.samples)
	}
}

func TestFunctionNameReturnsSelf(t *testing.T) {
	f := NewFunction(math.Sin)
	ret := f.Name("myfn")
	if ret != f {
		t.Fatal("Name did not return receiver")
	}
	if f.name != "myfn" {
		t.Errorf("name = %q", f.name)
	}
}

func TestFunctionColorReturnsSelf(t *testing.T) {
	f := NewFunction(math.Sin)
	ret := f.Color(Red)
	if ret != f {
		t.Fatal("Color did not return receiver")
	}
	if f.color != Red {
		t.Errorf("color not set")
	}
}

func TestFunctionLogYReturnsSelf(t *testing.T) {
	f := NewFunction(math.Abs)
	ret := f.LogY(true)
	if ret != f {
		t.Fatal("LogY did not return receiver")
	}
	if f.yKind != Logarithmic {
		t.Errorf("yKind = %v, want Logarithmic", f.yKind)
	}
	f.LogY(false)
	if f.yKind != Linear {
		t.Errorf("LogY(false) yKind = %v", f.yKind)
	}
}

func TestNaNReturnValue(t *testing.T) {
	v := NaN()
	if !math.IsNaN(v) {
		t.Error("NaN() did not return NaN")
	}
}

func TestFunctionSampleNaNAndInf(t *testing.T) {
	f := NewFunction(func(x float64) float64 {
		if x < 0 {
			return math.NaN()
		}
		if x > 5 {
			return math.Inf(1)
		}
		return x
	})
	f.Domain(-10, 10)
	segs := f.sample(80)
	// NaN/Inf at boundaries should split into segments; at least one segment
	// should exist for the finite portion.
	if len(segs) == 0 {
		t.Fatal("no segments from NaN/Inf function")
	}
	total := 0
	for _, seg := range segs {
		total += len(seg)
	}
	if total == 0 {
		t.Error("all points were NaN/Inf")
	}
}

func TestFunctionRenderNoColor(t *testing.T) {
	f := NewFunction(math.Sin).Title("sin").Name("sin").Color(Red)
	out := renderD(f, WithWidth(40))
	if out == "" {
		t.Fatal("empty render")
	}
}

func TestFunctionHeightHint(t *testing.T) {
	f := NewFunction(math.Sin)
	h := f.HeightHint(80)
	if h < 9 || h > 24 {
		t.Errorf("default height hint %d out of range", h)
	}
	f.SetSize(5)
	if f.HeightHint(80) != 5 {
		t.Error("SetSize override ignored")
	}
}
