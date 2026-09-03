package tuichart

import (
	"io"
	"os"
	"strings"
	"sync"
)

const (
	DefaultWidth  = 80
	DefaultHeight = 24
)

type Info struct {
	W, H    int
	Level   Level
	Unicode bool
}

var (
	detectMu        sync.Mutex
	profileOverride = Level(-1)
	unicodeOverride = int8(-1)
	cachedEnvLvl    = Level(-1)
)

func SetProfile(l Level) {
	detectMu.Lock()
	profileOverride = l
	detectMu.Unlock()
}

func SetUnicode(on bool) {
	detectMu.Lock()
	if on {
		unicodeOverride = 1
	} else {
		unicodeOverride = 0
	}
	detectMu.Unlock()
}

func ResetDetection() {
	detectMu.Lock()
	profileOverride = -1
	unicodeOverride = -1
	cachedEnvLvl = -1
	detectMu.Unlock()
}

func Detect() Info {
	return DetectWriter(os.Stdout)
}

func DetectWriter(w io.Writer) Info {
	detectMu.Lock()
	defer detectMu.Unlock()
	info := Info{W: DefaultWidth, H: DefaultHeight, Unicode: localeSupportsUnicode()}
	if f, ok := w.(*os.File); ok {
		if platformIsTTY(f) {
			width, height, ok2 := termSize(f)
			if ok2 {
				info.W, info.H = width, height
				info.Level = envLevel()
			}
		} else {
			info.Level = LevelNone
		}
	}
	return applyOverrides(info)
}

func applyOverrides(info Info) Info {
	if profileOverride >= 0 {
		info.Level = profileOverride
	}
	if unicodeOverride >= 0 {
		info.Unicode = unicodeOverride == 1
	}
	return info
}

func envLevel() Level {
	if cachedEnvLvl >= 0 {
		return cachedEnvLvl
	}
	lvl := detectEnvLevel()
	cachedEnvLvl = lvl
	return lvl
}

func detectEnvLevel() Level {
	if v := os.Getenv("NO_COLOR"); v != "" {
		return LevelNone
	}
	term := os.Getenv("TERM")
	switch term {
	case "dumb", "":
		return LevelNone
	}
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit":
		return LevelTrue
	}
	if os.Getenv("WT_SESSION") != "" {
		return LevelTrue
	}
	if strings.Contains(term, "256color") || strings.Contains(term, "xterm-kitty") ||
		strings.Contains(term, "alacritty") {
		return Level256
	}
	return Level16
}

func localeSupportsUnicode() bool {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v := os.Getenv(k)
		if v == "" {
			continue
		}
		v = strings.ToLower(v)
		return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
	}
	return true
}
