package main

import (
	"math"
	"math/rand"
)

// Biome kinds for the surface.
type Biome int

const (
	BiomePlains Biome = iota
	BiomeForest
	BiomeDesert
	BiomeSnowy
	BiomeMountains
	BiomeSwamp
)

// Generator produces terrain deterministically from a seed.
type Generator struct {
	seed    int64
	terrain *Perlin
	detail  *Perlin
	caveA   *Perlin // spaghetti caves
	caveB   *Perlin // cheese caverns
	ore     *Perlin
	biomeN  *Perlin
	lushN   *Perlin
	sculkN  *Perlin
	mushN   *Perlin
}

func NewGenerator(seed int64) *Generator {
	return &Generator{
		seed:    seed,
		terrain: NewPerlin(seed),
		detail:  NewPerlin(seed ^ 0x51ed270b),
		caveA:   NewPerlin(seed ^ 0x2c9277b5),
		caveB:   NewPerlin(seed ^ 0x6a09e667),
		ore:     NewPerlin(seed ^ 0x3c6ef372),
		biomeN:  NewPerlin(seed ^ 0x1f83d9ab),
		lushN:   NewPerlin(seed ^ 0x5be0cd19),
		sculkN:  NewPerlin(seed ^ 0x428a2f98),
		mushN:   NewPerlin(seed ^ 0x71374491),
	}
}

// BiomeAt returns the biome for a world column.
func (g *Generator) BiomeAt(x int) Biome {
	n := g.biomeN.Octave1(float64(x)*0.0015, 3, 0.5)
	m := g.biomeN.Noise2(float64(x)*0.0011, 91.3)
	switch {
	case n > 0.42:
		return BiomeMountains
	case n > 0.22:
		return BiomeForest
	case n < -0.42:
		return BiomeDesert
	case n < -0.22:
		return BiomeSnowy
	case m > 0.35:
		return BiomeSwamp
	default:
		return BiomePlains
	}
}

// SurfaceY returns terrain surface height (smaller y = higher).
func (g *Generator) SurfaceY(x int) int {
	fx := float64(x)
	base := g.terrain.Octave1(fx*0.004, 4, 0.5) * 22
	rough := g.detail.Octave1(fx*0.02, 3, 0.5) * 5
	mountain := 0.0
	bn := g.biomeN.Octave1(fx*0.0015, 3, 0.5)
	if bn > 0.30 { // mountains rise steeply
		mountain = (bn - 0.30) * 130
	}
	h := SurfaceBaseY - int(base+rough+mountain)
	if h < 8 {
		h = 8
	}
	if h > WorldH-40 {
		h = WorldH - 40
	}
	return h
}

// caveAt reports whether (x, y) is carved out as cave air.
func (g *Generator) caveAt(x, y int, surface int) bool {
	if y >= BedrockY-1 {
		return false
	}
	depth := y - surface
	if depth < 4 {
		return false
	}
	fx, fy := float64(x), float64(y)
	// Spaghetti tunnels: thin winding corridors, widening with depth.
	w := 0.055 + math.Min(float64(depth)/600.0, 0.06)
	a := g.caveA.Octave2(fx*0.015, fy*0.03, 2, 0.5)
	if math.Abs(a) < w {
		return true
	}
	// Second tunnel system at an offset frequency for more connectivity.
	a2 := g.caveA.Octave2(fx*0.011+500, fy*0.022, 2, 0.5)
	if math.Abs(a2) < w*0.8 {
		return true
	}
	// Cheese caverns: big open rooms, much more common deep down.
	t := 0.62 - math.Min(float64(depth)/900.0, 0.27)
	b := g.caveB.Octave2(fx*0.006, fy*0.012, 3, 0.5)
	return b > t
}

