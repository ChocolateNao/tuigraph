package tuichart

const brailleBase = 0x2800

var brailleBits = [4][2]byte{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

func setDot(cv *Canvas, gx, gy int, c Color) {
	cx, cy := gx>>1, gy>>2
	cur := cv.At(cx, cy)
	var bits byte
	if cur.ch >= brailleBase && cur.ch <= brailleBase+0xff {
		bits = byte(cur.ch - brailleBase)
	}
	bits |= brailleBits[gy&3][gx&1]
	cv.Set(cx, cy, brailleBase+rune(bits), S(c))
}

func intBresenham(x0, y0, x1, y1 int, fn func(x, y, i int)) {
	dx, dy := absInt(x1-x0), absInt(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for i := 0; ; i++ {
		fn(x0, y0, i)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// dotLine draws a segment in dot-space (each cell is 2x4 dots).
func dotLine(cv *Canvas, gx0, gy0, gx1, gy1 int, c Color, dashed bool) {
	intBresenham(gx0, gy0, gx1, gy1, func(x, y, i int) {
		if dashed && (i/4)%2 == 1 {
			return
		}
		setDot(cv, x, y, c)
	})
}

func slopeChar(dx, dy int) rune {
	if dx == 0 && dy == 0 {
		return '*'
	}
	if dy == 0 {
		return '-'
	}
	if dx == 0 {
		return '|'
	}
	if absInt(dx) >= 2*absInt(dy) {
		return '-'
	}
	if absInt(dy) >= 2*absInt(dx) {
		return '|'
	}
	if (dx > 0) == (dy > 0) {
		return '\\'
	}
	return '/'
}

func asciiLine(cv *Canvas, x0, y0, x1, y1 int, st Style) {
	intBresenham(x0, y0, x1, y1, func(x, y, _ int) {
		cv.At(x, y)
		dx, dy := x1-x0, y1-y0
		cv.Set(x, y, slopeChar(dx, dy), st)
	})
}
