//go:build windows

package graph

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle = kernel32.NewProc("GetStdHandle")
	procGetCSBI      = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type coord struct{ X, Y int16 }

type smallRect struct{ Left, Top, Right, Bottom int16 }

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

func stdoutHandle() uintptr {
	h, _, _ := procGetStdHandle.Call(^uintptr(10))
	return h
}

func platformIsTTY(f *os.File) bool {
	if f == os.Stdout || f == os.Stderr || f == os.Stdin {
		var info consoleScreenBufferInfo
		r, _, _ := procGetCSBI.Call(stdoutHandle(), uintptr(unsafe.Pointer(&info)))
		return r != 0
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func termSize(f *os.File) (int, int, bool) {
	if f != os.Stdout && f != os.Stderr && f != os.Stdin {
		return DefaultWidth, DefaultHeight, false
	}
	var info consoleScreenBufferInfo
	r, _, _ := procGetCSBI.Call(stdoutHandle(), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return DefaultWidth, DefaultHeight, false
	}
	w := int(info.Window.Right-info.Window.Left) + 1
	h := int(info.Window.Bottom-info.Window.Top) + 1
	if w < 1 || h < 1 {
		return DefaultWidth, DefaultHeight, false
	}
	return w, h, true
}