// oreAt decides an ore block for a stone position, or BAir for none.
func (g *Generator) oreAt(x, y int) Block {
	r := hashFloat(g.seed^0x0ee1, x, y)
	deep := y >= DeepslateY
	pick := func(normal, deepB Block) Block {
		if deep {
			return deepB
		}
		return normal
	}
	// Vein shaping: ores cluster along ore-noise bands.
	v := g.ore.Octave2(float64(x)*0.05, float64(y)*0.05, 2, 0.5)
	cluster := math.Abs(v) < 0.22

	switch {
	case y > 90 && r < 0.014 && cluster:
		return pick(BCoalOre, BDeepCoalOre)
	case y > 110 && r >= 0.014 && r < 0.026 && cluster:
		return pick(BIronOre, BDeepIronOre)
	case y > 100 && y < 200 && r >= 0.026 && r < 0.034 && cluster:
		return BCopperOre
	case y > 145 && r >= 0.034 && r < 0.041 && cluster:
		return pick(BGoldOre, BDeepGoldOre)
	case y > 150 && r >= 0.041 && r < 0.050 && cluster:
		return pick(BRedstoneOre, BDeepRedstoneOre)
	case y > 150 && y < 260 && r >= 0.050 && r < 0.056 && cluster:
		return pick(BLapisOre, BDeepLapisOre)
	case y > 240 && r >= 0.056 && r < 0.0615 && cluster:
		return pick(BDiamondOre, BDeepDiamondOre)
	case y > 220 && r >= 0.0615 && r < 0.064:
		return pick(BEmeraldOre, BDeepEmeraldOre) // emeralds: rare, no cluster needed
	}
	return BAir
}

// Generate builds one chunk.
func (g *Generator) Generate(cx int, w *World) *Chunk {
	c := &Chunk{CX: cx}
	x0 := cx * ChunkW

	surfaces := make([]int, ChunkW)
	biomes := make([]Biome, ChunkW)
	for lx := 0; lx < ChunkW; lx++ {
		surfaces[lx] = g.SurfaceY(x0 + lx)
		biomes[lx] = g.BiomeAt(x0 + lx)
	}

	for lx := 0; lx < ChunkW; lx++ {
		x := x0 + lx
		surface := surfaces[lx]
		biome := biomes[lx]
		for y := 0; y < WorldH; y++ {
			var b Block
			switch {
			case y >= BedrockY:
				b = BBedrock
			case y == BedrockY-1 && hashFloat(g.seed^0xbed, x, y) < 0.5:
				b = BBedrock
			case y < surface:
				b = BAir
			case y == surface:
				b = g.surfaceBlock(biome)
			case y < surface+4:
				b = g.subsurfaceBlock(biome, y)
			default:
				if y >= DeepslateY {
					b = BDeepslate
				} else {
					b = BStone
				}
				// Underground biome material patches.
				b = g.undergroundMaterial(x, y, b)
				if ore := g.oreAt(x, y); ore != BAir {
					b = ore
				}
			}
			c.Set(lx, y, b)
		}

		// Carve caves.
		for y := surfaces[lx] - 2; y < BedrockY-1; y++ {
			if y < 0 {
				continue
			}
			if g.caveAt(x, y, surfaces[lx]) {
				if y >= LavaLevel {
					c.Set(lx, y, BLava)
				} else {
					c.Set(lx, y, BAir)
				}
			}
		}

		// Ocean / lake water above terrain.
		if surfaces[lx] > SeaLevel {
			for y := SeaLevel; y < surfaces[lx]; y++ {
				if c.At(lx, y) == BAir {
					if biomes[lx] == BiomeSnowy && y == SeaLevel {
						c.Set(lx, y, BIce)
					} else {
						c.Set(lx, y, BWater)
					}
				}
			}
		}
	}

	g.decorateUnderground(c, x0)
	g.placeStructures(c, x0, w)
	g.decorateSurface(c, x0, surfaces, biomes)

	c.recalcHeights()
	return c
}

func (g *Generator) surfaceBlock(b Biome) Block {
	switch b {
	case BiomeDesert:
		return BSand
	case BiomeSnowy:
		return BSnow
	case BiomeMountains:
		return BStone
	default:
		return BGrass
	}
}

func (g *Generator) subsurfaceBlock(b Biome, y int) Block {
	switch b {
	case BiomeDesert:
		if y%2 == 0 {
			return BSandstone
		}
		return BSand
	case BiomeMountains:
		return BStone
	default:
		return BDirt
	}
}

