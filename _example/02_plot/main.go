package main

import (
	"fmt"
	"math"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(tuichart.WithWidth(72), tuichart.WithNoColor())

	pts := make([]tuichart.Point, 60)
	for i := range pts {
		x := float64(i)
		pts[i] = tuichart.Point{X: x, Y: math.Sin(x/9)*40 + 50}
	}

	p := tuichart.NewPlot().
		Title("sine wave").
		XLabel("x").YLabel("y").
		Add(tuichart.NewLine("sin", pts...).Color(tuichart.Cyan))

	g.Add(p)
	fmt.Print(g.Render())
}
