package tuichart

import "sync"

// Ring is a fixed-capacity rolling buffer of samples, safe for concurrent
// use. It is the standard data source for live charts: producers call Push
// from their own goroutine while the render loop reads Values.
type Ring struct {
	buf []float64
	n   int
	i   int
	mu  sync.Mutex
}

// NewRing creates a ring holding up to capacity samples.
func NewRing(capacity int) *Ring {
	if capacity < 1 {
		capacity = 1
	}
	return &Ring{buf: make([]float64, capacity)}
}

// Push appends samples, overwriting the oldest once full.
func (r *Ring) Push(vs ...float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range vs {
		r.buf[r.i] = v
		r.i = (r.i + 1) % len(r.buf)
		if r.n < len(r.buf) {
			r.n++
		}
	}
}

// Len reports how many samples are currently stored.
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n
}

// Cap reports the capacity.
func (r *Ring) Cap() int { return len(r.buf) }

// Values returns a copy of the stored samples, oldest first.
func (r *Ring) Values() []float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]float64, r.n)
	if r.n == 0 {
		return out
	}
	start := 0
	if r.n == len(r.buf) {
		start = r.i
	}
	copy(out, r.buf[start:])
	if start+r.n > len(r.buf) {
		copy(out[len(r.buf)-start:], r.buf[:r.i])
	}
	return out
}
