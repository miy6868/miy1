package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenW = 960
	ScreenH = 540
)

// UIMode is which overlay panel is open.
type UIMode int

const (
	UINone UIMode = iota
	UIInventory
	UICraftTable
	UIFurnace
	UIChest
)

// Game is the top-level state.
type Game struct {
	World  World
	Player Player
	rng    *rand.Rand

	SpawnX     int
	CamX, CamY float64

	Mode     UIMode
	ChestKey [2]int
	Carried  ItemStack // stack held by the mouse cursor in UIs
	craftScroll int

	CursorX, CursorY int
	CursorReach      bool

	Arrows []*Arrow

	liquidAcc float64
	spawnAcc  float64
	tickAcc   float64
	glows     []glowPoint
	saveMsg   string
	msgTimer  float64

	Paused bool
}

func NewGame(seed int64) *Game {
	g := &Game{rng: rand.New(rand.NewSource(seed ^ 0x5eed))}
	g.World = *NewWorld(seed)
	g.SpawnX = 0
	// Find a dry-ish spawn column.
	for x := 0; x < 400; x += 7 {
		if g.World.SurfaceY(x) < SeaLevel {
			g.SpawnX = x
			break
		}
	}
	g.Player = NewPlayer(0, 0)
	g.Player.Respawn(g)
	g.CamX = g.Player.X
	g.CamY = g.Player.Y

	// Starter kit.
	g.Player.Give(ItemStack{IWoodPickaxe, 1})
	g.Player.Give(ItemStack{IWoodSword, 1})
	g.Player.Give(ItemStack{ITorch, 12})
	g.Player.Give(ItemStack{IBread, 3})
	return g
}

func (g *Game) UIOpen() bool { return g.Mode != UINone }

// TeleportToCave moves the player into the nearest open cave at the given
// depth (debug helper for exploring the underground).
func (g *Game) TeleportToCave(depth int) {
	w := &g.World
	for x := g.SpawnX; x < g.SpawnX+2000; x++ {
		for y := depth - 25; y < depth+25; y++ {
			if w.Block(x, y) == BAir && w.Block(x, y-1) == BAir && w.Block(x, y+1).Solid() {
				g.Player.X = float64(x) + 0.125
				g.Player.Y = float64(y) - g.Player.H + 0.9
				g.Player.VX, g.Player.VY = 0, 0
				g.CamX, g.CamY = g.Player.X, g.Player.Y
				g.Player.Give(ItemStack{ITorch, 64})
				g.Player.Give(ItemStack{IIronPickaxe, 1})
				return
			}
		}
	}
}

// MouseWorld converts the cursor to world block coordinates.
func (g *Game) MouseWorld() (float64, float64) {
	mx, my := ebiten.CursorPosition()
	wx := g.CamX + (float64(mx)-ScreenW/2)/TileSize
	wy := g.CamY + (float64(my)-ScreenH/2)/TileSize
	return wx, wy
}

func (g *Game) OpenCrafting(table bool) {
	if table {
		g.Mode = UICraftTable
	} else {
		g.Mode = UIInventory
	}
	g.craftScroll = 0
}

func (g *Game) OpenFurnace() { g.Mode = UIFurnace; g.craftScroll = 0 }

func (g *Game) OpenChest(bx, by int) {
	g.ChestKey = [2]int{bx, by}
	g.Mode = UIChest
}

func (g *Game) Message(s string) { g.saveMsg = s; g.msgTimer = 3 }

