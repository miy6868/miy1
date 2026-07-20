package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Procedurally generated textures for every block and item: no external
// assets. Blocks are drawn into a CPU-side buffer, given a beveled-edge
// relief pass, then uploaded as GPU images. Liquids get animation frames.

var (
	blockTex    [BBlockCount]*ebiten.Image
	waterFrames [4]*ebiten.Image
	lavaFrames  [4]*ebiten.Image
	cloudTex    [3]*ebiten.Image
	itemTex     = map[Item]*ebiten.Image{}
	whiteTex    *ebiten.Image
)

type rgb struct{ r, g, b uint8 }

func (c rgb) shade(f float64) rgb {
	cl := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return rgb{cl(float64(c.r) * f), cl(float64(c.g) * f), cl(float64(c.b) * f)}
}

var blockBase = map[Block]rgb{
	BGrass: {106, 170, 64}, BDirt: {134, 96, 67}, BStone: {125, 125, 125},
	BCobble: {110, 110, 110}, BBedrock: {50, 50, 55}, BSand: {219, 207, 163},
	BGravel: {132, 127, 123}, BLog: {102, 81, 50}, BPlanks: {157, 128, 79},
	BLeaves: {58, 122, 40}, BCoalOre: {125, 125, 125}, BIronOre: {125, 125, 125},
	BCopperOre: {125, 125, 125}, BGoldOre: {125, 125, 125}, BRedstoneOre: {125, 125, 125},
	BLapisOre: {125, 125, 125}, BDiamondOre: {125, 125, 125}, BEmeraldOre: {125, 125, 125},
	BWater: {52, 90, 180}, BLava: {217, 90, 20}, BTorch: {255, 216, 96},
	BCraftTable: {140, 106, 62}, BFurnace: {100, 100, 100}, BChest: {160, 116, 48},
	BLadder: {150, 118, 70}, BGlass: {200, 225, 235}, BMossyCobble: {95, 115, 85},
	BSpawner: {35, 55, 75}, BMossBlock: {74, 110, 46}, BGlowBerries: {245, 195, 80},
	BSculk: {12, 40, 55}, BSculkSensor: {20, 65, 85}, BAmethyst: {135, 95, 190},
	BAmethystCluster: {170, 125, 225}, BObsidian: {28, 20, 45}, BMushroomStem: {200, 195, 185},
	BMushroomCap: {155, 60, 50}, BBrownMushroom: {145, 105, 70}, BRedMushroom: {190, 55, 45},
	BTallGrass: {96, 160, 60}, BFlowerYellow: {215, 200, 60}, BFlowerRed: {200, 60, 50},
	BClay: {150, 155, 165}, BSandstone: {210, 195, 150}, BSnow: {235, 240, 245},
	BIce: {150, 190, 235}, BCactus: {70, 130, 50}, BDeepslate: {75, 75, 80},
	BDeepCoalOre: {75, 75, 80}, BDeepIronOre: {75, 75, 80}, BDeepGoldOre: {75, 75, 80},
	BDeepRedstoneOre: {75, 75, 80}, BDeepLapisOre: {75, 75, 80}, BDeepDiamondOre: {75, 75, 80},
	BDeepEmeraldOre: {75, 75, 80}, BGlowstone: {235, 190, 105}, BFence: {157, 128, 79},
	BRail: {130, 110, 90}, BStoneBricks: {118, 118, 118}, BMossyStoneBricks: {100, 115, 95},
	BDripstone: {140, 118, 100}, BTNT: {190, 60, 45}, BBookshelf: {150, 120, 75},
	BWool: {225, 225, 225}, BDoor: {150, 120, 75},
}

var oreSpeck = map[Block]rgb{
	BCoalOre: {40, 40, 40}, BDeepCoalOre: {35, 35, 35},
	BIronOre: {216, 175, 147}, BDeepIronOre: {216, 175, 147},
	BCopperOre: {200, 115, 75},
	BGoldOre:  {250, 220, 80}, BDeepGoldOre: {250, 220, 80},
	BRedstoneOre: {220, 40, 40}, BDeepRedstoneOre: {220, 40, 40},
	BLapisOre: {40, 70, 200}, BDeepLapisOre: {40, 70, 200},
	BDiamondOre: {95, 230, 220}, BDeepDiamondOre: {95, 230, 220},
	BEmeraldOre: {60, 210, 100}, BDeepEmeraldOre: {60, 210, 100},
}

