package tuichart

import (
	"bytes"
	"context"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func liveChart() (*Chart, *Line) {
	g := New(WithNoColor(), WithUnicode(true))
	line := NewLineVals("feed", []float64{1, 2, 1})
	p := NewPlot().Title("live")
	p.Add(line)
	g.Add(p)
	return g, line
}

func TestLiveFrameDeterministic(t *testing.T) {
	g, _ := liveChart()
	l := NewLive(g, WithLiveOutput(&bytes.Buffer{}))
	a := l.Frame(50)
	b := l.Frame(50)
	if a == "" || a != b {
		t.Error("Frame not deterministic or empty")
	}
	if !strings.Contains(a, "┌") {
		t.Error("frame missing border")
	}
}

func TestLiveRunPaintsFramesAndRestores(t *testing.T) {
	g, _ := liveChart()
	var buf bytes.Buffer
	l := NewLive(g,
		WithLiveOutput(&buf),
		WithInterval(10*time.Millisecond),
	)
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- l.Run(ctx) }()
	time.Sleep(40 * time.Millisecond)
	cancel()
	if err := <-errc; err != context.Canceled {
		t.Errorf("Run err = %v", err)
	}
	out := buf.String()
	for _, want := range []string{seqEnterAlt, seqHideCursor, "live", seqShowCursor, seqExitAlt} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
	if n := strings.Count(out, seqHome); n < 2 {
		t.Errorf("expected >=2 repaints, got %d", n)
	}
}

