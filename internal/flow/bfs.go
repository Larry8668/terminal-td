package flow

func ComputeDistances(field *Field, walkable [][]bool, baseX, baseY int) {
	for y := 0; y < field.Height; y++ {
		for x := 0; x < field.Width; x++ {
			field.Distances[y][x] = Inf
		}
	}
	if baseX < 0 || baseX >= field.Width || baseY < 0 || baseY >= field.Height {
		return
	}
	if !walkable[baseY][baseX] {
		return
	}

	field.Distances[baseY][baseX] = 0
	type cell struct{ x, y int }
	q := []cell{{baseX, baseY}}

	dx := []int{0, 0, -1, 1}
	dy := []int{-1, 1, 0, 0}

	for len(q) > 0 {
		c := q[0]
		q = q[1:]
		curDist := field.Distances[c.y][c.x]

		for i := 0; i < 4; i++ {
			nx, ny := c.x+dx[i], c.y+dy[i]
			if nx < 0 || nx >= field.Width || ny < 0 || ny >= field.Height {
				continue
			}
			if !walkable[ny][nx] {
				continue
			}
			newDist := curDist + 1
			if newDist < field.Distances[ny][nx] {
				field.Distances[ny][nx] = newDist
				q = append(q, cell{nx, ny})
			}
		}
	}
}

func ComputeDirections(field *Field, walkable [][]bool) {
	dx := []int{0, 0, -1, 1}
	dy := []int{-1, 1, 0, 0}

	for y := 0; y < field.Height; y++ {
		for x := 0; x < field.Width; x++ {
			if !walkable[y][x] {
				field.Directions[y][x] = Vec2{}
				continue
			}
			cur := field.Distances[y][x]
			bestNx, bestNy := -1, -1
			bestDist := cur

			// Neighbor order: up, down, left, right. When multiple neighbors have the same distance, we pick the first (deterministic).
			for i := 0; i < 4; i++ {
				nx, ny := x+dx[i], y+dy[i]
				if nx < 0 || nx >= field.Width || ny < 0 || ny >= field.Height {
					continue
				}
				if !walkable[ny][nx] {
					continue
				}
				d := field.Distances[ny][nx]
				if d < bestDist {
					bestDist = d
					bestNx, bestNy = nx, ny
				}
			}

			if bestNx < 0 {
				field.Directions[y][x] = Vec2{}
				continue
			}
			dir := Vec2{
				X: float64(bestNx - x),
				Y: float64(bestNy - y),
			}
			field.Directions[y][x] = dir.Normalize()
		}
	}
}

func Compute(width, height int, walkable [][]bool, baseX, baseY int) *Field {
	field := NewField(width, height)
	ComputeDistances(field, walkable, baseX, baseY)
	ComputeDirections(field, walkable)
	return field
}
