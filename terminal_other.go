//go:build !unix && !windows

package tuichart

import "os"

func platformIsTTY(f *os.File) bool {
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func termSize(f *os.File) (int, int, bool) {
	return DefaultWidth, DefaultHeight, false
}
