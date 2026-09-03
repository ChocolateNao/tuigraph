package tuichart

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var _ io.WriterTo = (*Chart)(nil)

func sampleRenderChart() *Chart {
	g := New(WithNoColor(), WithUnicode(false), WithWidth(40))
	g.Title("t")
	g.Add(NewSpark(1, 2, 3).Title("s"))
	return g
}

func TestRenderToBufferMatchesRender(t *testing.T) {
	c := sampleRenderChart()
	var buf bytes.Buffer
	if err := c.RenderTo(&buf); err != nil {
		t.Fatalf("RenderTo: %v", err)
	}
	if buf.String() != c.Render() {
		t.Error("RenderTo output != Render output")
	}
}

func TestRenderToExplicitWidth(t *testing.T) {
	c := sampleRenderChart()
	var buf bytes.Buffer
	if err := c.RenderTo(&buf, 60); err != nil {
		t.Fatalf("RenderTo: %v", err)
	}
	if buf.String() != c.Render(60) {
		t.Error("RenderTo(w, 60) != Render(60)")
	}
	for _, line := range splitLines(buf.String()) {
		if runeLen(line) > 60 {
			t.Errorf("line exceeds width 60: %q", line)
		}
	}
}

func TestRenderToFile(t *testing.T) {
	c := sampleRenderChart()
	path := filepath.Join(t.TempDir(), "chart.txt")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.RenderTo(f); err != nil {
		t.Fatalf("RenderTo(file): %v", err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != c.Render() {
		t.Error("file content != Render output")
	}
}

type failWriter struct{ after int }

func (fw *failWriter) Write(p []byte) (int, error) {
	if fw.after <= 0 {
		return 0, errors.New("boom")
	}
	fw.after--
	if len(p) > 3 {
		return 0, errors.New("boom")
	}
	return len(p), nil
}

func TestRenderToWriterError(t *testing.T) {
	c := sampleRenderChart()
	err := c.RenderTo(&failWriter{})
	if err == nil {
		t.Fatal("expected error from failing writer")
	}
	sentinel := errors.New("boom")
	if !errors.Is(err, sentinel) && !strings.Contains(err.Error(), "boom") {
		t.Errorf("error not wrapped with cause: %v", err)
	}
	if !strings.Contains(err.Error(), "tuichart") {
		t.Errorf("error missing library context: %v", err)
	}
}

func TestChartWriteToViaIoCopy(t *testing.T) {
	c := sampleRenderChart()
	var buf bytes.Buffer
	n, err := io.Copy(&buf, c.Reader())
	if err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	if n != int64(buf.Len()) || buf.String() != c.Render() {
		t.Error("io.Copy output mismatch")
	}
}

func TestShortWriteDetected(t *testing.T) {
	c := sampleRenderChart()
	err := c.RenderTo(shortWriter{})
	if err == nil || !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("short write not detected: %v", err)
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