// noBevel lists blocks that should not receive the edge-relief pass
// (non-cube decorations, liquids, foliage).
var noBevel = map[Block]bool{
	BTorch: true, BLadder: true, BFence: true, BRail: true, BTallGrass: true,
	BFlowerYellow: true, BFlowerRed: true, BBrownMushroom: true, BRedMushroom: true,
	BGlowBerries: true, BAmethystCluster: true, BDripstone: true, BDoor: true,
	BWater: true, BLava: true, BLeaves: true, BSculkSensor: true, BCactus: true,
}

func blockAvgColor(b Block) (uint8, uint8, uint8) {
	c, ok := blockBase[b]
	if !ok {
		return 128, 128, 128
	}
	return c.r, c.g, c.b
}

// buildTextures creates all block and item textures once at startup.
func buildTextures() {
	whiteTex = ebiten.NewImage(3, 3)
	whiteTex.Fill(color.White)

	for b := Block(1); b < BBlockCount; b++ {
		blockTex[b] = makeBlockTex(b)
	}
	for f := 0; f < 4; f++ {
		waterFrames[f] = makeLiquidTex(BWater, f)
		lavaFrames[f] = makeLiquidTex(BLava, f)
	}
	buildClouds()
	buildItemTextures()
	buildGlowTex()
	buildVignette()
}

// texBuf is a small CPU pixel buffer for composing block art.
type texBuf struct{ img *image.NRGBA }

func newTexBuf() *texBuf {
	return &texBuf{image.NewNRGBA(image.Rect(0, 0, TileSize, TileSize))}
}

func (t *texBuf) set(x, y int, c rgb) {
	if x < 0 || y < 0 || x >= TileSize || y >= TileSize {
		return
	}
	t.img.SetNRGBA(x, y, color.NRGBA{c.r, c.g, c.b, 255})
}