func TestLiveStopReturnsNil(t *testing.T) {
	g, _ := liveChart()
	var buf bytes.Buffer
	l := NewLive(g, WithLiveOutput(&buf), WithInterval(5*time.Millisecond))
	errc := make(chan error, 1)
	go func() { errc <- l.Run(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	l.Stop()
	select {
	case err := <-errc:
		if err != nil {
			t.Errorf("Stop run err = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after Stop")
	}
	<-l.Done()
}

func TestOnUpdateRunsPerTick(t *testing.T) {
	g, line := liveChart()
	var n int
	var mu sync.Mutex
	var buf bytes.Buffer
	l := NewLive(g,
		WithLiveOutput(&buf),
		WithInterval(10*time.Millisecond),
		OnUpdate(func() {
			mu.Lock()
			n++
			line.Values(float64(n))
			mu.Unlock()
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	go l.Run(ctx)
	time.Sleep(60 * time.Millisecond)
	cancel()
	<-l.Done()

	mu.Lock()
	defer mu.Unlock()
	if n < 2 {
		t.Errorf("update ran %d times", n)
	}
}

func TestLiveUpdateImmediate(t *testing.T) {
	g, _ := liveChart()
	var buf bytes.Buffer
	l := NewLive(g, WithLiveOutput(&buf))
	l.Update(func() { g.Title("mutated") })
	if !strings.Contains(buf.String(), "mutated") {
		t.Error("Update did not repaint with mutation")
	}
}

func TestLiveDoubleStopAndRerun(t *testing.T) {
	g, _ := liveChart()
	l := NewLive(g, WithLiveOutput(&bytes.Buffer{}), WithInterval(5*time.Millisecond))
	go l.Run(context.Background())
	time.Sleep(15 * time.Millisecond)
	l.Stop()
	l.Stop()
	<-l.Done()

	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- l.Run(ctx) }()
	time.Sleep(15 * time.Millisecond)
	cancel()
	if err := <-errc; err != context.Canceled {
		t.Errorf("second Run err = %v", err)
	}
}

func TestRingWrapAround(t *testing.T) {
	r := NewRing(3)
	r.Push(1, 2, 3, 4, 5)
	if r.Len() != 3 || r.Cap() != 3 {
		t.Fatalf("len=%d cap=%d", r.Len(), r.Cap())
	}
	got := r.Values()
	want := []float64{3, 4, 5}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestRingPartialFill(t *testing.T) {
	r := NewRing(5)
	r.Push(7, 8)
	if got := r.Values(); len(got) != 2 || got[0] != 7 || got[1] != 8 {
		t.Errorf("partial: %v", got)
	}
}

func TestRingConcurrent(t *testing.T) {
	r := NewRing(64)
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		rng := rand.New(rand.NewSource(1))
		for {
			select {
			case <-stop:
				return
			default:
				r.Push(rng.NormFloat64())
			}
		}
	}()
	line := NewLine("live")
	for i := 0; i < 100; i++ {
		line.SetValues(r.Values())
		time.Sleep(time.Millisecond)
	}
	close(stop)
	wg.Wait()
}

func TestSetValuesReplacesPoints(t *testing.T) {
	l := NewLineVals("x", []float64{1, 2, 3})
	l.SetValues([]float64{9, 8})
	if len(l.pts) != 2 || l.pts[0].Y != 9 || l.pts[1].Y != 8 {
		t.Errorf("pts = %v", l.pts)
	}
}

func TestLiveGaugeUpdates(t *testing.T) {
	g := New(WithWidth(40), WithNoColor(), WithUnicode(true))
	gauge := NewGauge(0, 100).Title("progress")
	g.Add(gauge)

	var buf bytes.Buffer
	var n int
	l := NewLive(g,
		WithLiveOutput(&buf),
		WithInterval(10*time.Millisecond),
		OnUpdate(func() {
			n += 25
			gauge.Value(float64(n))
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	go l.Run(ctx)
	time.Sleep(80 * time.Millisecond)
	cancel()
	<-l.Done()

	out := buf.String()
	if strings.Count(out, seqHome) < 2 {
		t.Fatalf("expected multiple repaints, got %d", strings.Count(out, seqHome))
	}
	for _, want := range []string{"25%", "50%"} {
		if !strings.Contains(out, want) {
			t.Errorf("gauge output missing evolving %s", want)
		}
	}
}

func TestLiveGaugeFrameTracksValue(t *testing.T) {
	g := New(WithWidth(30), WithNoColor(), WithUnicode(true))
	gauge := NewGauge(0, 10).Style(GaugeSegments).ShowPercent(false)
	g.Add(gauge)
	l := NewLive(g, WithLiveOutput(io.Discard))

	f1 := l.Frame(30)
	gauge.Value(10)
	f2 := l.Frame(30)
	if f1 == f2 {
		t.Fatal("frame did not change after value update")
	}
	full := strings.Count(f2, "▰")
	empty := strings.Count(f2, "▱")
	if full == 0 || empty != 0 {
		t.Errorf("full gauge render wrong: filled=%d empty=%d", full, empty)
	}
	if strings.Count(f1, "▰") != 0 {
		t.Errorf("zero gauge has filled segments: %q", f1)
	}
}

func TestLiveGaugeSegmentStretch(t *testing.T) {
	cv := NewCanvas(20, 1)
	gg := NewGauge(5, 10).Style(GaugeSegments).ShowPercent(false)
	gg.Draw(NewRenderCtx(Info{Level: LevelNone, Unicode: true}), cv)
	row := splitLines(cv.Plain())[0]
	row = strings.TrimRight(row, " ")
	if runeLen(row) != 20 {
		t.Errorf("stretched segments span %d cols, want 20: %q", runeLen(row), row)
	}
}

func TestFPSOption(t *testing.T) {
	l := NewLive(New(), WithFPS(10))
	if l.interval != 100*time.Millisecond {
		t.Errorf("interval = %v", l.interval)
	}
}

// sanity: a full live session end-to-end produces a valid plot of ring data
func TestLiveSessionIntegration(t *testing.T) {
	g := New(WithWidth(60), WithNoColor(), WithUnicode(true), WithColor256())
	p := NewPlot().Title("walk")
	line := NewLine("rand")
	p.Add(line)
	g.Add(p)

	r := NewRing(50)
	l := NewLive(g,
		WithLiveOutput(io.Discard),
		WithInterval(10*time.Millisecond),
		OnUpdate(func() {
			line.SetValues(r.Values())
			r.Push(math.Sin(float64(r.Len())) + rand.NormFloat64()*0.1)
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	go l.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-l.Done()
	frame := l.Frame(60)
	if !strings.Contains(frame, "walk") {
		t.Error("final frame missing title")
	}
}
