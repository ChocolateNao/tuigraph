package main

import (
	"fmt"
	"math"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(tuichart.WithWidth(80), tuichart.WithNoColor())

	// Line plot
	pts := make([]tuichart.Point, 40)
	for i := range pts {
		x := float64(i)
		pts[i] = tuichart.Point{X: x, Y: 30 + 20*math.Sin(x/6)}
	}
	plot := tuichart.NewPlot().
		Title("cpu usage").
		Add(tuichart.NewLine("cpu", pts...).Color(tuichart.Cyan))
	g.Add(plot)

	// Bar chart
	g.Add(tuichart.NewBar(
		[]string{"a", "b", "c", "d"},
		tuichart.BarSeries{Name: "v", Values: []float64{40, 70, 55, 85}},
	).Title("tasks"))

	// Pie chart
	g.Add(tuichart.NewPie().
		Title("disk").
		Slice("root", 45).
		Slice("home", 30).
		Slice("var", 15).
		Slice("tmp", 10))

	// Side-by-side gauges
	g.Row(
		tuichart.NewGauge(72, 100).Title("cpu"),
		tuichart.NewGauge(45, 100).Style(tuichart.GaugeBrackets).Title("mem"),
	)

	fmt.Print(g.Render())
}
