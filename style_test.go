package tuichart

import (
	"strings"
	"testing"
)

func TestSeqNoColor(t *testing.T) {
	if LevelNone.seq(S(Red)) != "" {
		t.Error("no-color profile must emit nothing")
	}
}

func TestSeq16(t *testing.T) {
	got := Level16.seq(S(Red))
	if got != "\x1b[91m" {
		t.Errorf("bright red got %q", got)
	}
	if got := Level16.seq(S(Green)); got != "\x1b[32m" {
		t.Errorf("green got %q", got)
	}
}

func TestRGBDowngrade(t *testing.T) {
	c := RGB(255, 0, 0)
	if got := LevelTrue.seq(S(c)); got != "\x1b[38;2;255;0;0m" {
		t.Errorf("truecolor got %q", got)
	}
	if !strings.Contains(Level256.seq(S(c)), "38;5;") {
		t.Errorf("256 downgrade got %q", Level256.seq(S(c)))
	}
	if !strings.Contains(Level16.seq(S(c)), "[") {
		t.Errorf("16 downgrade got %q", Level16.seq(S(c)))
	}
}

func TestIndexedDowngrade(t *testing.T) {
	c := Indexed(200)
	s := Level16.seq(S(c))
	if s == "" || s == "\x1b[m" {
		t.Errorf("indexed 200 at level 16 got %q", s)
	}
}

func TestBgCodes(t *testing.T) {
	st := S(Default).On(Blue)
	if got := Level16.seq(st); got != "\x1b[104m" {
		t.Errorf("bg blue got %q", got)
	}
}

func TestNearest16Sanity(t *testing.T) {
	if n := nearest16([3]int{255, 0, 0}); n != 9 && n != 1 {
		t.Errorf("pure red -> %d", n)
	}
	if n := nearest16([3]int{0, 0, 0}); n != 0 {
		t.Errorf("black -> %d", n)
	}
	if n := nearest16([3]int{255, 255, 255}); n != 15 {
		t.Errorf("white -> %d", n)
	}
}

func TestMixColors(t *testing.T) {
	a := RGB(0, 0, 0)
	b := RGB(255, 255, 255)
	m := mix(a, b, 0.5)
	if m.r != 128 && m.r != 127 {
		t.Errorf("mid gray r=%d", m.r)
	}
	if mix(a, b, -1).r != 0 || mix(a, b, 2).r != 255 {
		t.Error("mix clamp failed")
	}
}

func TestEnvLevelDetection(t *testing.T) {
	defer ResetDetection()
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "")
	if lvl := detectEnvLevel(); lvl != LevelNone {
		t.Errorf("NO_COLOR should win, got %v", lvl)
	}
	t.Setenv("NO_COLOR", "")
	if lvl := detectEnvLevel(); lvl != Level256 {
		t.Errorf("256color term got %v", lvl)
	}
	t.Setenv("TERM", "dumb")
	if lvl := detectEnvLevel(); lvl != LevelNone {
		t.Errorf("dumb got %v", lvl)
	}
}
