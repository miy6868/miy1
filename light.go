package main

// Lighting: two channels per block. SkyLight is daylight filtering down from
// the sky (scaled by time of day at render); BlockLight comes from torches,
// lava, glowstone etc. Both propagate with -1 falloff per block.

type lightNode struct {
	x, y  int
	level uint8
}

// RecomputeLight relights the chunk window [cx0, cx1] if any chunk in it is
// dirty. Neighbors one chunk out are included as a stable border.
func (w *World) RecomputeLight(cx0, cx1 int) {
	dirty := false
	for cx := cx0; cx <= cx1; cx++ {
		if w.ChunkAt(cx).LightDirty {
			dirty = true
			break
		}
	}
	if !dirty {
		return
	}
	// Work area with one chunk margin so borders look right.
	ax0 := (cx0 - 1) * ChunkW
	ax1 := (cx1+2)*ChunkW - 1
	width := ax1 - ax0 + 1

	sky := make([]uint8, width*WorldH)
	blk := make([]uint8, width*WorldH)
	idx := func(x, y int) int { return y*width + (x - ax0) }

	var queue []lightNode

	// Sky light: full brightness above the highest opaque block per column.
	for x := ax0; x <= ax1; x++ {
		c := w.ChunkAt(floorDiv(x, ChunkW))
		h := c.Height[mod(x, ChunkW)]
		for y := 0; y < h && y < WorldH; y++ {
			sky[idx(x, y)] = 15
			// Seed horizontal spread from the column edges.
			queue = append(queue, lightNode{x, y, 15})
		}
	}
	w.propagate(sky, queue, ax0, ax1, width, idx)

	// Block light: seed from emissive blocks.
	queue = queue[:0]
	for x := ax0; x <= ax1; x++ {
		c := w.ChunkAt(floorDiv(x, ChunkW))
		lx := mod(x, ChunkW)
		for y := 0; y < WorldH; y++ {
			if e := c.At(lx, y).Info().LightEmit; e > 0 {
				blk[idx(x, y)] = uint8(e)
				queue = append(queue, lightNode{x, y, uint8(e)})
			}
		}
	}
	w.propagate(blk, queue, ax0, ax1, width, idx)

	// Copy results into the chunks and clear dirty flags.
	for cx := cx0 - 1; cx <= cx1+1; cx++ {
		c := w.ChunkAt(cx)
		for lx := 0; lx < ChunkW; lx++ {
			x := cx*ChunkW + lx
			for y := 0; y < WorldH; y++ {
				c.SkyLight[y*ChunkW+lx] = sky[idx(x, y)]
				c.BlockLight[y*ChunkW+lx] = blk[idx(x, y)]
			}
		}
		if cx >= cx0 && cx <= cx1 {
			c.LightDirty = false
		}
	}
}

// propagate runs BFS flood fill with per-block attenuation.
func (w *World) propagate(grid []uint8, queue []lightNode, ax0, ax1, width int, idx func(int, int) int) {
	for head := 0; head < len(queue); head++ {
		n := queue[head]
		if grid[idx(n.x, n.y)] > n.level {
			continue
		}
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := n.x+d[0], n.y+d[1]
			if nx < ax0 || nx > ax1 || ny < 0 || ny >= WorldH {
				continue
			}
			b := w.Block(nx, ny)
			att := uint8(1)
			if b.Opaque() {
				att = 4
			} else if b == BWater {
				att = 2
			}
			if n.level <= att {
				continue
			}
			nl := n.level - att
			if grid[idx(nx, ny)] < nl {
				grid[idx(nx, ny)] = nl
				queue = append(queue, lightNode{nx, ny, nl})
			}
		}
	}
}

// lightLevelAt combines both channels with time of day (0..15).
func (g *Game) lightLevelAt(x, y int) int {
	if y < 0 {
		y = 0
	}
	if y >= WorldH {
		return 0
	}
	c := g.World.ChunkAt(floorDiv(x, ChunkW))
	lx := mod(x, ChunkW)
	skyF := g.World.DayFactor()
	sky := float64(c.SkyLight[y*ChunkW+lx]) * (0.25 + 0.75*skyF)
	blk := float64(c.BlockLight[y*ChunkW+lx])
	l := sky
	if blk > l {
		l = blk
	}
	return int(l)
}
