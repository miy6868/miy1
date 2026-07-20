package main

import (
	"math"
	"testing"
)

// TestGeneration checks that generated terrain contains the expected
// content across a wide span of chunks and several seeds.
func TestGeneration(t *testing.T) {
	for _, seed := range []int64{1, 42, -987654321, 20260720} {
		w := NewWorld(seed)
		counts := map[Block]int{}
		for cx := -40; cx <= 40; cx++ {
			c := w.ChunkAt(cx)
			for i := range c.Blocks {
				counts[c.Blocks[i]]++
			}
		}
		must := []Block{
			BStone, BDeepslate, BBedrock, BDirt, BGrass,
			BCoalOre, BIronOre, BGoldOre, BRedstoneOre, BLapisOre,
			BLava, BWater, BMossBlock, BSculk, BAir,
		}
		for _, b := range must {
			if counts[b] == 0 {
				t.Errorf("seed %d: no %s generated in 81 chunks", seed, b.Info().Name)
			}
		}
		// Diamonds/emeralds only exist in the deepslate layer.
		if counts[BDeepDiamondOre] == 0 {
			t.Errorf("seed %d: no diamond ore generated", seed)
		}
		if counts[BDeepEmeraldOre]+counts[BEmeraldOre] == 0 {
			t.Errorf("seed %d: no emerald ore generated", seed)
		}
		// Structures should appear somewhere in this span.
		structure := counts[BChest] + counts[BSpawner] + counts[BAmethyst] + counts[BStoneBricks]
		if structure == 0 {
			t.Errorf("seed %d: no structures generated", seed)
		}
		// Bedrock floor must be unbroken at the very bottom.
		for x := -100; x < 100; x++ {
			if w.Block(x, WorldH-1) != BBedrock {
				t.Fatalf("seed %d: bottom layer not bedrock at x=%d", seed, x)
			}
		}
	}
}

// TestDeterminism: the same seed must generate identical chunks.
func TestDeterminism(t *testing.T) {
	a := NewWorld(777)
	b := NewWorld(777)
	for _, cx := range []int{-5, 0, 13} {
		ca, cb := a.ChunkAt(cx), b.ChunkAt(cx)
		if ca.Blocks != cb.Blocks {
			t.Fatalf("chunk %d differs between identical seeds", cx)
		}
	}
	c := NewWorld(778)
	if a.ChunkAt(0).Blocks == c.ChunkAt(0).Blocks {
		t.Fatal("different seeds produced identical chunk 0")
	}
}

func TestLighting(t *testing.T) {
	g := &Game{}
	g.World = *NewWorld(5)
	g.World.RecomputeLight(-1, 1)
	// Sky above the surface must be fully lit at day.
	sy := g.World.SurfaceY(0)
	c := g.World.ChunkAt(0)
	if c.SkyLight[(sy-3)*ChunkW+0] != 15 {
		t.Errorf("sky light above surface = %d, want 15", c.SkyLight[(sy-3)*ChunkW+0])
	}
	// Deep solid rock must be dark (away from lava).
	dark := false
	for lx := 0; lx < ChunkW && !dark; lx++ {
		for y := 200; y < 240; y++ {
			if c.At(lx, y).Opaque() && c.SkyLight[y*ChunkW+lx] == 0 {
				dark = true
				break
			}
		}
	}
	if !dark {
		t.Error("expected dark blocks deep underground")
	}
}

func TestCrafting(t *testing.T) {
	p := NewPlayer(0, 0)
	p.Give(ItemStack{ILogItem, 2})
	var planks, sticks *Recipe
	for i := range recipes {
		switch recipes[i].Out.Item {
		case IPlanks:
			planks = &recipes[i]
		case IStick:
			sticks = &recipes[i]
		}
	}
	if !p.CanCraft(planks) || !p.Craft(planks) {
		t.Fatal("cannot craft planks from logs")
	}
	if p.Count(IPlanks) != 4 {
		t.Fatalf("planks = %d, want 4", p.Count(IPlanks))
	}
	if !p.Craft(sticks) {
		t.Fatal("cannot craft sticks")
	}
	if p.Count(IStick) != 4 || p.Count(IPlanks) != 2 {
		t.Fatalf("after sticks: sticks=%d planks=%d", p.Count(IStick), p.Count(IPlanks))
	}
	// Furnace recipe requires coal.
	var smelt *Recipe
	for i := range recipes {
		if recipes[i].Out.Item == IIronIngot {
			smelt = &recipes[i]
		}
	}
	p.Give(ItemStack{IRawIron, 1})
	if p.CanCraft(smelt) {
		t.Fatal("smelting should require coal")
	}
	p.Give(ItemStack{ICoal, 1})
	if !p.Craft(smelt) {
		t.Fatal("smelting failed with coal")
	}
	if p.Count(IIronIngot) != 1 || p.Count(ICoal) != 0 {
		t.Fatal("smelting did not consume coal / produce ingot")
	}
}

func TestPhysicsAndBreaking(t *testing.T) {
	g := NewGame(99)
	p := &g.Player
	// Drop the player from above the surface; it must land on solid ground.
	sy := g.World.SurfaceY(g.SpawnX)
	p.X, p.Y = float64(g.SpawnX), float64(sy)-10
	p.VY = 0
	for i := 0; i < 600; i++ {
		p.MoveAndCollide(&g.World, 1.0/60)
	}
	if !p.OnGround {
		t.Fatal("player never landed")
	}
	if p.Y > float64(WorldH) {
		t.Fatal("player fell through the world")
	}

	// Breaking a dirt block drops a dirt item.
	bx, by := g.SpawnX, g.World.SurfaceY(g.SpawnX)+1
	g.World.SetBlock(bx, by, BDirt)
	nd := len(g.World.Drops)
	g.BreakBlock(bx, by, TierNone, false)
	if g.World.Block(bx, by) != BAir {
		t.Fatal("block not removed")
	}
	if len(g.World.Drops) != nd+1 {
		t.Fatal("no drop spawned")
	}

	// Stone without a pickaxe drops nothing.
	g.World.SetBlock(bx, by, BStone)
	nd = len(g.World.Drops)
	g.BreakBlock(bx, by, TierNone, false)
	if len(g.World.Drops) != nd {
		t.Fatal("stone dropped an item without a pickaxe")
	}
}

func TestChestsHaveLoot(t *testing.T) {
	w := NewWorld(2026)
	found := 0
	for cx := -60; cx <= 60; cx++ {
		c := w.ChunkAt(cx)
		for lx := 0; lx < ChunkW; lx++ {
			for y := 0; y < WorldH; y++ {
				if c.At(lx, y) == BChest {
					loot := w.Chests[[2]int{cx*ChunkW + lx, y}]
					if len(loot) == 0 {
						t.Errorf("generated chest at %d,%d has no loot", cx*ChunkW+lx, y)
					}
					found++
				}
			}
		}
	}
	if found == 0 {
		t.Error("no loot chests generated in 121 chunks")
	}
}

func TestExplosion(t *testing.T) {
	g := NewGame(7)
	sy := g.World.SurfaceY(300)
	g.Player.X, g.Player.Y = 1000, 10 // far away, unhurt
	g.Explode(300, float64(sy)+3, 3, 10)
	air := 0
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			if g.World.Block(300+dx, sy+3+dy) == BAir {
				air++
			}
		}
	}
	if air < 5 {
		t.Errorf("explosion carved only %d air blocks", air)
	}
	if math.Abs(g.Player.HP-g.Player.MaxHP) > 0.001 {
		t.Error("distant player took explosion damage")
	}
}
