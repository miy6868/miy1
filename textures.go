package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Procedurally generated 16x16 textures for every block, plus item icons.

var (
	blockTex [BBlockCount]*ebiten.Image
	itemTex  = map[Item]*ebiten.Image{}
	whiteTex *ebiten.Image
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

func blockAvgColor(b Block) (uint8, uint8, uint8) {
	c, ok := blockBase[b]
	if !ok {
		return 128, 128, 128
	}
	return c.r, c.g, c.b
}

// buildTextures creates all block and item textures once at startup.
func buildTextures() {
	whiteTex = ebiten.NewImage(1, 1)
	whiteTex.Fill(color.White)

	for b := Block(1); b < BBlockCount; b++ {
		blockTex[b] = makeBlockTex(b)
	}
	buildItemTextures()
}

func makeBlockTex(b Block) *ebiten.Image {
	img := ebiten.NewImage(TileSize, TileSize)
	rng := rand.New(rand.NewSource(int64(b) * 7919))
	base, ok := blockBase[b]
	if !ok {
		base = rgb{128, 128, 128}
	}
	set := func(x, y int, c rgb) {
		img.Set(x, y, color.RGBA{c.r, c.g, c.b, 255})
	}
	noiseFill := func(c rgb, amount float64) {
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				f := 1 + (rng.Float64()-0.5)*amount
				set(x, y, c.shade(f))
			}
		}
	}

	switch b {
	case BGrass:
		noiseFill(rgb{134, 96, 67}, 0.25)
		for x := 0; x < TileSize; x++ {
			d := 3 + rng.Intn(3)
			for y := 0; y < d; y++ {
				set(x, y, rgb{106, 170, 64}.shade(1 + (rng.Float64()-0.5)*0.3))
			}
		}
	case BSnow:
		noiseFill(rgb{134, 96, 67}, 0.25)
		for x := 0; x < TileSize; x++ {
			d := 4 + rng.Intn(2)
			for y := 0; y < d; y++ {
				set(x, y, rgb{238, 242, 248}.shade(1+(rng.Float64()-0.5)*0.1))
			}
		}
	case BLog:
		for x := 0; x < TileSize; x++ {
			c := base
			if x%4 == 0 {
				c = base.shade(0.7)
			}
			for y := 0; y < TileSize; y++ {
				set(x, y, c.shade(1+(rng.Float64()-0.5)*0.15))
			}
		}
	case BPlanks, BBookshelf:
		for y := 0; y < TileSize; y++ {
			c := base
			if y%4 == 3 {
				c = base.shade(0.6)
			}
			for x := 0; x < TileSize; x++ {
				set(x, y, c.shade(1+(rng.Float64()-0.5)*0.12))
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
				if f < 0.15 {
					set(x, y, base.shade(0.5))
				} else {
					set(x, y, base.shade(0.8+f*0.5))
				}
			}
		}
	case BCoalOre, BIronOre, BCopperOre, BGoldOre, BRedstoneOre, BLapisOre,
		BDiamondOre, BEmeraldOre, BDeepCoalOre, BDeepIronOre, BDeepGoldOre,
		BDeepRedstoneOre, BDeepLapisOre, BDeepDiamondOre, BDeepEmeraldOre:
		noiseFill(base, 0.2)
		sp := oreSpeck[b]
		for i := 0; i < 5; i++ {
			x, y := 1+rng.Intn(13), 1+rng.Intn(13)
			set(x, y, sp)
			set(x+1, y, sp.shade(0.8))
			set(x, y+1, sp.shade(0.7))
			if rng.Intn(2) == 0 {
				set(x+1, y+1, sp.shade(0.9))
			}
		}
	case BWater:
		noiseFill(base, 0.15)
		for x := 0; x < TileSize; x++ {
			set(x, 0, base.shade(1.3))
		}
	case BLava:
		noiseFill(base, 0.3)
		for i := 0; i < 4; i++ {
			x, y := rng.Intn(14), rng.Intn(14)
			set(x, y, rgb{255, 220, 90})
			set(x+1, y, rgb{255, 180, 60})
		}
	case BTorch:
		// transparent background
		img.Clear()
		for y := 5; y < 14; y++ {
			set(7, y, rgb{110, 85, 50})
			set(8, y, rgb{130, 100, 60})
		}
		set(7, 4, rgb{255, 220, 100})
		set(8, 4, rgb{255, 190, 60})
		set(7, 3, rgb{255, 240, 160})
		set(8, 3, rgb{255, 240, 160})
	case BLadder:
		img.Clear()
		for y := 0; y < TileSize; y++ {
			set(2, y, rgb{150, 118, 70})
			set(3, y, rgb{130, 100, 60})
			set(12, y, rgb{150, 118, 70})
			set(13, y, rgb{130, 100, 60})
		}
		for _, y := range []int{2, 7, 12} {
			for x := 4; x < 12; x++ {
				set(x, y, rgb{160, 128, 78})
			}
		}
	case BFence:
		img.Clear()
		for y := 0; y < TileSize; y++ {
			set(7, y, base)
			set(8, y, base.shade(0.8))
		}
		for _, y := range []int{3, 9} {
			for x := 0; x < TileSize; x++ {
				set(x, y, base.shade(0.9))
			}
		}
	case BRail:
		img.Clear()
		for x := 0; x < TileSize; x++ {
			if x%4 < 3 {
				set(x, 12, rgb{110, 90, 70})
			}
		}
		for x := 0; x < TileSize; x++ {
			set(x, 10, rgb{160, 160, 170})
			set(x, 14, rgb{160, 160, 170})
		}
	case BGlass, BIce:
		noiseFill(base, 0.05)
		for i := 0; i < TileSize; i++ {
			set(i, 0, base.shade(1.2))
			set(i, 15, base.shade(0.85))
			set(0, i, base.shade(1.2))
			set(15, i, base.shade(0.85))
		}
		set(3, 3, rgb{255, 255, 255})
		set(4, 4, rgb{255, 255, 255})
	case BChest:
		noiseFill(base, 0.12)
		for x := 0; x < TileSize; x++ {
			set(x, 6, base.shade(0.6))
		}
		set(7, 6, rgb{120, 120, 130})
		set(8, 6, rgb{120, 120, 130})
		set(7, 7, rgb{120, 120, 130})
		set(8, 7, rgb{120, 120, 130})
	case BCraftTable:
		noiseFill(base, 0.12)
		for x := 0; x < TileSize; x++ {
			set(x, 0, rgb{170, 135, 85})
			set(x, 1, rgb{160, 125, 78})
		}
		set(4, 6, rgb{90, 70, 45})
		set(11, 9, rgb{90, 70, 45})
	case BFurnace:
		noiseFill(base, 0.15)
		for y := 8; y < 14; y++ {
			for x := 5; x < 11; x++ {
				set(x, y, rgb{30, 30, 30})
			}
		}
		set(7, 10, rgb{255, 160, 40})
		set(8, 11, rgb{255, 120, 30})
	case BSpawner:
		noiseFill(base, 0.2)
		for i := 0; i < TileSize; i += 3 {
			for j := 0; j < TileSize; j++ {
				set(i, j, rgb{20, 30, 40})
				set(j, i, rgb{20, 30, 40})
			}
		}
	case BTNT:
		noiseFill(base, 0.1)
		for x := 0; x < TileSize; x++ {
			for y := 6; y < 10; y++ {
				set(x, y, rgb{235, 235, 225})
			}
			set(x, 0, base.shade(0.8))
		}
	case BTallGrass:
		img.Clear()
		for i := 0; i < 6; i++ {
			x := 2 + i*2 + rng.Intn(2)
			h := 6 + rng.Intn(8)
			for y := TileSize - h; y < TileSize; y++ {
				set(x, y, base.shade(0.8+rng.Float64()*0.5))
			}
		}
	case BFlowerYellow, BFlowerRed:
		img.Clear()
		for y := 8; y < 16; y++ {
			set(7, y, rgb{80, 140, 60})
		}
		fc := base
		set(7, 6, fc)
		set(6, 5, fc.shade(0.9))
		set(8, 5, fc.shade(0.9))
		set(7, 4, fc)
		set(7, 5, rgb{240, 220, 120})
	case BBrownMushroom, BRedMushroom:
		img.Clear()
		for y := 10; y < 16; y++ {
			set(7, y, rgb{225, 220, 205})
			set(8, y, rgb{200, 195, 180})
		}
		for x := 4; x < 12; x++ {
			set(x, 9, base)
			set(x, 8, base.shade(1.1))
		}
		for x := 5; x < 11; x++ {
			set(x, 7, base.shade(1.15))
		}
	case BGlowBerries:
		img.Clear()
		for y := 0; y < 12; y++ {
			set(7, y, rgb{80, 120, 50})
		}
		for _, p := range [][2]int{{6, 4}, {9, 7}, {6, 10}} {
			set(p[0], p[1], rgb{255, 210, 90})
			set(p[0]+1, p[1], rgb{255, 190, 60})
		}
	case BAmethystCluster:
		img.Clear()
		for i := 0; i < 5; i++ {
			x := 2 + rng.Intn(12)
			h := 4 + rng.Intn(6)
			for y := TileSize - h; y < TileSize; y++ {
				set(x, y, base.shade(0.9+rng.Float64()*0.4))
			}
			set(x, TileSize-h-1, rgb{230, 205, 255})
		}
	case BDripstone:
		img.Clear()
		for i := 0; i < 3; i++ {
			x := 2 + i*5 + rng.Intn(2)
			h := 6 + rng.Intn(9)
			for y := 0; y < h; y++ {
				w := 1
				if y < h/2 {
					w = 2
				}
				for dx := 0; dx < w; dx++ {
					set(x+dx, y, base.shade(0.85+rng.Float64()*0.3))
				}
			}
		}
	case BSculkSensor:
		noiseFill(base, 0.25)
		set(4, 4, rgb{90, 220, 235})
		set(11, 4, rgb{90, 220, 235})
		set(4, 5, rgb{60, 170, 190})
		set(11, 5, rgb{60, 170, 190})
	case BSculk:
		noiseFill(base, 0.35)
		for i := 0; i < 3; i++ {
			set(2+rng.Intn(12), 2+rng.Intn(12), rgb{60, 200, 220})
		}
	case BGlowstone:
		noiseFill(base, 0.2)
		for i := 0; i < 6; i++ {
			set(1+rng.Intn(14), 1+rng.Intn(14), rgb{255, 240, 170})
		}
	case BStoneBricks, BMossyStoneBricks:
		noiseFill(base, 0.1)
		for y := 0; y < TileSize; y += 4 {
			for x := 0; x < TileSize; x++ {
				set(x, y, base.shade(0.6))
			}
		}
		for y := 0; y < TileSize; y += 4 {
			off := (y / 4 % 2) * 4
			for yy := y; yy < y+4 && yy < TileSize; yy++ {
				set((off+8)%TileSize, yy, base.shade(0.6))
				set(off, yy, base.shade(0.6))
			}
		}
	case BDoor:
		img.Clear()
		for y := 0; y < TileSize; y++ {
			for x := 4; x < 12; x++ {
				c := base
				if x == 4 || x == 11 {
					c = base.shade(0.7)
				}
				set(x, y, c)
			}
		}
		set(10, 8, rgb{60, 50, 35})
	case BCactus:
		noiseFill(base, 0.2)
		for y := 0; y < TileSize; y++ {
			set(0, y, base.shade(0.6))
			set(15, y, base.shade(0.6))
		}
		for i := 0; i < 4; i++ {
			set(2+rng.Intn(12), 2+rng.Intn(12), rgb{200, 220, 160})
		}
	case BObsidian:
		noiseFill(base, 0.3)
		for i := 0; i < 4; i++ {
			set(1+rng.Intn(14), 1+rng.Intn(14), rgb{90, 60, 130})
		}
	default:
		noiseFill(base, 0.22)
	}
	return img
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
		img := ebiten.NewImage(TileSize, TileSize)
		rng := rand.New(rand.NewSource(int64(it) * 104729))
		// A rough blob icon with highlight.
		for y := 4; y < 13; y++ {
			for x := 4; x < 13; x++ {
				dx, dy := x-8, y-8
				if dx*dx+dy*dy <= 18 {
					f := 1 + (rng.Float64()-0.5)*0.2
					img.Set(x, y, color.RGBA{c.shade(f).r, c.shade(f).g, c.shade(f).b, 255})
				}
			}
		}
		img.Set(6, 6, color.RGBA{255, 255, 255, 200})
		itemTex[it] = img
	}
	for it, c := range toolHead {
		img := ebiten.NewImage(TileSize, TileSize)
		handle := rgb{120, 90, 55}
		inf := it.Info()
		px := func(x, y int, cc rgb) { img.Set(x, y, color.RGBA{cc.r, cc.g, cc.b, 255}) }
		switch inf.Tool {
		case ToolPickaxe:
			for i := 0; i < 8; i++ {
				px(4+i, 12-i, handle)
			}
			for i := 0; i < 8; i++ {
				px(3+i, 3, c)
			}
			px(3, 4, c)
			px(2, 5, c)
			px(10, 4, c)
			px(11, 5, c)
		case ToolAxe:
			for i := 0; i < 9; i++ {
				px(4+i, 13-i, handle)
			}
			for y := 2; y < 7; y++ {
				for x := 8; x < 13; x++ {
					px(x, y, c)
				}
			}
		case ToolShovel:
			for i := 0; i < 9; i++ {
				px(4+i, 13-i, handle)
			}
			for y := 2; y < 6; y++ {
				for x := 10; x < 14; x++ {
					px(x-1, y, c)
				}
			}
		case ToolSword:
			for i := 0; i < 9; i++ {
				px(5+i, 12-i, c)
			}
			px(5, 11, rgb{90, 70, 45})
			px(4, 12, rgb{90, 70, 45})
			px(6, 13, rgb{90, 70, 45})
		default: // armor
			for y := 4; y < 13; y++ {
				for x := 4; x < 12; x++ {
					if y == 4 && x > 6 && x < 9 {
						continue
					}
					px(x, y, c)
				}
			}
		}
		itemTex[it] = img
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