// Update runs one frame (60 TPS).
func (g *Game) Update() error {
	dt := 1.0 / 60.0

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.Mode != UINone {
			g.Mode = UINone
			g.dropCarried()
		} else {
			g.Paused = !g.Paused
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		if g.Mode == UINone {
			g.Mode = UIInventory
			g.craftScroll = 0
		} else {
			g.Mode = UINone
			g.dropCarried()
		}
	}
	if g.Paused {
		return nil
	}

	g.World.Time += dt
	if g.World.Time >= DayLength {
		g.World.Time -= DayLength
	}
	g.msgTimer = math.Max(0, g.msgTimer-dt)

	g.UpdatePlayer(dt)
	if g.UIOpen() {
		g.updateUI()
	}

	// Camera follows with slight smoothing.
	tx := g.Player.X + g.Player.W/2
	ty := g.Player.Y + g.Player.H/2
	g.CamX += (tx - g.CamX) * math.Min(1, dt*10)
	g.CamY += (ty - g.CamY) * math.Min(1, dt*10)
	// Clamp vertical so the view stays inside the world.
	half := float64(ScreenH) / TileSize / 2
	if g.CamY < half {
		g.CamY = half
	}
	if g.CamY > float64(WorldH)-half {
		g.CamY = float64(WorldH) - half
	}

	// Ensure visible chunks exist, then light them.
	cx0 := floorDiv(int(math.Floor(g.CamX))-ScreenW/TileSize/2-8, ChunkW)
	cx1 := floorDiv(int(math.Floor(g.CamX))+ScreenW/TileSize/2+8, ChunkW)
	for cx := cx0; cx <= cx1; cx++ {
		g.World.ChunkAt(cx)
	}
	g.World.RecomputeLight(cx0, cx1)

	// Simulation ticks.
	g.liquidAcc += dt
	if g.liquidAcc >= 0.18 {
		g.liquidAcc = 0
		g.tickLiquids()
	}
	g.spawnAcc += dt
	if g.spawnAcc >= 1.5 {
		g.spawnAcc = 0
		g.trySpawnMobs()
		g.tickSpawnerBlocks()
	}
	g.tickAcc += dt
	if g.tickAcc >= 0.8 {
		g.tickAcc = 0
		g.randomTicks()
	}

	g.updateMobs(dt)
	g.updateDrops(dt)
	g.updateArrows(dt)
	g.updateParticles(dt)
	g.ambientParticles()
	return nil
}

func (g *Game) dropCarried() {
	if !g.Carried.Empty() {
		g.Player.Give(g.Carried)
		g.Carried = ItemStack{}
	}
}

// ---- Simulation subsystems ----

// tickLiquids runs simple finite-fluid flow near the player.
func (g *Game) tickLiquids() {
	w := &g.World
	px := int(g.Player.X)
	r := ScreenW/TileSize/2 + 12
	for y := WorldH - 2; y >= 1; y-- {
		for x := px - r; x <= px+r; x++ {
			b := w.Block(x, y)
			if !b.Liquid() {
				continue
			}
			// Lava + water interactions.
			if b == BLava {
				for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					if w.Block(x+d[0], y+d[1]) == BWater {
						w.SetBlock(x, y, BObsidian)
						break
					}
				}
				if w.Block(x, y) != BLava {
					continue
				}
				if g.rng.Intn(3) != 0 {
					continue // lava flows slower
				}
			}
			below := w.Block(x, y+1)
			if below == BAir {
				w.SetBlock(x, y+1, b)
				w.SetBlock(x, y, BAir)
				continue
			}
			if b == BWater && below == BLava {
				w.SetBlock(x, y+1, BStone)
				continue
			}
			// Spread sideways on solid ground.
			dir := 1
			if g.rng.Intn(2) == 0 {
				dir = -1
			}
			for _, d := range [2]int{dir, -dir} {
				side := w.Block(x+d, y)
				if side == BAir {
					sideBelow := w.Block(x+d, y+1)
					if sideBelow == BAir || sideBelow.Solid() || sideBelow == b {
						w.SetBlock(x+d, y, b)
						w.SetBlock(x, y, BAir)
					}
					break
				}
			}
		}
	}
}

// randomTicks: grass spread, glow berry growth, ice melting near light.
func (g *Game) randomTicks() {
	w := &g.World
	px, py := int(g.Player.X), int(g.Player.Y)
	for i := 0; i < 40; i++ {
		x := px + g.rng.Intn(80) - 40
		y := py + g.rng.Intn(60) - 30
		if y < 1 || y >= WorldH-1 {
			continue
		}
		switch w.Block(x, y) {
		case BDirt:
			// Grow grass if lit and open above, and grass is adjacent.
			if w.Block(x, y-1) == BAir && g.lightLevelAt(x, y-1) > 8 {
				for _, d := range [4][2]int{{1, 0}, {-1, 0}, {1, -1}, {-1, -1}} {
					if w.Block(x+d[0], y+d[1]) == BGrass {
						w.SetBlock(x, y, BGrass)
						break
					}
				}
			}
		case BGrass:
			if w.Block(x, y-1).Opaque() {
				w.SetBlock(x, y, BDirt)
			}
		case BMossBlock:
			if w.Block(x, y+1) == BAir && g.rng.Intn(6) == 0 {
				w.SetBlock(x, y+1, BGlowBerries)
			}
		}
	}
}

