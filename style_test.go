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
	c := HotPink
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

func TestSetProfileDetect(t *testing.T) {
	defer ResetDetection()
	SetProfile(Level16)
	info := DetectWriter(nil)
	if info.Level != Level16 {
		t.Errorf("Detect level = %v, want Level16", info.Level)
	}
	SetProfile(LevelTrue)
	info = DetectWriter(nil)
	if info.Level != LevelTrue {
		t.Errorf("level after SetProfile(True) = %v", info.Level)
	}
}

func TestSetUnicodeDetect(t *testing.T) {
	defer ResetDetection()
	SetProfile(Level16)
	SetUnicode(false)
	info := DetectWriter(nil)
	if info.Unicode {
		t.Error("SetUnicode(false) ignored")
	}
	SetUnicode(true)
	info = DetectWriter(nil)
	if !info.Unicode {
		t.Error("SetUnicode(true) ignored")
	}
}

func TestLevelNoneForcesUnicodeOff(t *testing.T) {
	defer ResetDetection()
	SetProfile(LevelNone)
	SetUnicode(true)
	info := DetectWriter(nil)
	if info.Unicode {
		t.Error("LevelNone should force unicode off even when enabled")
	}
}

func TestApplyOverrides(t *testing.T) {
	defer ResetDetection()
	SetProfile(Level16)
	SetUnicode(false)
	detectMu.Lock()
	got := applyOverrides(Info{Level: Level256, Unicode: true, W: 4, H: 4})
	detectMu.Unlock()
	if got.Level != Level16 || got.Unicode || got.W != 4 || got.H != 4 {
		t.Errorf("applyOverrides = %+v", got)
	}
}

func TestResetDetectionClears(t *testing.T) {
	SetProfile(Level16)
	SetUnicode(true)
	ResetDetection()
	detectMu.Lock()
	got := applyOverrides(Info{Level: LevelNone, Unicode: false})
	detectMu.Unlock()
	if got.Level != LevelNone || got.Unicode {
		t.Errorf("applyOverrides after reset = %+v", got)
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

// ── style.go ────────────────────────────────────────────────────────────────

func TestStyleWithFg(t *testing.T) {
	st := S(Red)
	orig := st
	newSt := st.WithFg(Blue)
	if newSt.Fg != Blue {
		t.Errorf("WithFg Fg = %v, want Blue", newSt.Fg)
	}
	// Original must not be mutated
	if st.Fg != orig.Fg {
		t.Error("WithFg mutated original")
	}
}

func TestStyleWithFgZero(t *testing.T) {
	st := S(Red).WithFg(Default)
	if !st.Fg.IsZero() {
		t.Error("WithFg(Default) should produce zero fg")
	}
}

func TestStyleOnReturnsCopy(t *testing.T) {
	st := S(Red)
	newSt := st.On(Blue)
	if newSt.Bg != Blue {
		t.Error("On did not set bg")
	}
	if st.Bg != Default {
		t.Error("On mutated original")
	}
}

func TestStyleBolderReturnsCopy(t *testing.T) {
	st := S(Red)
	newSt := st.Bolder()
	if !newSt.Bold {
		t.Error("Bolder did not set bold")
	}
	if st.Bold {
		t.Error("Bolder mutated original")
	}
}

func TestColorRGB(t *testing.T) {
	c := RGB(10, 20, 30)
	r, g, b, ok := c.RGB()
	if !ok || r != 10 || g != 20 || b != 30 {
		t.Errorf("RGB() = %d,%d,%d,%v", r, g, b, ok)
	}
}

func TestIndexedClamp(t *testing.T) {
	lo := Indexed(-5)
	if lo.idx != 0 {
		t.Errorf("negative index clamped to %d", lo.idx)
	}
	hi := Indexed(300)
	if hi.idx != 255 {
		t.Errorf("overflow index clamped to %d", hi.idx)
	}
}

func TestStyleEq(t *testing.T) {
	a := Style{Fg: Red, Bold: true}
	b := Style{Fg: Red, Bold: true}
	if !a.eq(b) {
		t.Error("equal styles reported unequal")
	}
	b.Bold = false
	if a.eq(b) {
		t.Error("unequal styles reported equal")
	}
}

func TestStyleIsZero(t *testing.T) {
	var zero Style
	if !zero.isZero() {
		t.Error("zero Style should be isZero")
	}
	if S(Red).isZero() {
		t.Error("non-zero Style should not be isZero")
	}
	if (Style{Bold: true}).isZero() {
		t.Error("bold Style should not be isZero")
	}
}

func TestSeqBold(t *testing.T) {
	seq := Level16.seq(Style{Bold: true, Fg: Red})
	if !strings.Contains(seq, "1;") {
		t.Errorf("bold seq missing '1': %q", seq)
	}
}

func TestSeqZeroStyle(t *testing.T) {
	if Level16.seq(Style{}) != "" {
		t.Error("zero style at Level16 should emit nothing")
	}
}

func TestMixWithIndexed(t *testing.T) {
	a := Indexed(9)  // Red
	b := Indexed(12) // Blue
	m := mix(a, b, 0.5)
	if m.kind != colorRGB {
		t.Error("mix of indexed should produce RGB")
	}
}

func TestRGBTo256Grayscale(t *testing.T) {
	// Pure gray near 0
	i := rgbTo256(0, 0, 0)
	if i != 16 {
		t.Errorf("black maps to %d, want 16", i)
	}
	// Pure gray near max
	i = rgbTo256(255, 255, 255)
	if i != 231 {
		t.Errorf("white maps to %d, want 231", i)
	}
}

func TestMaxInt(t *testing.T) {
	if maxInt(3, 5) != 5 || maxInt(5, 3) != 5 {
		t.Error("maxInt wrong")
	}
}

func TestMinInt(t *testing.T) {
	if minInt(3, 5) != 3 || minInt(5, 3) != 3 {
		t.Error("minInt wrong")
	}
}

func TestTruncStrEdgeCases(t *testing.T) {
	if s := truncStr("ab", 0); s != "" {
		t.Errorf("truncStr(s, 0) = %q", s)
	}
	if s := truncStr("ab", 1); s != "…" {
		t.Errorf("truncStr(s, 1) = %q, want …", s)
	}
	if s := truncStr("abcdef", 3); runeLen(s) != 3 {
		t.Errorf("truncStr(s, 3) len = %d", runeLen(s))
	}
}

func TestEllipTruncASCII(t *testing.T) {
	// n <= 3 truncates to the raw substring (no ellipsis)
	if s := ellipTrunc("hello", 3, false); s != "hel" {
		t.Errorf("ellipTrunc ASCII = %q, want %q", s, "hel")
	}
	if s := ellipTrunc("hello", 8, false); s != "hello" {
		t.Errorf("ellipTrunc ASCII no-op = %q", s)
	}
	if s := ellipTrunc("hello world", 10, false); s != "hello w..." {
		t.Errorf("ellipTrunc ASCII ellipsis = %q, want %q", s, "hello w...")
	}
}

func TestTruncASCII(t *testing.T) {
	if truncASCII("abcdef", 10) != "abcdef" {
		t.Error("truncASCII no-op failed")
	}
	if s := truncASCII("abcdef", 3); len(s) > 3 {
		t.Errorf("truncASCII(3) = %q len %d", s, len(s))
	}
	if truncASCII("abcdef", 0) != "" {
		t.Error("truncASCII(0) not empty")
	}
}