// undergroundMaterial swaps stone for biome materials: lush moss, sculk,
// mushroom soil, gravel/dirt pockets, clay near water depth.
func (g *Generator) undergroundMaterial(x, y int, base Block) Block {
	fx, fy := float64(x), float64(y)
	// Lush cave patches (mid depth).
	if y > 110 && y < 230 {
		if g.lushN.Octave2(fx*0.008, fy*0.016, 2, 0.5) > 0.45 {
			return BMossBlock
		}
	}
	// Mushroom cave patches.
	if y > 130 && y < 260 {
		if g.mushN.Octave2(fx*0.007, fy*0.014, 2, 0.5) > 0.52 {
			return BDirt
		}
	}
	// Deep dark sculk.
	if y > DeepDarkY {
		if g.sculkN.Octave2(fx*0.009, fy*0.018, 2, 0.5) > 0.40 {
			return BSculk
		}
	}
	// Scattered pockets.
	r := hashFloat(g.seed^0x9dc5, x/4, y/4)
	switch {
	case r < 0.02:
		return BGravel
	case r < 0.035 && y < DeepslateY:
		return BDirt
	case r < 0.045 && y > 200 && y < 290:
		return BObsidian
	}
	return base
}

// decorateUnderground adds features that live inside caves: glow berries,
// mushrooms, dripstone, amethyst clusters, glowstone in deep caverns.
func (g *Generator) decorateUnderground(c *Chunk, x0 int) {
	for lx := 0; lx < ChunkW; lx++ {
		x := x0 + lx
		for y := 80; y < BedrockY-1; y++ {
			cur := c.At(lx, y)
			if cur != BAir {
				continue
			}
			above := c.At(lx, y-1)
			below := c.At(lx, y+1)
			r := hashFloat(g.seed^0xdec0, x, y)

			// Hanging from ceilings.
			if above.Opaque() {
				switch {
				case above == BMossBlock && r < 0.30:
					c.Set(lx, y, BGlowBerries)
				case above == BDripstone || (above == BStone || above == BDeepslate) && r < 0.05:
					c.Set(lx, y, BDripstone)
				case y > 270 && r >= 0.05 && r < 0.075:
					c.Set(lx, y, BGlowstone)
				}
				continue
			}
			// Growing on floors.
			if below.Opaque() {
				switch {
				case below == BMossBlock && r < 0.22:
					if r < 0.08 {
						c.Set(lx, y, BTallGrass)
					} else if r < 0.13 {
						c.Set(lx, y, BFlowerYellow)
					} else {
						c.Set(lx, y, BBrownMushroom)
					}
				case below == BDirt && y > 130 && r < 0.16:
					if r < 0.08 {
						c.Set(lx, y, BBrownMushroom)
					} else {
						c.Set(lx, y, BRedMushroom)
					}
				case (below == BStone || below == BDeepslate) && r < 0.03:
					c.Set(lx, y, BDripstone)
				case below == BSculk && r < 0.10:
					c.Set(lx, y, BSculkSensor)
				}
			}
		}
	}

	// Underground water pools: fill cave floors at mid depth occasionally.
	for lx := 0; lx < ChunkW; lx++ {
		x := x0 + lx
		for y := 120; y < LavaLevel-6; y++ {
			if c.At(lx, y) != BAir {
				continue
			}
			if hashFloat(g.seed^0x77a7e4, x/12, y/8) < 0.10 && c.At(lx, y+1).Solid() {
				c.Set(lx, y, BWater)
			}
		}
	}
}

// ---- Structures ----
// Structures are decided per fixed-size region so that any chunk overlapping
// a structure generates its own slice of it deterministically.

const structRegion = 4 // chunks per structure region

type box struct{ x0, y0, x1, y1 int }

func (g *Generator) placeStructures(c *Chunk, x0 int, w *World) {
	cr0 := floorDiv(c.CX-1, structRegion)
	cr1 := floorDiv(c.CX+1, structRegion)
	for r := cr0; r <= cr1; r++ {
		g.applyMineshaft(c, r, w)
		g.applyDungeon(c, r, w)
		g.applyGeode(c, r)
		g.applyAncientVault(c, r, w)
	}
}