// tickSpawnerBlocks activates monster spawners near the player.
func (g *Game) tickSpawnerBlocks() {
	w := &g.World
	px, py := int(g.Player.X), int(g.Player.Y)
	hostiles := 0
	for _, m := range w.Mobs {
		if m.Info().Hostile {
			hostiles++
		}
	}
	if hostiles >= 24 {
		return
	}
	for dx := -12; dx <= 12; dx++ {
		for dy := -10; dy <= 10; dy++ {
			x, y := px+dx, py+dy
			if w.Block(x, y) != BSpawner {
				continue
			}
			if g.rng.Float64() > 0.35 {
				continue
			}
			// Spawn beside the spawner in open space.
			for try := 0; try < 6; try++ {
				sx := x + g.rng.Intn(7) - 3
				sy := y + g.rng.Intn(3) - 1
				if !w.Block(sx, sy).Solid() && !w.Block(sx, sy-1).Solid() && w.Block(sx, sy+1).Solid() {
					kinds := []MobKind{MobZombie, MobZombie, MobSkeleton, MobSpider}
					w.Mobs = append(w.Mobs, NewMob(kinds[g.rng.Intn(len(kinds))], float64(sx)+0.5, float64(sy)+1))
					break
				}
			}
		}
	}
}

func (g *Game) updateMobs(dt float64) {
	w := &g.World
	alive := w.Mobs[:0]
	px := g.Player.X
	for _, m := range w.Mobs {
		m.Update(g, dt)
		if m.HP <= 0 {
			continue
		}
		if math.Abs(m.X-px) > 90 {
			continue // despawn far away
		}
		if m.Y > float64(WorldH)+5 {
			continue
		}
		alive = append(alive, m)
	}
	w.Mobs = alive
}

func (g *Game) updateDrops(dt float64) {
	w := &g.World
	p := &g.Player
	kept := w.Drops[:0]
	for _, d := range w.Drops {
		d.Age += dt
		d.Pickup = math.Max(0, d.Pickup-dt)
		// Magnet toward the player.
		dx := (p.X + p.W/2) - (d.X + d.W/2)
		dy := (p.Y + p.H/2) - (d.Y + d.H/2)
		dist := math.Hypot(dx, dy)
		if d.Pickup <= 0 && dist < 3 && !p.Dead {
			d.VX += dx / (dist + 0.01) * 30 * dt
			d.VY += dy / (dist + 0.01) * 30 * dt
		} else {
			d.VX *= 1 - 3*dt
		}
		d.MoveAndCollide(w, dt)
		if d.InLava {
			continue // burned
		}
		if d.Pickup <= 0 && dist < 1.1 && !p.Dead {
			got := p.Give(d.Stack)
			if got == d.Stack.Count {
				continue
			}
			d.Stack.Count -= got
		}
		if d.Age > 180 || d.Y > float64(WorldH)+5 {
			continue
		}
		kept = append(kept, d)
	}
	w.Drops = kept
}

func (g *Game) updateArrows(dt float64) {
	w := &g.World
	p := &g.Player
	kept := g.Arrows[:0]
	for _, a := range g.Arrows {
		a.Life -= dt
		if a.Life <= 0 {
			continue
		}
		a.VY += gravity * 0.6 * dt
		a.X += a.VX * dt
		a.Y += a.VY * dt
		if w.Block(int(a.X), int(a.Y)).Solid() {
			continue
		}
		hit := false
		if a.FromPlayer {
			for _, m := range w.Mobs {
				if a.X >= m.X && a.X <= m.X+m.W && a.Y >= m.Y && a.Y <= m.Y+m.H {
					kb := 4.0
					if a.VX < 0 {
						kb = -kb
					}
					m.Hurt(g, a.Damage, kb)
					hit = true
					break
				}
			}
		} else if !p.Dead &&
			a.X >= p.X && a.X <= p.X+p.W && a.Y >= p.Y && a.Y <= p.Y+p.H {
			p.TakeDamage(g, a.Damage)
			hit = true
		}
		if hit {
			continue
		}
		kept = append(kept, a)
	}
	g.Arrows = kept
}

