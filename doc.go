// Package tuichart renders charts, plots and other diagrams as ANSI-styled text
// that works in virtually any terminal.
//
// A Chart collects multiple diagrams (plots, bars, pies, heatmaps, ...),
// lays them out in rows and renders them to a string sized for the current
// terminal. Colors degrade automatically (truecolor -> 256 -> 16 -> none)
// and Unicode glyphs fall back to ASCII on dumb terminals.
//
// Quick start:
//
//	g := tuichart.New()
//	g.Add(tuichart.NewPlot().
//	    Title("CPU").
//	    XLabel("time").YLabel("%").
//	    Add(tuichart.NewLine("user", pts...).Color(tuichart.Cyan)))
//	fmt.Println(g.Render())
//
// Every diagram implements the Drawable interface, so custom diagram types
// plug into the same container, layout and styling machinery.
package tuichart
