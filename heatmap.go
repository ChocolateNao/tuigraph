package tuichart

import "math"

type Heatmap struct {
	grid      [][]float64
	rowLabels []string
	colLabels []string
	chartBase
	low  Color
	high Color
}

func NewHeat(grid [][]float64) *Heatmap {
	return &Heatmap{
		chartBase: newChartBase(),
		grid:      grid,
		low:       Indexed(21),
		high:      Indexed(196),
	}
}

// Title sets the chart title.
func (h *Heatmap) Title(t string) *Heatmap { h.SetTitle(t); return h }

func (h *Heatmap) Colors(low, high Color) *Heatmap { h.low, h.high = low, high; return h }
func (h *Heatmap) RowLabels(ls ...string) *Heatmap { h.rowLabels = ls; return h }
func (h *Heatmap) ColLabels(ls ...string) *Heatmap { h.colLabels = ls; return h }

// CellWidth fixes each cell's width in columns instead of stretching cells
// to fill the frame. Useful when few columns would otherwise produce huge
// solid blocks.
func (h *Heatmap) CellWidth(n int) *Heatmap { h.SetCellWidth(n); return h }

// ShowValues prints the numeric value inside each cell (when it fits).
func (h *Heatmap) ShowValues(v bool) *Heatmap { h.SetShowValues(v); return h }

func (h *Heatmap) HeightHint(width int) int {
	if h := h.chartBase.HeightHint(width); h > 0 {
		return h
	}
	rows := len(h.grid)
	if rows < 4 {
		rows = 4
	}
	if rows > 30 {
		rows = 30
	}
	hh := rows + 3 // grid + colorbar + frame
	if len(h.colLabels) > 0 {
		hh++ // column title row above the grid
	}
	return hh
}

var heatRampASCII = []rune{' ', '.', ':', '-', '=', '+', '*', '#', '%', '@'}

func (h *Heatmap) Draw(rc *Ctx, cv *Canvas) {
	inner := h.frameTitle(cv, rc.Info.Unicode)
	uni := rc.Info.Unicode
	mono := rc.Info.Level == LevelNone

	lo, hi := math.Inf(1), math.Inf(-1)
	nCols := 0
	for _, row := range h.grid {
		if len(row) > nCols {
			nCols = len(row)
		}
		for _, v := range row {
			if math.IsNaN(v) {
				continue
			}
			lo = math.Min(lo, v)
			hi = math.Max(hi, v)
		}
	}
	if math.IsInf(lo, 1) || nCols == 0 || inner.W < 2 || inner.H < 2 {
		cv.TextCenter(cv.Width()/2, cv.Height()/2, "(no data)", S(Gray))
		return
	}
	if hi == lo {
		hi = lo + 1
	}

	labelGutter := 0
	for _, l := range h.rowLabels {
		if n := runeLen(l); n+1 > labelGutter && n*2 < inner.W {
			labelGutter = n + 1
		}
	}
	cellW := inner.W - labelGutter
	// Reserve the top inner row for column titles when present, and the
	// bottom row for the color scale.
	gridTop := inner.Y
	cellH := inner.H - 1
	if len(h.colLabels) > 0 {
		gridTop++
		cellH--
	}
	cw := cellW / nCols
	if h.cellW > 0 {
		// Fixed cell width; keep the grid inside the frame.
		cw = h.cellW
		if cw > cellW {
			cw = cellW
		}
	}
	if cw < 1 {
		cw = 1
	}
	chh := cellH / len(h.grid)
	if chh < 1 {
		chh = 1
	}

	if len(h.colLabels) > 0 {
		for c := 0; c < nCols; c++ {
			lbl := ""
			if c < len(h.colLabels) {
				lbl = h.colLabels[c]
			}
			cv.TextCenter(
				inner.X+labelGutter+c*cw+cw/2,
				gridTop-1,
				ellipTrunc(lbl, cw, uni),
				S(Silver),
			)
		}
	}

	// The color scale always occupies the last inner row; the grid must
	// stay above it even when the frame is shorter than the hint.
	barY := inner.Y2()
	gridBottom := barY - 1
	for ri, row := range h.grid {
		y := gridTop + ri*chh
		if y > gridBottom {
			break
		}
		if labelGutter > 0 && ri < len(h.rowLabels) {
			cv.TextRight(
				inner.X+labelGutter-2,
				y,
				ellipTrunc(h.rowLabels[ri], labelGutter-1, uni),
				S(Default),
			)
		}
		for ci, v := range row {
			if math.IsNaN(v) {
				continue
			}
			t := (v - lo) / (hi - lo)
			st := S(mix(h.low, h.high, t))
			ch := '█'
			if mono || !uni {
				idx := int(t*float64(len(heatRampASCII)-1) + 0.5)
				if idx < 0 {
					idx = 0
				}
				if idx >= len(heatRampASCII) {
					idx = len(heatRampASCII) - 1
				}
				ch = heatRampASCII[idx]
			}
			for yy := 0; yy < chh && y+yy <= gridBottom; yy++ {
				for xx := 0; xx < cw && ci*cw+xx < cellW; xx++ {
					cv.Set(inner.X+labelGutter+ci*cw+xx, y+yy, ch, st)
				}
			}
			if h.showVals {
				lbl := FormatValue(v)
				if runeLen(lbl) <= cw {
					cv.TextCenter(inner.X+labelGutter+ci*cw+cw/2, y, lbl, S(Default))
				}
			}
		}
	}

	barW := minInt(inner.W-8, 24)
	if barW > 2 {
		y := barY
		if y > gridTop {
			for i := 0; i < barW; i++ {
				t := float64(i) / float64(barW-1)
				ch := '█'
				if mono {
					idx := int(t*float64(len(heatRampASCII)-1) + 0.5)
					ch = heatRampASCII[idx]
				}
				cv.Set(inner.X+i, y, ch, S(mix(h.low, h.high, t)))
			}
			lblLo := FormatValue(lo)
			lblHi := FormatValue(hi)
			cv.Text(inner.X+barW+1, y, lblLo+".."+lblHi, S(Gray))
		}
	}
}