func (t *texBuf) scalePx(x, y int, f float64) {
	c := t.img.NRGBAAt(x, y)
	if c.A == 0 {
		return
	}
	cl := func(v float64) uint8 {
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	t.img.SetNRGBA(x, y, color.NRGBA{cl(float64(c.R) * f), cl(float64(c.G) * f), cl(float64(c.B) * f), c.A})
}

// bevel applies a subtle top-lit relief so blocks read as raised cubes.
func (t *texBuf) bevel() {
	for x := 0; x < TileSize; x++ {
		t.scalePx(x, 0, 1.22)
		t.scalePx(x, 1, 1.10)
		t.scalePx(x, TileSize-1, 0.70)
		t.scalePx(x, TileSize-2, 0.86)
	}
	for y := 1; y < TileSize-1; y++ {
		t.scalePx(0, y, 1.12)
		t.scalePx(TileSize-1, y, 0.80)
	}
}

func makeBlockTex(b Block) *ebiten.Image {
	t := newTexBuf()
	rng := rand.New(rand.NewSource(int64(b) * 7919))
	base, ok := blockBase[b]
	if !ok {
		base = rgb{128, 128, 128}
	}
	set := t.set
	noiseFill := func(c rgb, amount float64) {
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				f := 1 + (rng.Float64()-0.5)*amount
				set(x, y, c.shade(f))
			}
		}
	}
	// Soft blobby large-scale shading to break up flat noise.
	softShade := func(strength float64) {
		ox, oy := rng.Float64()*8, rng.Float64()*8
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				v := math.Sin((float64(x)+ox)*0.5)*math.Cos((float64(y)+oy)*0.45) * strength
				t.scalePx(x, y, 1+v)
			}
		}
	}

	switch b {
	case BGrass:
		noiseFill(rgb{134, 96, 67}, 0.25)
		for x := 0; x < TileSize; x++ {
			d := 3 + rng.Intn(3)
			for y := 0; y < d; y++ {
				g := rgb{106, 170, 64}.shade(1 + (rng.Float64()-0.5)*0.3)
				if y == 0 {
					g = g.shade(1.15)
				}
				set(x, y, g)
			}
			// A few grass blades poking above the soil line.
			if rng.Intn(3) == 0 {
				set(x, d, rgb{96, 155, 58})
			}
		}
	case BSnow:
		noiseFill(rgb{134, 96, 67}, 0.25)
		for x := 0; x < TileSize; x++ {
			d := 4 + rng.Intn(2)
			for y := 0; y < d; y++ {
				set(x, y, rgb{238, 242, 248}.shade(1+(rng.Float64()-0.5)*0.08))
			}
		}
	case BStone, BDeepslate:
		noiseFill(base, 0.14)
		softShade(0.10)
		// Hairline cracks.
		for i := 0; i < 3; i++ {
			x, y := rng.Intn(12)+2, rng.Intn(12)+2
			l := 3 + rng.Intn(4)
			for j := 0; j < l; j++ {
				set(x, y, base.shade(0.72))
				x += rng.Intn(3) - 1
				y++
			}
		}
	case BCobble, BMossyCobble:
		noiseFill(base.shade(0.55), 0.15)
		// Rounded pebbles with light tops.
		for i := 0; i < 9; i++ {
			cx, cy := rng.Intn(14)+1, rng.Intn(14)+1
			r := 2 + rng.Intn(2)
			pc := base.shade(0.85 + rng.Float64()*0.4)
			if b == BMossyCobble && rng.Intn(3) == 0 {
				pc = rgb{95, 135, 75}.shade(0.9 + rng.Float64()*0.3)
			}
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if dx*dx+dy*dy <= r*r {
						f := 1.0
						if dy < 0 {
							f = 1.18
						}
						set(cx+dx, cy+dy, pc.shade(f))
					}
				}
			}
		}
	case BLog:
		for x := 0; x < TileSize; x++ {
			c := base
			if x%4 == 0 {
				c = base.shade(0.68)
			} else if x%4 == 2 {
				c = base.shade(1.12)
			}
			for y := 0; y < TileSize; y++ {
				f := 1 + (rng.Float64()-0.5)*0.12
				if rng.Intn(11) == 0 {
					f *= 0.8 // knots
				}
				set(x, y, c.shade(f))
			}
		}
	case BPlanks, BBookshelf:
		for y := 0; y < TileSize; y++ {
			c := base
			if y%4 == 3 {
				c = base.shade(0.58)
			} else if y%4 == 0 {
				c = base.shade(1.08)
			}
			for x := 0; x < TileSize; x++ {
				set(x, y, c.shade(1+(rng.Float64()-0.5)*0.10))
			}
		}
		// Plank end-joints.
		for _, p := range [][3]int{{5, 0, 4}, {11, 4, 8}, {3, 8, 12}, {13, 12, 16}} {
			for y := p[1]; y < p[2]; y++ {
				set(p[0], y, base.shade(0.62))
			}
		}
		if b == BBookshelf {
			for i, c := range []rgb{{170, 50, 50}, {60, 90, 160}, {80, 140, 70}, {180, 150, 60}} {
				for y := 5; y < 11; y++ {
					set(3+i*3, y, c)
					set(4+i*3, y, c.shade(0.8))
				}
			}
		}
	case BLeaves:
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				f := rng.Float64()
				switch {
				case f < 0.12:
					// gaps showing darkness behind
					set(x, y, base.shade(0.35))
				case f < 0.30:
					set(x, y, base.shade(0.6))
				case f > 0.85:
					set(x, y, base.shade(1.35))
				default:
					set(x, y, base.shade(0.85+f*0.4))
				}
			}
		}
	case BCoalOre, BIronOre, BCopperOre, BGoldOre, BRedstoneOre, BLapisOre,
		BDiamondOre, BEmeraldOre, BDeepCoalOre, BDeepIronOre, BDeepGoldOre,
		BDeepRedstoneOre, BDeepLapisOre, BDeepDiamondOre, BDeepEmeraldOre:
		noiseFill(base, 0.14)
		softShade(0.08)
		sp := oreSpeck[b]
		for i := 0; i < 4; i++ {
			x, y := 1+rng.Intn(12), 1+rng.Intn(12)
			// Diamond-shaped nuggets with a specular glint.
			set(x+1, y, sp)
			set(x, y+1, sp.shade(0.85))
			set(x+2, y+1, sp.shade(0.85))
			set(x+1, y+2, sp.shade(0.7))
			set(x+1, y+1, sp.shade(1.25))
			set(x, y, sp.shade(1.5))
		}
	case BWater, BLava:
		noiseFill(base, 0.12) // static fallback; animation frames drawn separately
	case BTorch:
		for y := 5; y < 14; y++ {
			set(7, y, rgb{110, 85, 50})
			set(8, y, rgb{130, 100, 60})
		}
		set(7, 4, rgb{255, 200, 60})
		set(8, 4, rgb{255, 170, 40})
		set(7, 3, rgb{255, 240, 160})
		set(8, 3, rgb{255, 240, 160})
		set(7, 2, rgb{255, 250, 210})
		set(8, 2, rgb{255, 250, 210})
	case BLadder:
		for y := 0; y < TileSize; y++ {
			set(2, y, rgb{150, 118, 70})
			set(3, y, rgb{130, 100, 60})
			set(12, y, rgb{150, 118, 70})
			set(13, y, rgb{130, 100, 60})
		}
		for _, y := range []int{2, 7, 12} {
			for x := 4; x < 12; x++ {
				set(x, y, rgb{170, 136, 84})
				set(x, y+1, rgb{130, 100, 60})
			}
		}
	case BFence:
		for y := 0; y < TileSize; y++ {
			set(7, y, base.shade(1.05))
			set(8, y, base.shade(0.78))
		}
		for _, y := range []int{3, 9} {
			for x := 0; x < TileSize; x++ {
				set(x, y, base.shade(1.0))
				set(x, y+1, base.shade(0.72))
			}
		}
	case BRail:
		for x := 0; x < TileSize; x++ {
			if x%4 < 3 {
				set(x, 11, rgb{120, 95, 70})
				set(x, 12, rgb{100, 78, 58})
			}
		}
		for x := 0; x < TileSize; x++ {
			set(x, 9, rgb{190, 190, 200})
			set(x, 10, rgb{140, 140, 150})
			set(x, 13, rgb{190, 190, 200})
			set(x, 14, rgb{140, 140, 150})
		}
	case BGlass, BIce:
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				set(x, y, base.shade(0.97+rng.Float64()*0.06))
			}
		}
		for i := 0; i < TileSize; i++ {
			set(i, 0, base.shade(1.25))
			set(i, 15, base.shade(0.82))
			set(0, i, base.shade(1.2))
			set(15, i, base.shade(0.85))
		}
		// Diagonal shine streak.
		for i := 0; i < 7; i++ {
			set(3+i, 9-i, rgb{255, 255, 255})
			set(4+i, 9-i, base.shade(1.35))
		}
	case BChest:
		noiseFill(base, 0.10)
		for x := 0; x < TileSize; x++ {
			set(x, 6, base.shade(0.5))
			set(x, 5, base.shade(1.12))
		}
		for _, p := range [][2]int{{7, 5}, {8, 5}, {7, 6}, {8, 6}, {7, 7}, {8, 7}} {
			set(p[0], p[1], rgb{200, 175, 90})
		}
		set(7, 6, rgb{120, 100, 50})
	case BCraftTable:
		noiseFill(base, 0.10)
		for x := 0; x < TileSize; x++ {
			set(x, 0, rgb{175, 140, 88})
			set(x, 1, rgb{165, 130, 80})
			set(x, 2, rgb{120, 92, 55})
		}
		// Tool motifs.
		set(4, 6, rgb{200, 200, 205})
		set(5, 7, rgb{90, 70, 45})
		set(11, 8, rgb{200, 200, 205})
		set(10, 9, rgb{90, 70, 45})
	case BFurnace:
		noiseFill(base, 0.12)
		softShade(0.08)
		for y := 8; y < 14; y++ {
			for x := 5; x < 11; x++ {
				set(x, y, rgb{25, 25, 25})
			}
		}
		set(6, 11, rgb{255, 190, 60})
		set(7, 10, rgb{255, 140, 30})
		set(8, 11, rgb{255, 190, 60})
		set(9, 12, rgb{230, 100, 25})
	case BSpawner:
		noiseFill(base, 0.18)
		for i := 0; i < TileSize; i += 3 {
			for j := 0; j < TileSize; j++ {
				set(i, j, rgb{18, 26, 36})
				set(j, i, rgb{18, 26, 36})
			}
		}
		set(7, 7, rgb{120, 160, 230})
		set(8, 8, rgb{120, 160, 230})
	case BTNT:
		noiseFill(base, 0.08)
		for x := 0; x < TileSize; x++ {
			for y := 6; y < 10; y++ {
				set(x, y, rgb{235, 235, 225})
			}
			set(x, 0, base.shade(0.8))
		}
		// "TNT" hint marks.
		for _, x := range []int{4, 7, 8, 11} {
			for y := 7; y < 9; y++ {
				set(x, y, rgb{40, 40, 40})
			}
		}
	case BTallGrass:
		for i := 0; i < 6; i++ {
			x := 2 + i*2 + rng.Intn(2)
			h := 6 + rng.Intn(8)
			for y := TileSize - h; y < TileSize; y++ {
				f := 0.7 + 0.6*float64(y-(TileSize-h))/float64(h)
				set(x, y, base.shade(f))
			}
		}
	case BFlowerYellow, BFlowerRed:
		for y := 8; y < 16; y++ {
			set(7, y, rgb{80, 140, 60})
		}
		set(9, 11, rgb{80, 140, 60})
		fc := base
		for _, p := range [][2]int{{7, 4}, {6, 5}, {8, 5}, {7, 6}, {5, 5}, {9, 5}, {7, 3}} {
			set(p[0], p[1], fc.shade(0.9+rng.Float64()*0.25))
		}
		set(7, 5, rgb{250, 235, 140})
	case BBrownMushroom, BRedMushroom:
		for y := 10; y < 16; y++ {
			set(7, y, rgb{225, 220, 205})
			set(8, y, rgb{200, 195, 180})
		}
		for x := 4; x < 12; x++ {
			set(x, 9, base)
			set(x, 8, base.shade(1.12))
		}
		for x := 5; x < 11; x++ {
			set(x, 7, base.shade(1.2))
		}
		if b == BRedMushroom {
			set(6, 8, rgb{240, 235, 225})
			set(9, 7, rgb{240, 235, 225})
		}
	case BGlowBerries:
		for y := 0; y < 12; y++ {
			set(7, y, rgb{80, 120, 50})
			if y%3 == 1 {
				set(6, y, rgb{70, 110, 45})
			}
		}
		for _, p := range [][2]int{{6, 4}, {9, 7}, {6, 10}} {
			set(p[0], p[1], rgb{255, 220, 110})
			set(p[0]+1, p[1], rgb{255, 190, 60})
			set(p[0], p[1]+1, rgb{235, 160, 45})
		}
	case BAmethystCluster:
		for i := 0; i < 5; i++ {
			x := 2 + rng.Intn(12)
			h := 4 + rng.Intn(6)
			for y := TileSize - h; y < TileSize; y++ {
				set(x, y, base.shade(0.8+rng.Float64()*0.35))
				if y > TileSize-h+1 {
					set(x+1, y, base.shade(0.6))
				}
			}
			set(x, TileSize-h-1, rgb{235, 215, 255})
		}
	case BDripstone:
		for i := 0; i < 3; i++ {
			x := 2 + i*5 + rng.Intn(2)
			h := 6 + rng.Intn(9)
			for y := 0; y < h; y++ {
				w := 1
				if y < h/2 {
					w = 2
				}
				for dx := 0; dx < w; dx++ {
					set(x+dx, y, base.shade(0.8+rng.Float64()*0.35))
				}
			}
			set(x, h-1, base.shade(1.2))
		}
	case BSculkSensor:
		noiseFill(base, 0.22)
		for _, p := range [][2]int{{4, 4}, {11, 4}} {
			set(p[0], p[1], rgb{110, 235, 250})
			set(p[0], p[1]+1, rgb{60, 170, 190})
			set(p[0], p[1]-1, rgb{60, 170, 190})
		}
	case BSculk:
		noiseFill(base, 0.30)
		softShade(0.15)
		for i := 0; i < 4; i++ {
			x, y := 2+rng.Intn(12), 2+rng.Intn(12)
			set(x, y, rgb{60, 200, 220})
			set(x+1, y, rgb{35, 120, 140})
		}
	case BGlowstone:
		noiseFill(base, 0.16)
		for i := 0; i < 7; i++ {
			x, y := 1+rng.Intn(13), 1+rng.Intn(13)
			set(x, y, rgb{255, 245, 190})
			set(x+1, y, rgb{250, 225, 150})
			set(x, y+1, rgb{250, 225, 150})
		}
	case BStoneBricks, BMossyStoneBricks:
		noiseFill(base, 0.08)
		mortar := base.shade(0.55)
		for y := 0; y < TileSize; y += 4 {
			for x := 0; x < TileSize; x++ {
				set(x, y, mortar)
			}
			off := (y / 4 % 2) * 4
			for yy := y + 1; yy < y+4 && yy < TileSize; yy++ {
				set(off, yy, mortar)
				set((off+8)%TileSize, yy, mortar)
			}
			// Brick top highlight.
			for x := 0; x < TileSize; x++ {
				if x != off && x != (off+8)%TileSize {
					set(x, y+1, base.shade(1.12))
				}
			}
		}
		if b == BMossyStoneBricks {
			for i := 0; i < 8; i++ {
				set(rng.Intn(16), rng.Intn(16), rgb{95, 140, 80})
			}
		}
	case BDoor:
		for y := 0; y < TileSize; y++ {
			for x := 4; x < 12; x++ {
				c := base
				if x == 4 || x == 11 {
					c = base.shade(0.7)
				} else if x == 5 {
					c = base.shade(1.1)
				}
				set(x, y, c)
			}
		}
		for _, y := range []int{5, 10} {
			for x := 5; x < 11; x++ {
				set(x, y, base.shade(0.72))
			}
		}
		set(10, 8, rgb{220, 200, 120})
	case BCactus:
		noiseFill(base, 0.16)
		for y := 0; y < TileSize; y++ {
			set(0, y, base.shade(0.55))
			set(1, y, base.shade(1.2))
			set(15, y, base.shade(0.55))
			set(14, y, base.shade(0.8))
		}
		for i := 0; i < 5; i++ {
			set(2+rng.Intn(11), 1+rng.Intn(14), rgb{215, 230, 180})
		}
	case BObsidian:
		noiseFill(base, 0.25)
		softShade(0.2)
		for i := 0; i < 5; i++ {
			x, y := 1+rng.Intn(14), 1+rng.Intn(14)
			set(x, y, rgb{100, 65, 145})
			set(x+1, y+1, rgb{60, 40, 95})
		}
	case BBedrock:
		noiseFill(base, 0.45)
		softShade(0.2)
	default:
		noiseFill(base, 0.20)
		softShade(0.08)
	}

	if !noBevel[b] {
		t.bevel()
	}
	return ebiten.NewImageFromImage(t.img)
}