// applyMineshaft draws abandoned mineshaft corridors for region r that
// intersect chunk c: plank floor, fences+torch supports, rails, chests.
func (g *Generator) applyMineshaft(c *Chunk, r int, w *World) {
	h := hash2(g.seed^0x111e, r, 0)
	if h%100 >= 55 { // 55% of regions have a mineshaft
		return
	}
	rng := rand.New(rand.NewSource(int64(h)))
	baseX := r*structRegion*ChunkW + rng.Intn(ChunkW*2)
	baseY := 130 + rng.Intn(140)
	nCorr := 2 + rng.Intn(3)
	x0 := c.CX * ChunkW

	for i := 0; i < nCorr; i++ {
		y := baseY + (rng.Intn(5)-2)*8
		if y < 100 {
			y = 100
		}
		if y > BedrockY-10 {
			y = BedrockY - 10
		}
		cx0 := baseX + rng.Intn(60) - 30
		length := 35 + rng.Intn(50)
		hasRail := rng.Intn(2) == 0
		for x := cx0; x < cx0+length; x++ {
			lx := x - x0
			if lx < 0 || lx >= ChunkW {
				continue
			}
			// 3-high corridor with plank floor.
			c.Set(lx, y+1, BPlanks)
			for dy := 0; dy > -3; dy-- {
				c.Set(lx, y+dy, BAir)
			}
			if hasRail {
				c.Set(lx, y, BRail)
			}
			// Support beams every 6 blocks.
			if mod(x-cx0, 6) == 0 {
				c.Set(lx, y, BFence)
				c.Set(lx, y-1, BFence)
				c.Set(lx, y-2, BPlanks)
			} else if mod(x-cx0, 6) == 3 && rng.Intn(4) == 0 {
				c.Set(lx, y-2, BTorch)
			}
			// Loot chest.
			if hash2(g.seed^0xc4e57, x, y)%97 == 0 {
				c.Set(lx, y, BChest)
				w.Chests[[2]int{x, y}] = mineshaftLoot(g.seed, x, y)
			}
			// Cobweb-ish: string drops via spawner corridors — place spawner rarely.
			if hash2(g.seed^0x59a44, x, y)%251 == 0 {
				c.Set(lx, y, BSpawner)
			}
		}
		// Vertical ladder shaft connecting corridors to the one above.
		if i > 0 {
			sx := cx0 + rng.Intn(length)
			lx := sx - x0
			if lx >= 0 && lx < ChunkW {
				for yy := y - 8*3; yy <= y; yy++ {
					if yy > 0 && yy < BedrockY {
						c.Set(lx, yy, BLadder)
					}
				}
			}
		}
	}
}

// applyDungeon carves a mossy cobble room with a spawner and loot chests.
func (g *Generator) applyDungeon(c *Chunk, r int, w *World) {
	h := hash2(g.seed^0xd00e, r, 1)
	if h%100 >= 60 {
		return
	}
	rng := rand.New(rand.NewSource(int64(h)))
	roomX := r*structRegion*ChunkW + rng.Intn(structRegion*ChunkW)
	roomY := 180 + rng.Intn(110)
	rw, rh := 9+rng.Intn(4), 6+rng.Intn(2)
	x0 := c.CX * ChunkW
	b := box{roomX, roomY, roomX + rw, roomY + rh}
	for x := b.x0; x <= b.x1; x++ {
		lx := x - x0
		if lx < 0 || lx >= ChunkW {
			continue
		}
		for y := b.y0; y <= b.y1; y++ {
			if y <= 0 || y >= BedrockY {
				continue
			}
			edge := x == b.x0 || x == b.x1 || y == b.y0 || y == b.y1
			if edge {
				if hash2(g.seed^0x30553, x, y)%3 == 0 {
					c.Set(lx, y, BMossyCobble)
				} else {
					c.Set(lx, y, BCobble)
				}
			} else {
				c.Set(lx, y, BAir)
			}
		}
	}
	// Spawner in the center, chests at the sides.
	sx, sy := (b.x0+b.x1)/2, b.y1-1
	if lx := sx - x0; lx >= 0 && lx < ChunkW {
		c.Set(lx, sy, BSpawner)
	}
	for _, chX := range []int{b.x0 + 1, b.x1 - 1} {
		if hash2(g.seed^0x10071, chX, sy)%2 == 0 {
			if lx := chX - x0; lx >= 0 && lx < ChunkW {
				c.Set(lx, sy, BChest)
				w.Chests[[2]int{chX, sy}] = dungeonLoot(g.seed, chX, sy)
			}
		}
	}
}

