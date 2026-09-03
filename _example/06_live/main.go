package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(
		tuichart.WithColor256(),
		tuichart.WithUnicode(true),
	)

	// --- Plot: CPU + memory random walk ---
	plot := tuichart.NewPlot().
		Title("system metrics (live)").
		XLabel("seconds").YLabel("%")

	cpuLine := tuichart.NewLine("cpu").Color(tuichart.Salmon)
	memLine := tuichart.NewLine("mem").Color(tuichart.DodgerBlue)
	plot.Add(cpuLine, memLine)
	g.Add(plot)

	// --- Sparkline: network throughput ---
	netSpark := tuichart.NewSpark().
		Title("network I/O").
		Color(tuichart.Cyan)
	g.Add(netSpark)

	// --- Row of gauges ---
	gaugeCPU := tuichart.NewGauge(0, 100).
		Title("cpu").
		Style(tuichart.GaugeBrackets).
		Color(tuichart.Salmon)
	gaugeMem := tuichart.NewGauge(0, 100).
		Title("memory").
		Style(tuichart.GaugeArrow).
		Color(tuichart.DodgerBlue)
	gaugeDisk := tuichart.NewGauge(0, 100).
		Title("disk").
		Style(tuichart.GaugeSegments).
		Color(tuichart.Green)
	g.Row(gaugeCPU, gaugeMem, gaugeDisk)

	// --- Data buffers ---
	cpuRing := tuichart.NewRing(60)
	memRing := tuichart.NewRing(60)
	netVals := make([]float64, 0, 60)

	var cpuAvg, memAvg float64

	lv := tuichart.NewLive(g,
		tuichart.WithInterval(120*time.Millisecond),
		tuichart.OnUpdate(func() {
			// CPU random walk
			cpuLast := 50.0
			if v := cpuRing.Values(); len(v) > 0 {
				cpuLast = v[len(v)-1]
			}
			cpuNow := math.Min(95, math.Max(5, cpuLast+(rand.Float64()-0.5)*10))
			cpuRing.Push(cpuNow)
			cpuLine.SetValues(cpuRing.Values())

			// Memory random walk
			memLast := 55.0
			if v := memRing.Values(); len(v) > 0 {
				memLast = v[len(v)-1]
			}
			memNow := math.Min(90, math.Max(10, memLast+(rand.Float64()-0.5)*6))
			memRing.Push(memNow)
			memLine.SetValues(memRing.Values())

			// Network sparkline
			netNow := 20 + 15*math.Sin(float64(len(netVals))/4) + rand.Float64()*12
			netVals = append(netVals, netNow)
			if len(netVals) > 60 {
				netVals = netVals[1:]
			}
			netSpark.SetValues(netVals)

			// Gauges
			cpuAvg = cpuAvg*0.8 + cpuNow*0.2
			memAvg = memAvg*0.8 + memNow*0.2
			diskNow := 30 + 10*math.Sin(float64(len(netVals))/20)
			gaugeCPU.Value(cpuAvg)
			gaugeMem.Value(memAvg)
			gaugeDisk.Value(diskNow)
		}),
	)

	fmt.Fprintln(os.Stderr, "live multi-diagram demo — Ctrl+C to exit")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()
	lv.Run(ctx)
}
