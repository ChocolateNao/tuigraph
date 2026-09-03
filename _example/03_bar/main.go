package main

import (
	"fmt"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(tuichart.WithWidth(60), tuichart.WithNoColor())

	bar := tuichart.NewBar(
		[]string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		tuichart.BarSeries{Name: "sales", Values: []float64{120, 95, 140, 160, 130}},
		tuichart.BarSeries{Name: "returns", Values: []float64{15, 8, 12, 5, 10}, Color: tuichart.Red},
	).
		Title("weekly sales").
		ShowValues(true)

	g.Add(bar)
	fmt.Print(g.Render())
}