// applyGeode places an amethyst geode: obsidian shell, amethyst lining,
// hollow crystal-filled center.
func (g *Generator) applyGeode(c *Chunk, r int) {
	h := hash2(g.seed^0x6e0de, r, 2)
	if h%100 >= 40 {
		return
	}
	rng := rand.New(rand.NewSource(int64(h)))
	gx := r*structRegion*ChunkW + rng.Intn(structRegion*ChunkW)
	gy := 200 + rng.Intn(90)
	rad := 5 + rng.Intn(3)
	x0 := c.CX * ChunkW
	for x := gx - rad; x <= gx+rad; x++ {
		lx := x - x0
		if lx < 0 || lx >= ChunkW {
			continue
		}
		for y := gy - rad; y <= gy+rad; y++ {
			if y <= 0 || y >= BedrockY {
				continue
			}
			dx, dy := float64(x-gx), float64(y-gy)
			d := math.Sqrt(dx*dx + dy*dy)
			fr := float64(rad)
			switch {
			case d > fr:
				// outside
			case d > fr-1.2:
				c.Set(lx, y, BObsidian)
			case d > fr-2.4:
				c.Set(lx, y, BAmethyst)
			case d > fr-3.4:
				if hash2(g.seed^0xa3e, x, y)%3 == 0 {
					c.Set(lx, y, BAmethystCluster)
				} else {
					c.Set(lx, y, BAir)
				}
			default:
				c.Set(lx, y, BAir)
			}
		}
	}
}

// applyAncientVault builds a deep-dark stone-brick vault with rich loot,
// guarded by spawners — the deepest structure.
func (g *Generator) applyAncientVault(c *Chunk, r int, w *World) {
	h := hash2(g.seed^0xa9c1e97, r, 3)
	if h%100 >= 25 {
		return
	}
	rng := rand.New(rand.NewSource(int64(h)))
	vx := r*structRegion*ChunkW + rng.Intn(structRegion*ChunkW)
	vy := DeepDarkY + 15 + rng.Intn(35)
	vw, vh := 15+rng.Intn(8), 8+rng.Intn(3)
	x0 := c.CX * ChunkW
	b := box{vx, vy, vx + vw, vy + vh}
	for x := b.x0; x <= b.x1; x++ {
		lx := x - x0
		if lx < 0 || lx >= ChunkW {
			continue
		}
		for y := b.y0; y <= b.y1; y++ {
			if y <= 0 || y >= BedrockY {
				continue
			}
			edge := x == b.x0 || x == b.x1 || y == b.y0 || y == b.y1
			if edge {
				if hash2(g.seed^0x8f2, x, y)%4 == 0 {
					c.Set(lx, y, BMossyStoneBricks)
				} else {
					c.Set(lx, y, BStoneBricks)
				}
			} else {
				c.Set(lx, y, BAir)
			}
		}
	}
	floor := b.y1 - 1
	// Pillars and decoration inside.
	for x := b.x0 + 3; x < b.x1-2; x += 4 {
		if lx := x - x0; lx >= 0 && lx < ChunkW {
			c.Set(lx, floor, BStoneBricks)
			c.Set(lx, floor-1, BStoneBricks)
			c.Set(lx, floor-2, BGlowstone)
		}
	}
	// Loot chests + guardian spawners.
	for _, chX := range []int{b.x0 + 1, (b.x0 + b.x1) / 2, b.x1 - 1} {
		if lx := chX - x0; lx >= 0 && lx < ChunkW {
			if hash2(g.seed^0x77c, chX, floor)%3 != 0 {
				c.Set(lx, floor, BChest)
				w.Chests[[2]int{chX, floor}] = vaultLoot(g.seed, chX, floor)
			} else {
				c.Set(lx, floor, BSpawner)
			}
		}
	}
}

// ---- Loot tables ----

func rollLoot(seed int64, x, y int, table []lootEntry) []ItemStack {
	rng := rand.New(rand.NewSource(int64(hash2(seed^0x100c, x, y))))
	var out []ItemStack
	for _, e := range table {
		if rng.Float64() < e.chance {
			n := e.min
			if e.max > e.min {
				n += rng.Intn(e.max - e.min + 1)
			}
			out = append(out, ItemStack{e.item, n})
		}
	}
	if len(out) == 0 {
		out = append(out, ItemStack{IBread, 1})
	}
	return out
}

type lootEntry struct {
	item   Item
	min    int
	max    int
	chance float64
}