func (g *Game) updateParticles(dt float64) {
	w := &g.World
	kept := w.Particles[:0]
	for _, p := range w.Particles {
		p.Life -= dt
		if p.Life <= 0 {
			continue
		}
		p.VY += p.Grav * dt
		p.X += p.VX * dt
		p.Y += p.VY * dt
		kept = append(kept, p)
	}
	w.Particles = kept
}

// ambientParticles sprinkles atmosphere: torch embers, lava sparks, sculk
// motes, amethyst glints, dripping water from dripstone.
func (g *Game) ambientParticles() {
	if len(g.World.Particles) > 220 {
		return
	}
	w := &g.World
	halfW := ScreenW / TileSize / 2
	halfH := ScreenH / TileSize / 2
	for i := 0; i < 6; i++ {
		x := int(g.CamX) + g.rng.Intn(halfW*2+2) - halfW - 1
		y := int(g.CamY) + g.rng.Intn(halfH*2+2) - halfH - 1
		if y < 0 || y >= WorldH {
			continue
		}
		fx := float64(x) + g.rng.Float64()
		fy := float64(y) + g.rng.Float64()
		switch w.Block(x, y) {
		case BTorch:
			if g.rng.Float64() < 0.35 {
				w.Particles = append(w.Particles, &Particle{
					X: float64(x) + 0.5, Y: float64(y) + 0.2,
					VX: (g.rng.Float64() - 0.5) * 0.8, VY: -1.2 - g.rng.Float64(),
					Life: 0.5 + g.rng.Float64()*0.5, Grav: -1,
					R: 255, G: 180, B: 70,
				})
			}
		case BLava:
			if !w.Block(x, y-1).Liquid() && g.rng.Float64() < 0.25 {
				w.Particles = append(w.Particles, &Particle{
					X: fx, Y: float64(y),
					VX: (g.rng.Float64() - 0.5) * 3, VY: -3 - g.rng.Float64()*3,
					Life: 0.4 + g.rng.Float64()*0.4, Grav: gravity * 0.6,
					R: 255, G: 140, B: 40,
				})
			}
		case BSculk, BSculkSensor:
			if g.rng.Float64() < 0.2 {
				w.Particles = append(w.Particles, &Particle{
					X: fx, Y: float64(y) - 0.1,
					VX: (g.rng.Float64() - 0.5) * 0.5, VY: -0.4,
					Life: 0.8 + g.rng.Float64(), Grav: -0.2,
					R: 60, G: 210, B: 230,
				})
			}
		case BAmethystCluster:
			if g.rng.Float64() < 0.2 {
				w.Particles = append(w.Particles, &Particle{
					X: fx, Y: fy,
					VX: 0, VY: -0.2,
					Life: 0.5 + g.rng.Float64()*0.5, Grav: 0,
					R: 210, G: 170, B: 250,
				})
			}
		case BDripstone:
			// Water drips off stalactites.
			if w.Block(x, y+1) == BAir && g.rng.Float64() < 0.12 {
				w.Particles = append(w.Particles, &Particle{
					X: float64(x) + 0.4 + g.rng.Float64()*0.2, Y: float64(y) + 0.9,
					VX: 0, VY: 1,
					Life: 0.9, Grav: gravity * 0.6,
					R: 110, G: 150, B: 220,
				})
			}
		case BGlowBerries:
			if g.rng.Float64() < 0.1 {
				w.Particles = append(w.Particles, &Particle{
					X: fx, Y: fy,
					VX: 0, VY: 0.1,
					Life: 0.7, Grav: 0,
					R: 255, G: 215, B: 110,
				})
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func depthName(y int, surface int) string {
	switch {
	case y < surface+8:
		return "Surface"
	case y < DeepslateY:
		return "Caves"
	case y < DeepDarkY:
		return "Deepslate Depths"
	case y < LavaLevel:
		return "Deep Dark"
	default:
		return "Lava Zone"
	}
}

func fmtCoord(g *Game) string {
	p := &g.Player
	return fmt.Sprintf("X %d  Y %d  %s", int(p.X), int(p.Y), depthName(int(p.Y), g.World.SurfaceY(int(p.X))))
}
