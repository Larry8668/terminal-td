package mapdata

// TilesOnSegment returns grid coordinates (x,y) that the line segment from (x0,y0) to (x1,y1) passes through (Bresenham).
func TilesOnSegment(x0, y0, x1, y1 int) [][2]int {
	var out [][2]int
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		out = append(out, [2]int{x0, y0})
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
	return out
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
