package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	rng := rand.New(rand.NewSource(42))

	g := tuichart.New()

	var cpu, mem []tuichart.Point
	for i := 0; i < 60; i++ {
		x := float64(i)
		cpu = append(cpu, tuichart.Point{
			X: x,
			Y: 40 + 25*math.Sin(x/9) + rng.Float64()*8,
		})
		mem = append(mem, tuichart.Point{
			X: x,
			Y: 55 + 10*math.Cos(x/14) + rng.Float64()*4,
		})
	}
	scatter := make([]tuichart.Point, 0, 30)
	for i := 0; i < 30; i++ {
		x := rng.Float64() * 60
		scatter = append(scatter, tuichart.Point{X: x, Y: 20 + x*0.5 + rng.Float64()*15})
	}

	plot := tuichart.NewPlot().
		Title("system load").
		XLabel("seconds").YLabel("%").
		Add(
			tuichart.NewLine("cpu", cpu...).Color(tuichart.Salmon),
			tuichart.NewLine("mem", mem...).Color(tuichart.DodgerBlue),
			tuichart.NewScatter("samples", scatter...).Color(tuichart.Yellow).Marker('×'),
		)
	g.Add(plot)

	fn := tuichart.NewFunction(func(x float64) float64 { return math.Sin(x) * math.Exp(-x*x/50) }).
		Domain(-8*math.Pi, 8*math.Pi).
		Name("damped sin(x)")
	fn.Title("function plot")
	g.Add(fn)

	bar := tuichart.NewBar(
		[]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
		tuichart.BarSeries{Name: "GET", Values: []float64{120, 95, 140, 160, 130, 80, 60}},
		tuichart.BarSeries{
			Name:   "POST",
			Values: []float64{60, 70, 90, 110, 85, 45, 30},
			Color:  tuichart.Green,
		},
	).Title("requests / day").
		ShowValues(true)
	g.Add(bar)

	histData := make([]float64, 400)
	for i := range histData {
		histData[i] = rng.NormFloat64()*15 + 50
	}
	hist := tuichart.NewHistogram(histData).
		Bins(14).
		Color(tuichart.Cyan).
		Title("latency distribution (ms)").
		ShowValues(true)
	hist.SetXFormatter(func(f float64) string { return fmt.Sprintf("%.0fms", f) })
	g.Add(hist)

	pie := tuichart.NewPie().
		Title("disk usage").
		Slice("home", 42).
		Slice("usr", 25).
		Slice("var", 13).
		Slice("tmp", 8).
		Slice("other", 12).
		ShowValues(true)
	g.Add(pie)

	heat := tuichart.NewHeat(randomGrid(6, 4, rng)).
		Title("activity heatmap").
		RowLabels("node-a", "node-b", "node-c", "node-d", "node-e", "node-f")
	g.Add(heat)

	smallHeat := tuichart.NewHeat([][]float64{{12, 48}, {85, 30}, {60, 74}}).
		Title("cell width + values").
		CellWidth(8).
		RowLabels("1-row", "2-row", "3-row").
		ColLabels("1-col", "2-col").
		ShowValues(true)
	g.Add(smallHeat)

	sparks := tuichart.NewSpark()
	for i := 0; i < 78; i++ {
		sparks.Values(20 + 15*math.Sin(float64(i)/4) + rng.Float64()*10)
	}
	sparks.Title("network throughput")
	g.Add(sparks)

	base := time.Now().Add(-3 * time.Hour)
	tl := tuichart.NewTimeline().
		Title("deploy timeline").
		Events(
			tuichart.TimelineEvent{
				At:     base.Add(0 * time.Minute),
				Label:  "v2.0 cut",
				Detail: "tag v2.0.0 checking lenth look at how long the message is",
			},
			tuichart.TimelineEvent{At: base.Add(45 * time.Minute), Label: "unit tests"},
			tuichart.TimelineEvent{At: base.Add(80 * time.Minute), Label: "", Detail: "auto merge"},
			tuichart.TimelineEvent{
				At:     base.Add(110 * time.Minute),
				Label:  "canary 5%",
				Detail: "error rate 0.2%",
				Side:   tuichart.SideAbove,
			},
			tuichart.TimelineEvent{At: base.Add(160 * time.Minute), Label: "full rollout"},
		)
	tl.Format("15:04")
	g.Add(tl)

	gt := tuichart.NewGantt().Title("sprint board").
		Bar("research", base, base.Add(36*time.Hour)).
		Duration("api", base.Add(12*time.Hour), 60*time.Hour).
		Task(
			tuichart.GanttBar{
				Name:  "ui",
				Start: base.Add(24 * time.Hour),
				End:   base.Add(72 * time.Hour),
			},
		).
		Bar("release prep", base.Add(60*time.Hour), base.Add(84*time.Hour))
	gt.Format("Jan 02 15:04")
	g.Add(gt)

	ts := tuichart.NewTimeSeries().Title("server load — 48h")
	load := ts.Line("load %")
	for i := 0; i <= 96; i++ {
		wob := math.Sin(float64(i)/6)*25 + float64(rng.Intn(9))
		load.Add(base.Add(time.Duration(i)*30*time.Minute), 45+wob)
	}
	ts.Line("capacity").Add(base, 80).Add(base.Add(48*time.Hour), 80).Color(tuichart.Red)
	g.Add(ts)

	g.Row(
		tuichart.NewGauge(62, 100).Title("cpu").Label("4 cores"),
		tuichart.NewGauge(72, 100).Style(tuichart.GaugeBrackets).Title("memory"),
		tuichart.NewGauge(34, 100).Style(tuichart.GaugeArrow).Title("disk"),
	)
	g.Row(
		tuichart.NewGauge(8, 10).
			Style(tuichart.GaugeSegments).
			Title("quota").
			Color(tuichart.Green).
			ShowPercent(false),
		tuichart.NewGauge(5, 10).Style(tuichart.GaugeASCII).Title("legacy panel"),
	)

	swapped := tuichart.NewBarValues(
		[]string{"mon", "tue", "wed", "thu", "fri"},
		[]float64{120, 95, 140, 160, 130},
	).
		Title("requests — horizontal (axis switch)").
		Orientation(tuichart.OrientHorizontal)
	g.Add(swapped)

	candles := tuichart.NewCandlestick().Title("ACME — daily OHLC")
	price := 142.0
	day := time.Now().Add(-30 * 24 * time.Hour)
	for i := 0; i < 26; i++ {
		o := price
		cl := o + rng.Float64()*7 - 3.4
		hi := math.Max(o, cl) + rng.Float64()*2.5
		lo := math.Min(o, cl) - rng.Float64()*2.5
		r := func(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
		candles.Candle(day.Add(time.Duration(i)*24*time.Hour), r(o), r(hi), r(lo), r(cl))
		price = cl
	}
	candles.Format("Jan 02")
	g.Add(candles)

	fmt.Print(g.Render())
	fmt.Println("\nsparkline:", tuichart.Spark(sparkVals()))
}

func randomGrid(rows, cols int, rng *rand.Rand) [][]float64 {
	grid := make([][]float64, rows)
	for r := range grid {
		grid[r] = make([]float64, cols)
		for c := range grid[r] {
			grid[r][c] = math.Abs(math.Sin(float64(c)/3+float64(r)/2)) * 100 * rng.Float64()
		}
	}
	return grid
}

func sparkVals() []float64 {
	out := make([]float64, 40)
	for i := range out {
		out[i] = 10 + 8*math.Sin(float64(i)/3)
	}
	return out
}