// makeLiquidTex builds one animation frame for water or lava.
func makeLiquidTex(b Block, frame int) *ebiten.Image {
	t := newTexBuf()
	rng := rand.New(rand.NewSource(int64(b)*31 + 7))
	base := blockBase[b]
	ph := float64(frame) * math.Pi / 2
	for y := 0; y < TileSize; y++ {
		for x := 0; x < TileSize; x++ {
			var f float64
			if b == BWater {
				// Rolling horizontal waves.
				f = 0.9 +
					0.12*math.Sin(float64(x)*0.8+float64(y)*0.5+ph) +
					0.08*math.Sin(float64(y)*1.1-ph*1.7) +
					(rng.Float64()-0.5)*0.06
			} else {
				// Slowly churning hot blobs.
				f = 0.9 +
					0.2*math.Sin(float64(x)*0.55+ph)*math.Cos(float64(y)*0.5-ph*0.8) +
					(rng.Float64()-0.5)*0.1
			}
			t.set(x, y, base.shade(f))
		}
	}
	if b == BLava {
		// Bright molten cracks drifting with the phase.
		for i := 0; i < 5; i++ {
			x := (rng.Intn(14) + frame*3) % 14
			y := rng.Intn(14)
			t.set(x, y, rgb{255, 225, 110})
			t.set(x+1, y, rgb{255, 180, 60})
			t.set(x, y+1, rgb{255, 150, 45})
		}
	} else {
		// Sparkling highlights on the surface rows.
		for x := 0; x < TileSize; x++ {
			if (x+frame*2)%5 == 0 {
				t.set(x, 0, base.shade(1.5))
			}
			t.set(x, 0, rgb{
				uint8(math.Min(255, float64(base.r)*1.3)),
				uint8(math.Min(255, float64(base.g)*1.3)),
				uint8(math.Min(255, float64(base.b)*1.25)),
			})
		}
	}
	return ebiten.NewImageFromImage(t.img)
}

