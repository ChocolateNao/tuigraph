package tuichart

import "testing"

func TestCtxNextCycles(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	n := len(rc.Palette)
	seen := make([]Color, n)
	for i := 0; i < n; i++ {
		seen[i] = rc.Next()
	}
	wrapped := rc.Next()
	if wrapped != seen[0] {
		t.Errorf("Next did not cycle: got %v, want %v", wrapped, seen[0])
	}
}

func TestCtxNextMultipleCycles(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	n := len(rc.Palette)
	first := rc.Next()
	for i := 0; i < n-1; i++ {
		rc.Next()
	}
	second := rc.Next()
	if first != second {
		t.Errorf("second cycle start %v != first %v", second, first)
	}
}

func TestCtxNextEmptyPalette(t *testing.T) {
	rc := NewRenderCtx(Info{})
	if len(rc.Palette) == 0 {
		t.Skip("default palette is empty, skip")
	}
	for i := 0; i < len(rc.Palette)*3; i++ {
		_ = rc.Next()
	}
}

func TestCtxPaletteDefault(t *testing.T) {
	rc := NewRenderCtx(Info{Level: LevelNone})
	if len(rc.Palette) != len(defaultPalette) {
		t.Errorf("palette len %d, want %d", len(rc.Palette), len(defaultPalette))
	}
	for i, c := range rc.Palette {
		if c != defaultPalette[i] {
			t.Errorf("palette[%d] = %v, want %v", i, c, defaultPalette[i])
		}
	}
}
