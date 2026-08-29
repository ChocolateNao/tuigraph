//go:build unix

package graph

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row, Col, Xpixel, Ypixel uint16
}

func tiocgwinsz() uintptr {
	switch runtime.GOOS {
	case "linux":
		return 0x5413
	default:
		return 0x40087468
	}
}

func ioctl(fd, req uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(unsafe.Pointer(&winsize{})))
	if errno != 0 {
		return errno
	}
	return nil
}

func platformIsTTY(f *os.File) bool {
	return ioctl(f.Fd(), tiocgwinsz()) == nil
}

func termSize(f *os.File) (int, int, bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), tiocgwinsz(), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return DefaultWidth, DefaultHeight, false
	}
	w, h := int(ws.Col), int(ws.Row)
	if w < 1 {
		w = DefaultWidth
	}
	if h < 1 {
		h = DefaultHeight
	}
	return w, h, true
}
