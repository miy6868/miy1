package main

const (
	TileSize = 16  // pixels per block
	ChunkW   = 32  // blocks per chunk horizontally
	WorldH   = 320 // world height in blocks

	SeaLevel     = 68  // water fills terrain below this
	DeepslateY   = 170 // stone turns to deepslate below this depth
	LavaLevel    = 300 // caves below this flood with lava
	BedrockY     = 316 // bedrock starts here
	DeepDarkY    = 250 // sculk patches appear below this
	SurfaceBaseY = 64  // average terrain height
)

// Chunk is a vertical slice of the world, ChunkW blocks wide.
type Chunk struct {
	CX         int
	Blocks     [ChunkW * WorldH]Block
	BlockLight [ChunkW * WorldH]uint8
	SkyLight   [ChunkW * WorldH]uint8
	Height     [ChunkW]int // first solid/opaque block from the top per column
	LightDirty bool
}

func (c *Chunk) At(lx, y int) Block {
	if y < 0 || y >= WorldH {
		return BAir
	}
	return c.Blocks[y*ChunkW+lx]
}

func (c *Chunk) Set(lx, y int, b Block) {
	if y < 0 || y >= WorldH {
		return
	}
	c.Blocks[y*ChunkW+lx] = b
}

// World holds loaded chunks and global simulation state.
type World struct {
	Seed   int64
	Chunks map[int]*Chunk
	Gen    *Generator

	Time      float64 // world time in seconds; a full day is DayLength
	Drops     []*ItemDrop
	Mobs      []*Mob
	Particles []*Particle
	Chests    map[[2]int][]ItemStack // contents of chest blocks

	liquidTick int
	randomTick int
}

const DayLength = 600.0 // seconds per full day/night cycle

func NewWorld(seed int64) *World {
	w := &World{
		Seed:   seed,
		Chunks: map[int]*Chunk{},
		Chests: map[[2]int][]ItemStack{},
		Time:   DayLength * 0.05, // start at morning
	}
	w.Gen = NewGenerator(seed)
	return w
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func mod(a, b int) int {
	m := a % b
	if m < 0 {
		m += b
	}
	return m
}

// Chunk returns the chunk containing world column x, generating it if needed.
func (w *World) ChunkAt(cx int) *Chunk {
	if c, ok := w.Chunks[cx]; ok {
		return c
	}
	c := w.Gen.Generate(cx, w)
	w.Chunks[cx] = c
	c.LightDirty = true
	return c
}

// Block returns the block at world coordinates.
func (w *World) Block(x, y int) Block {
	if y < 0 {
		return BAir
	}
	if y >= WorldH {
		return BBedrock
	}
	cx := floorDiv(x, ChunkW)
	return w.ChunkAt(cx).At(mod(x, ChunkW), y)
}

// SetBlock changes a block and marks lighting dirty around it.
func (w *World) SetBlock(x, y int, b Block) {
	if y < 0 || y >= WorldH {
		return
	}
	cx := floorDiv(x, ChunkW)
	c := w.ChunkAt(cx)
	c.Set(mod(x, ChunkW), y, b)
	c.LightDirty = true
	if n, ok := w.Chunks[cx-1]; ok {
		n.LightDirty = true
	}
	if n, ok := w.Chunks[cx+1]; ok {
		n.LightDirty = true
	}
	c.recalcHeightColumn(mod(x, ChunkW))
}

func (c *Chunk) recalcHeightColumn(lx int) {
	h := WorldH
	for y := 0; y < WorldH; y++ {
		if c.At(lx, y).Opaque() {
			h = y
			break
		}
	}
	c.Height[lx] = h
}

func (c *Chunk) recalcHeights() {
	for lx := 0; lx < ChunkW; lx++ {
		c.recalcHeightColumn(lx)
	}
}

// DayFactor is 1 at noon, 0 at midnight, smooth in between.
func (w *World) DayFactor() float64 {
	t := w.Time / DayLength // 0..1, 0 = dawn
	// Day from 0.0 to 0.5, night 0.5 to 1.0, with smooth transitions.
	switch {
	case t < 0.04:
		return t / 0.04
	case t < 0.46:
		return 1
	case t < 0.54:
		return 1 - (t-0.46)/0.08
	case t < 0.96:
		return 0
	default:
		return (t - 0.96) / 0.04
	}
}

func (w *World) IsNight() bool { return w.DayFactor() < 0.3 }

// SurfaceY returns the terrain height for a column (cheap, generator-based).
func (w *World) SurfaceY(x int) int { return w.Gen.SurfaceY(x) }

// Chunks are kept loaded for the whole session (~30 KB each) so player
// edits are never lost; only chunks the player has actually visited exist.
