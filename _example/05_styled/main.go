package main

import (
	"fmt"
	"math"
	"time"

	"github.com/ChocolateNao/tuichart"
)

func main() {
	g := tuichart.New(
		tuichart.WithWidth(80),
		tuichart.WithGap(1),
		tuichart.WithPalette(tuichart.Cyan, tuichart.Salmon, tuichart.Lime, tuichart.Orange),
	)

	// Heatmap with custom colors
	heat := tuichart.NewHeat([][]float64{
		{10, 25, 40, 55},
		{15, 30, 50, 65},
		{20, 35, 55, 75},
	}).
		Title("heatmap").
		RowLabels("low", "mid", "high").
		ColLabels("q1", "q2", "q3", "q4").
		Colors(tuichart.Azure, tuichart.BrightRed).
		CellWidth(6).
		ShowValues(true)
	g.Add(heat)

	// Timeline
	base := time.Now().Add(-4 * time.Hour)
	g.Add(tuichart.NewTimeline().
		Title("events").
		Format("15:04").
		Events(
			tuichart.TimelineEvent{At: base, Label: "start"},
			tuichart.TimelineEvent{At: base.Add(1 * time.Hour), Label: "deploy", Detail: "v1.2"},
			tuichart.TimelineEvent{At: base.Add(2*time.Hour + 30*time.Minute), Label: "rollback", Side: tuichart.SideAbove},
		))

	// Candlestick
	cs := tuichart.NewCandlestick().Title("OHLC").Format("Jan 02")
	price := 100.0
	day := time.Now().Add(-10 * 24 * time.Hour)
	for i := 0; i < 8; i++ {
		o := price
		c := o + (math.Sin(float64(i))*3 + 1)
		h := math.Max(o, c) + 1.5
		l := math.Min(o, c) - 1.5
		cs.Candle(day.Add(time.Duration(i)*24*time.Hour), o, h, l, c)
		price = c
	}
	g.Add(cs)

	fmt.Print(g.Render())
}
