package tuichart

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	seqEnterAlt   = "\x1b[?1049h"
	seqExitAlt    = "\x1b[?1049l"
	seqHideCursor = "\x1b[?25l"
	seqShowCursor = "\x1b[?25h"
	seqHome       = "\x1b[H"
	seqClearBelow = "\x1b[J"
)

const defaultInterval = 200 * time.Millisecond

// Live re-renders a Chart on a fixed interval, painting each frame over the
// previous one in the terminal's alternate screen buffer. Feed it data from
// your own goroutines and mutate diagrams either inside OnUpdate (invoked by
// the render loop) or via Update (mutate + immediate repaint).
//
// All of Live's methods are safe for concurrent use.
type Live struct {
	out      io.Writer
	chart    *Chart
	update   func()
	stop     chan struct{}
	done     chan struct{}
	interval time.Duration
	stopOnce sync.Once
	mu       sync.Mutex
	running  bool
}

// LiveOption configures a Live renderer.
type LiveOption func(*Live)

// WithInterval sets the time between frames. Values below 10ms are clamped.
func WithInterval(d time.Duration) LiveOption {
	return func(l *Live) { l.interval = d }
}

// WithFPS sets the frame rate; interval becomes 1/fps.
func WithFPS(fps float64) LiveOption {
	return func(l *Live) {
		if fps > 0 {
			l.interval = time.Duration(float64(time.Second) / fps)
		}
	}
}

// WithLiveOutput redirects frames; defaults to os.Stdout. Any io.Writer
// works, which makes tests deterministic (escape sequences land in the
// buffer).
func WithLiveOutput(w io.Writer) LiveOption {
	return func(l *Live) { l.out = w }
}

// OnUpdate registers a data-pump callback invoked under the render lock
// before every frame. Mutate the chart's diagrams there. Do not call Live
// methods from inside it.
func OnUpdate(fn func()) LiveOption {
	return func(l *Live) { l.update = fn }
}

// NewLive wraps a chart for continuous rendering.
func NewLive(c *Chart, opts ...LiveOption) *Live {
	l := &Live{
		chart:    c,
		interval: defaultInterval,
		out:      os.Stdout,
	}
	for _, o := range opts {
		o(l)
	}
	if l.interval < 10*time.Millisecond {
		l.interval = 10 * time.Millisecond
	}
	return l
}

// Frame renders one frame at the given width without touching the screen.
// It does not invoke the OnUpdate callback.
func (l *Live) Frame(width int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.chart.Render(width)
}

// Update runs fn while holding the render lock, then repaints immediately.
func (l *Live) Update(fn func()) {
	l.mu.Lock()
	fn()
	s := l.renderLocked(0)
	l.mu.Unlock()
	l.write(seqHome + s + "\n" + seqClearBelow)
}

// Repaint renders and paints a frame right away.
func (l *Live) Repaint() {
	l.mu.Lock()
	s := l.renderLocked(0)
	l.mu.Unlock()
	l.write(seqHome + s + "\n" + seqClearBelow)
}

// Run drives the render loop until ctx is canceled or Stop is called. It
// enters the alternate screen, repaints every interval, and restores the
// terminal on exit. Returns nil when stopped via Stop, or ctx.Err().
func (l *Live) Run(ctx context.Context) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return errors.New("tuichart: live already running")
	}
	l.running = true
	l.stop = make(chan struct{})
	l.stopOnce = sync.Once{}
	l.done = make(chan struct{})
	stop := l.stop
	done := l.done
	l.mu.Unlock()

	defer close(done)
	defer func() {
		l.mu.Lock()
		l.running = false
		l.stop = nil
		l.mu.Unlock()
	}()
	defer fmt.Fprint(l.out, seqShowCursor+seqExitAlt)
	fmt.Fprint(l.out, seqEnterAlt+seqHideCursor)

	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	l.Repaint()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stop:
			return nil
		case <-ticker.C:
			l.mu.Lock()
			if l.update != nil {
				l.update()
			}
			s := l.renderLocked(0)
			l.mu.Unlock()
			l.write(seqHome + s + "\n" + seqClearBelow)
		}
	}
}

// Stop asks Run to return; safe to call multiple times and from any
// goroutine.
func (l *Live) Stop() {
	l.mu.Lock()
	stop := l.stop
	l.mu.Unlock()
	if stop != nil {
		l.stopOnce.Do(func() { close(stop) })
	}
}

// Done returns a channel closed once Run has exited.
func (l *Live) Done() <-chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.done
}

// renderLocked renders the next frame; caller holds mu. Width comes from
// the chart's fixed option if set, otherwise the current terminal width is
// detected fresh each frame so window resizes are picked up.
func (l *Live) renderLocked(width int) string {
	w := width
	if w <= 0 {
		w = l.chart.opts.width
		if w <= 0 {
			w = DetectWriter(l.out).W
		}
	}
	return l.chart.Render(w)
}

func (l *Live) write(s string) {
	io.WriteString(l.out, s)
}