// liquidTexFor picks the animation frame for a liquid tile.
func liquidTexFor(b Block, time float64, x, y int) *ebiten.Image {
	f := (int(time*3) + (x*5+y*3)&3) & 3
	if b == BWater {
		return waterFrames[f]
	}
	return lavaFrames[f]
}

// buildClouds makes soft blobby cloud sprites.
func buildClouds() {
	for i := range cloudTex {
		w, h := 100+i*40, 26+i*6
		img := image.NewNRGBA(image.Rect(0, 0, w, h))
		rng := rand.New(rand.NewSource(int64(i)*997 + 5))
		// Union of soft ellipses.
		type blob struct{ cx, cy, rx, ry float64 }
		var blobs []blob
		n := 5 + i*2
		for j := 0; j < n; j++ {
			blobs = append(blobs, blob{
				cx: float64(w)*0.15 + rng.Float64()*float64(w)*0.7,
				cy: float64(h)*0.45 + rng.Float64()*float64(h)*0.35,
				rx: float64(w) * (0.12 + rng.Float64()*0.14),
				ry: float64(h) * (0.25 + rng.Float64()*0.25),
			})
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				a := 0.0
				for _, bl := range blobs {
					dx := (float64(x) - bl.cx) / bl.rx
					dy := (float64(y) - bl.cy) / bl.ry
					d := dx*dx + dy*dy
					if d < 1 {
						v := 1 - d
						if v > a {
							a = v
						}
					}
				}
				if a > 0 {
					// Flat-bottomed, brighter on top.
					br := 235 + int(20*(1-float64(y)/float64(h)))
					if br > 255 {
						br = 255
					}
					img.SetNRGBA(x, y, color.NRGBA{uint8(br), uint8(br), 255, uint8(math.Min(1, a*1.6) * 210)})
				}
			}
		}
		cloudTex[i] = ebiten.NewImageFromImage(img)
	}
}