func mineshaftLoot(seed int64, x, y int) []ItemStack {
	return rollLoot(seed, x, y, []lootEntry{
		{ICoal, 2, 6, 0.6}, {IIronIngot, 1, 3, 0.4}, {IBread, 1, 2, 0.5},
		{IRailItem, 2, 8, 0.5}, {ITorch, 2, 8, 0.6}, {IRedstone, 2, 5, 0.25},
		{ILapis, 2, 5, 0.2}, {IGoldIngot, 1, 2, 0.15}, {IDiamond, 1, 1, 0.06},
		{IWheatSeeds, 1, 3, 0.3},
	})
}

func dungeonLoot(seed int64, x, y int) []ItemStack {
	return rollLoot(seed, x, y, []lootEntry{
		{IBone, 1, 4, 0.6}, {IRottenFlesh, 1, 4, 0.5}, {IString, 1, 3, 0.4},
		{IIronIngot, 1, 4, 0.45}, {IGoldIngot, 1, 3, 0.3}, {IBread, 1, 2, 0.4},
		{IGoldenApple, 1, 1, 0.12}, {IBucket, 1, 1, 0.2}, {IDiamond, 1, 2, 0.1},
		{ICompass, 1, 1, 0.1}, {IArrow, 3, 8, 0.3},
	})
}

func vaultLoot(seed int64, x, y int) []ItemStack {
	return rollLoot(seed, x, y, []lootEntry{
		{IDiamond, 1, 3, 0.5}, {IEmerald, 1, 3, 0.4}, {IGoldenApple, 1, 2, 0.35},
		{IXPGem, 2, 5, 0.6}, {IEnderPearl, 1, 2, 0.3}, {IIronIngot, 2, 6, 0.5},
		{IGlowstoneBlk, 2, 4, 0.3}, {IObsidianItem, 2, 5, 0.25},
		{IDiamondPickaxe, 1, 1, 0.08}, {IDiamondSword, 1, 1, 0.08},
	})
}

// ---- Surface decoration ----

func (g *Generator) decorateSurface(c *Chunk, x0 int, surfaces []int, biomes []Biome) {
	// Trees may span chunk borders: consider columns a few blocks outside.
	for x := x0 - 3; x < x0+ChunkW+3; x++ {
		biome := g.BiomeAt(x)
		surface := g.SurfaceY(x)
		if surface >= SeaLevel {
			continue // underwater columns get no plants
		}
		r := hashFloat(g.seed^0x7ee5, x, 0)
		switch biome {
		case BiomeForest:
			if r < 0.16 {
				g.drawTree(c, x0, x, surface, r)
			}
		case BiomePlains, BiomeSwamp:
			if r < 0.05 {
				g.drawTree(c, x0, x, surface, r)
			}
		case BiomeSnowy:
			if r < 0.08 {
				g.drawTree(c, x0, x, surface, r)
			}
		case BiomeDesert:
			if r < 0.04 {
				lx := x - x0
				if lx >= 0 && lx < ChunkW {
					hgt := 1 + int(r*100)%3
					for i := 1; i <= hgt; i++ {
						c.Set(lx, surface-i, BCactus)
					}
				}
			}
		}
		// Small plants (only inside this chunk).
		lx := x - x0
		if lx >= 0 && lx < ChunkW && c.At(lx, surface) == BGrass && c.At(lx, surface-1) == BAir {
			r2 := hashFloat(g.seed^0x91a55, x, 1)
			switch {
			case r2 < 0.18:
				c.Set(lx, surface-1, BTallGrass)
			case r2 < 0.22:
				c.Set(lx, surface-1, BFlowerYellow)
			case r2 < 0.25:
				c.Set(lx, surface-1, BFlowerRed)
			}
		}
	}
}

func (g *Generator) drawTree(c *Chunk, x0, x, surface int, r float64) {
	trunkH := 4 + int(r*1000)%3
	topY := surface - 1 - trunkH
	// Trunk.
	if lx := x - x0; lx >= 0 && lx < ChunkW {
		if c.At(lx, surface) != BGrass && c.At(lx, surface) != BDirt && c.At(lx, surface) != BSnow {
			return
		}
		for i := 1; i <= trunkH; i++ {
			c.Set(lx, surface-i, BLog)
		}
	}
	// Canopy.
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 1; dy++ {
			if dx*dx+dy*dy > 5 {
				continue
			}
			lx := x + dx - x0
			y := topY + dy
			if lx >= 0 && lx < ChunkW && y > 0 {
				if c.At(lx, y) == BAir {
					c.Set(lx, y, BLeaves)
				}
			}
		}
	}
}