// buildItemTextures makes icons for non-block items.
func buildItemTextures() {
	itemColor := map[Item]rgb{
		IStick: {130, 100, 60}, ICoal: {45, 45, 45}, IRawIron: {216, 175, 147},
		IIronIngot: {220, 220, 225}, IRawGold: {250, 220, 100}, IGoldIngot: {250, 215, 70},
		IRawCopper: {200, 115, 75}, ICopperIngot: {210, 125, 80}, IDiamond: {95, 230, 220},
		IEmerald: {60, 210, 100}, IRedstone: {220, 40, 40}, ILapis: {40, 70, 200},
		IAmethystShard: {170, 125, 225}, IGlowstoneDust: {245, 215, 120}, IXPGem: {40, 190, 210},
		IClayBall: {160, 165, 175}, IBrick: {170, 90, 70}, ISnowball: {240, 245, 250},
		IGlowBerry: {255, 200, 80}, IApple: {210, 50, 45}, IGoldenApple: {250, 210, 60},
		IBread: {195, 150, 85}, IWheatSeeds: {120, 180, 70}, IPorkchop: {235, 160, 160},
		ICookedPorkchop: {190, 120, 80}, IBeef: {200, 80, 70}, ISteak: {150, 90, 55},
		IChicken: {235, 200, 180}, ICookedChicken: {200, 140, 80},
		IBrownMushroomItem: {145, 105, 70}, IRedMushroomItem: {190, 55, 45},
		IMushroomStew: {170, 120, 85}, IBone: {230, 230, 215}, IString: {235, 235, 235},
		IGunpowder: {90, 90, 90}, IRottenFlesh: {130, 90, 60}, ISpiderEye: {150, 40, 60},
		IEnderPearl: {20, 120, 110}, ISlimeball: {110, 200, 90}, IFeather: {235, 235, 240},
		ILeather: {150, 90, 50}, IArrow: {180, 180, 180}, IBow: {130, 100, 60},
		IBucket: {190, 190, 195}, IWaterBucket: {80, 110, 200}, ILavaBucket: {230, 110, 30},
		IFlintSteel: {140, 140, 145}, ICompass: {200, 60, 60}, ICobbledDeepslate: {70, 70, 75},
	}
	toolHead := map[Item]rgb{
		IWoodPickaxe: {157, 128, 79}, IStonePickaxe: {125, 125, 125}, IIronPickaxe: {220, 220, 225},
		IGoldPickaxe: {250, 215, 70}, IDiamondPickaxe: {95, 230, 220},
		IWoodAxe: {157, 128, 79}, IStoneAxe: {125, 125, 125}, IIronAxe: {220, 220, 225},
		IGoldAxe: {250, 215, 70}, IDiamondAxe: {95, 230, 220},
		IWoodShovel: {157, 128, 79}, IStoneShovel: {125, 125, 125}, IIronShovel: {220, 220, 225},
		IGoldShovel: {250, 215, 70}, IDiamondShovel: {95, 230, 220},
		IWoodSword: {157, 128, 79}, IStoneSword: {125, 125, 125}, IIronSword: {220, 220, 225},
		IGoldSword: {250, 215, 70}, IDiamondSword: {95, 230, 220},
		ILeatherArmor: {150, 90, 50}, IIronArmor: {220, 220, 225}, IDiamondArmor: {95, 230, 220},
	}

	for it, c := range itemColor {
		t := newTexBuf()
		rng := rand.New(rand.NewSource(int64(it) * 104729))
		// A shaded round blob with rim light and highlight.
		for y := 3; y < 14; y++ {
			for x := 3; x < 14; x++ {
				dx, dy := float64(x)-8, float64(y)-8
				d := dx*dx + dy*dy
				if d <= 24 {
					f := 1.05 - d/44 + (rng.Float64()-0.5)*0.12
					// Light from upper-left.
					f += (-dx - dy) * 0.022
					t.set(x, y, c.shade(f))
				}
			}
		}
		t.set(6, 6, rgb{255, 255, 255})
		t.set(7, 6, c.shade(1.5))
		t.set(6, 7, c.shade(1.5))
		itemTex[it] = ebiten.NewImageFromImage(t.img)
	}
	for it, c := range toolHead {
		t := newTexBuf()
		handle := rgb{120, 90, 55}
		inf := it.Info()
		px := t.set
		switch inf.Tool {
		case ToolPickaxe:
			for i := 0; i < 8; i++ {
				px(4+i, 12-i, handle)
				px(5+i, 12-i, handle.shade(0.8))
			}
			for i := 0; i < 8; i++ {
				px(3+i, 3, c)
				px(3+i, 4, c.shade(0.8))
			}
			px(2, 4, c)
			px(2, 5, c.shade(0.8))
			px(11, 4, c)
			px(11, 5, c.shade(0.8))
		case ToolAxe:
			for i := 0; i < 9; i++ {
				px(4+i, 13-i, handle)
				px(5+i, 13-i, handle.shade(0.8))
			}
			for y := 2; y < 7; y++ {
				for x := 8; x < 13; x++ {
					f := 1.0
					if x == 8 {
						f = 1.2
					}
					px(x, y, c.shade(f))
				}
			}
		case ToolShovel:
			for i := 0; i < 9; i++ {
				px(4+i, 13-i, handle)
				px(5+i, 13-i, handle.shade(0.8))
			}
			for y := 2; y < 6; y++ {
				for x := 9; x < 13; x++ {
					px(x, y, c.shade(1.05-float64(y-2)*0.08))
				}
			}
		case ToolSword:
			for i := 0; i < 9; i++ {
				px(5+i, 12-i, c)
				px(6+i, 12-i, c.shade(0.75))
			}
			px(13, 3, rgb{255, 255, 255})
			px(5, 11, rgb{90, 70, 45})
			px(4, 12, rgb{90, 70, 45})
			px(6, 13, rgb{90, 70, 45})
			px(3, 13, rgb{70, 55, 35})
		default: // armor
			for y := 4; y < 13; y++ {
				for x := 4; x < 12; x++ {
					if y == 4 && x > 6 && x < 9 {
						continue
					}
					f := 1.0 - float64(y-4)*0.03
					if x == 4 || x == 11 {
						f *= 0.85
					}
					px(x, y, c.shade(f))
				}
			}
			px(5, 5, c.shade(1.35))
		}
		itemTex[it] = ebiten.NewImageFromImage(t.img)
	}
}

// texForItem returns an icon: item texture or its block texture.
func texForItem(i Item) *ebiten.Image {
	if t, ok := itemTex[i]; ok {
		return t
	}
	if i.IsBlock() {
		return blockTex[Block(i)]
	}
	return whiteTex
}
