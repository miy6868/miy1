package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	InvCols  = 9
	InvRows  = 4 // row 0 is the hotbar
	InvSlots = InvCols * InvRows
)

// Player is the controllable character.
type Player struct {
	Entity
	HP        float64
	MaxHP     float64
	Hunger    float64
	Breath    float64
	XP        int
	Dead      bool
	DeadTimer float64
	Facing    float64 // 1 right, -1 left

	Inventory [InvSlots]ItemStack
	Armor     ItemStack
	Hotbar    int

	MineX, MineY int
	MineProgress float64
	SwingTimer   float64
	HurtTimer    float64
	fallStart    float64
	falling      bool

	EatTimer   float64
	regenTick  float64
	starveTick float64
}

func NewPlayer(x, y float64) Player {
	p := Player{
		Entity: Entity{X: x, Y: y, W: 0.75, H: 1.8, StepUp: true},
		HP:     20, MaxHP: 20, Hunger: 20, Breath: 10,
	}
	return p
}

// Held returns the hotbar item stack.
func (p *Player) Held() *ItemStack { return &p.Inventory[p.Hotbar] }

// Give inserts items into the inventory, returning the count that fit.
func (p *Player) Give(s ItemStack) int {
	if s.Empty() {
		return 0
	}
	remaining := s.Count
	max := s.Item.Info().MaxStack
	// Merge into existing stacks first.
	for i := range p.Inventory {
		sl := &p.Inventory[i]
		if sl.Item == s.Item && sl.Count < max {
			take := min(remaining, max-sl.Count)
			sl.Count += take
			remaining -= take
			if remaining == 0 {
				return s.Count
			}
		}
	}
	for i := range p.Inventory {
		sl := &p.Inventory[i]
		if sl.Empty() {
			take := min(remaining, max)
			*sl = ItemStack{s.Item, take}
			remaining -= take
			if remaining == 0 {
				return s.Count
			}
		}
	}
	return s.Count - remaining
}

// Count returns how many of an item the player carries.
func (p *Player) Count(it Item) int {
	n := 0
	for i := range p.Inventory {
		if p.Inventory[i].Item == it {
			n += p.Inventory[i].Count
		}
	}
	return n
}

// Consume removes n items, returning success.
func (p *Player) Consume(it Item, n int) bool {
	if p.Count(it) < n {
		return false
	}
	for i := range p.Inventory {
		sl := &p.Inventory[i]
		if sl.Item == it {
			take := min(n, sl.Count)
			sl.Count -= take
			n -= take
			if sl.Count == 0 {
				sl.Item = INone
			}
			if n == 0 {
				return true
			}
		}
	}
	return true
}

// ToolTier returns the held tool's tier for a given block, and whether the
// held tool kind matches the block's best tool.
func (p *Player) toolFor(b Block) (tier int, matches bool) {
	inf := p.Held().Item.Info()
	bi := b.Info()
	if inf.Tool != ToolNone && inf.Tool == bi.BestTool {
		return inf.Tier, true
	}
	return TierNone, false
}

// TakeDamage applies damage with armor reduction and brief invulnerability.
func (p *Player) TakeDamage(g *Game, dmg float64) {
	if p.Dead || p.HurtTimer > 0 || dmg <= 0 {
		return
	}
	if !p.Armor.Empty() {
		dmg *= 1 - p.Armor.Item.Info().Armor
	}
	p.HP -= dmg
	p.HurtTimer = 0.5
	if p.HP <= 0 {
		p.Die(g)
	}
}

// Die drops the inventory and starts the respawn timer.
func (p *Player) Die(g *Game) {
	p.Dead = true
	p.DeadTimer = 3
	p.HP = 0
	for i := range p.Inventory {
		if !p.Inventory[i].Empty() {
			g.World.Drops = append(g.World.Drops,
				NewItemDrop(p.X+p.W/2, p.Y+p.H/2, p.Inventory[i]))
			p.Inventory[i] = ItemStack{}
		}
	}
	if !p.Armor.Empty() {
		g.World.Drops = append(g.World.Drops, NewItemDrop(p.X+p.W/2, p.Y+p.H/2, p.Armor))
		p.Armor = ItemStack{}
	}
	p.XP = p.XP / 2
}

// Respawn places the player back at spawn.
func (p *Player) Respawn(g *Game) {
	sx := g.SpawnX
	sy := g.World.SurfaceY(sx)
	p.X = float64(sx) + 0.125
	p.Y = float64(sy) - p.H - 0.1
	p.VX, p.VY = 0, 0
	p.HP = p.MaxHP
	p.Hunger = 20
	p.Breath = 10
	p.Dead = false
	p.falling = false
}

// UpdatePlayer handles input, physics, mining and survival mechanics.
func (g *Game) UpdatePlayer(dt float64) {
	p := &g.Player
	w := &g.World

	if p.Dead {
		p.DeadTimer -= dt
		if p.DeadTimer <= 0 {
			p.Respawn(g)
		}
		return
	}

	// ---- Movement input ----
	speed := 4.5
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		speed = 2.0
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		speed = 6.5
	}
	move := 0.0
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		move -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		move += 1
	}
	if move != 0 {
		p.Facing = move
	}
	target := move * speed
	accel := 40.0
	if !p.OnGround {
		accel = 20.0
	}
	if p.VX < target {
		p.VX = math.Min(p.VX+accel*dt, target)
	} else {
		p.VX = math.Max(p.VX-accel*dt, target)
	}

	climbing := w.Block(int(p.X+p.W/2), int(p.Y+p.H/2)).Info().Climbable ||
		w.Block(int(p.X+p.W/2), int(p.Y+p.H-0.1)).Info().Climbable

	jump := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	if jump {
		switch {
		case climbing:
			p.VY = -4
		case p.InWater || p.InLava:
			p.VY = -4.5
		case p.OnGround:
			p.VY = -11.5
		}
	}
	if climbing && !jump && p.VY > 2.5 {
		p.VY = 2.5 // slide slowly down ladders
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		if climbing {
			p.VY = 4
		}
	}

	// Fall damage bookkeeping.
	wasFalling := p.falling
	fallFrom := p.fallStart
	if p.VY > 1 && !p.OnGround && !climbing && !p.InWater {
		if !p.falling {
			p.falling = true
			p.fallStart = p.Y
		}
	}
	p.MoveAndCollide(w, dt)
	if p.OnGround || p.InWater || climbing {
		if wasFalling && p.OnGround {
			fell := p.Y - fallFrom
			if fell > 4.5 {
				p.TakeDamage(g, math.Floor(fell-4))
			}
		}
		p.falling = false
	}

	// ---- Environmental damage ----
	if p.InLava {
		p.HurtTimer = 0 // lava ignores i-frames pacing a bit
		p.TakeDamage(g, 6*dt+1*dt)
		if p.Dead {
			return
		}
	}
	headBlock := w.Block(int(p.X+p.W/2), int(p.Y+0.2))
	if headBlock == BWater {
		p.Breath -= dt
		if p.Breath <= 0 {
			p.Breath = 0
			p.HurtTimer = 0
			p.TakeDamage(g, 2*dt)
		}
	} else {
		p.Breath = math.Min(10, p.Breath+dt*3)
	}
	// Cactus contact.
	for _, off := range [][2]float64{{0, p.H - 0.1}, {p.W, p.H - 0.1}, {p.W / 2, 0}} {
		if w.Block(int(p.X+off[0]), int(p.Y+off[1])) == BCactus {
			p.TakeDamage(g, 1)
			break
		}
	}
	if p.Y > float64(WorldH)+10 {
		p.Die(g)
	}
	if p.Dead {
		return
	}

	// ---- Hunger / regen ----
	drain := 0.012
	if math.Abs(p.VX) > 5 {
		drain = 0.03
	}
	p.Hunger = math.Max(0, p.Hunger-drain*dt)
	if p.Hunger >= 18 && p.HP < p.MaxHP {
		p.regenTick += dt
		if p.regenTick > 2.5 {
			p.regenTick = 0
			p.HP = math.Min(p.MaxHP, p.HP+1)
			p.Hunger = math.Max(0, p.Hunger-0.4)
		}
	}
	if p.Hunger <= 0 {
		p.starveTick += dt
		if p.starveTick > 3 {
			p.starveTick = 0
			p.HurtTimer = 0
			p.TakeDamage(g, 1)
		}
	}

	p.HurtTimer = math.Max(0, p.HurtTimer-dt)
	p.SwingTimer = math.Max(0, p.SwingTimer-dt)
	p.EatTimer = math.Max(0, p.EatTimer-dt)

	// ---- Hotbar selection ----
	for k := ebiten.Key1; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			p.Hotbar = int(k - ebiten.Key1)
		}
	}
	_, wheelY := ebiten.Wheel()
	if wheelY < 0 {
		p.Hotbar = (p.Hotbar + 1) % 9
	} else if wheelY > 0 {
		p.Hotbar = (p.Hotbar + 8) % 9
	}
	// Drop held item with Q.
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) && !p.Held().Empty() {
		d := NewItemDrop(p.X+p.W/2+p.Facing, p.Y+0.5, ItemStack{p.Held().Item, 1})
		d.VX = p.Facing * 7
		w.Drops = append(w.Drops, d)
		p.Held().Count--
		if p.Held().Count == 0 {
			p.Held().Item = INone
		}
	}

	if g.UIOpen() {
		p.MineProgress = 0
		return
	}

	// ---- Cursor world position ----
	mx, my := g.MouseWorld()
	bx, by := int(math.Floor(mx)), int(math.Floor(my))
	reach := 5.5
	inReach := math.Hypot(mx-(p.X+p.W/2), my-(p.Y+p.H/2)) <= reach
	g.CursorX, g.CursorY = bx, by
	g.CursorReach = inReach

	// ---- Attacking mobs ----
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && inReach {
		p.SwingTimer = 0.25
		for _, m := range w.Mobs {
			if mx >= m.X && mx <= m.X+m.W && my >= m.Y && my <= m.Y+m.H {
				dmg := 1.0
				if d := p.Held().Item.Info().Damage; d > 0 {
					dmg = d
				}
				kb := 5.0
				if m.X+m.W/2 < p.X {
					kb = -kb
				}
				m.Hurt(g, dmg, kb)
				p.Hunger = math.Max(0, p.Hunger-0.1)
				goto afterMine // don't also mine through the mob
			}
		}
	}

	// ---- Mining ----
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && inReach {
		b := w.Block(bx, by)
		bi := b.Info()
		if b != BAir && bi.Hardness >= 0 {
			if bx != p.MineX || by != p.MineY {
				p.MineX, p.MineY = bx, by
				p.MineProgress = 0
			}
			tier, matches := p.toolFor(b)
			mult := 1.0
			if matches {
				mult = 1 + float64(tier)*1.6
				if tier == TierGold {
					mult = 7 // gold digs fast
				}
			}
			p.SwingTimer = 0.2
			p.MineProgress += dt * mult
			if p.MineProgress >= bi.Hardness {
				p.MineProgress = 0
				g.BreakBlock(bx, by, tier, matches)
			}
		} else {
			p.MineProgress = 0
		}
	} else {
		p.MineProgress = 0
	}
afterMine:

	// ---- Right click: use / place / eat ----
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.UseAction(bx, by, inReach)
	}
}

// UseAction handles right-click: interact with blocks, place, eat.
func (g *Game) UseAction(bx, by int, inReach bool) {
	p := &g.Player
	w := &g.World
	b := w.Block(bx, by)

	if inReach {
		switch b {
		case BCraftTable:
			g.OpenCrafting(true)
			return
		case BFurnace:
			g.OpenFurnace()
			return
		case BChest:
			g.OpenChest(bx, by)
			return
		case BTNT:
			if p.Held().Item == IFlintSteel {
				w.SetBlock(bx, by, BAir)
				g.Explode(float64(bx)+0.5, float64(by)+0.5, 3.5, 10)
				return
			}
		}
	}

	held := p.Held()
	if held.Empty() {
		return
	}
	inf := held.Item.Info()

	// Eat food.
	if inf.Food > 0 && p.Hunger < 20 && p.EatTimer <= 0 {
		p.Hunger = math.Min(20, p.Hunger+float64(inf.Food))
		p.HP = math.Min(p.MaxHP, p.HP+inf.Heal)
		p.EatTimer = 0.4
		held.Count--
		if held.Count == 0 {
			held.Item = INone
		}
		if held.Item == IMushroomStew {
			held.Item = INone
		}
		return
	}
	// Equip armor.
	if inf.Armor > 0 {
		old := p.Armor
		p.Armor = ItemStack{held.Item, 1}
		*held = old
		return
	}
	// Buckets.
	if inReach {
		switch held.Item {
		case IBucket:
			switch b {
			case BWater:
				w.SetBlock(bx, by, BAir)
				held.Item = IWaterBucket
				return
			case BLava:
				w.SetBlock(bx, by, BAir)
				held.Item = ILavaBucket
				return
			}
		case IWaterBucket:
			if b == BAir || b.Info().Replace {
				w.SetBlock(bx, by, BWater)
				held.Item = IBucket
				return
			}
		case ILavaBucket:
			if b == BAir || b.Info().Replace {
				w.SetBlock(bx, by, BLava)
				held.Item = IBucket
				return
			}
		}
	}
	// Bow.
	if held.Item == IBow {
		if p.Consume(IArrow, 1) {
			mx, my := g.MouseWorld()
			ox, oy := p.X+p.W/2, p.Y+0.5
			d := math.Hypot(mx-ox, my-oy)
			if d > 0.1 {
				g.SpawnArrow(ox, oy, (mx-ox)/d*24, (my-oy)/d*24, 6, true)
			}
		}
		return
	}

	// Place a block.
	if !held.Item.IsBlock() || !inReach {
		return
	}
	place := Block(held.Item)
	if !b.Info().Replace {
		return
	}
	// Cannot place inside self or mobs (unless non-solid).
	if place.Solid() {
		if rectsOverlap(float64(bx), float64(by), 1, 1, p.X, p.Y, p.W, p.H) {
			return
		}
		for _, m := range w.Mobs {
			if rectsOverlap(float64(bx), float64(by), 1, 1, m.X, m.Y, m.W, m.H) {
				return
			}
		}
	}
	// Needs an adjacent block to attach to.
	adjacent := false
	for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nb := w.Block(bx+d[0], by+d[1])
		if nb != BAir && !nb.Liquid() {
			adjacent = true
			break
		}
	}
	if !adjacent {
		return
	}
	w.SetBlock(bx, by, place)
	if place == BChest {
		w.Chests[[2]int{bx, by}] = make([]ItemStack, 0)
	}
	held.Count--
	if held.Count == 0 {
		held.Item = INone
	}
}

// BreakBlock removes a block, spawning drops and particles.
func (g *Game) BreakBlock(bx, by int, tier int, toolMatches bool) {
	w := &g.World
	b := w.Block(bx, by)
	if b == BAir || b.Info().Hardness < 0 {
		return
	}
	bi := b.Info()
	w.SpawnBreakParticles(bx, by, b)

	// Chests spill their contents.
	if b == BChest {
		key := [2]int{bx, by}
		for _, s := range w.Chests[key] {
			if !s.Empty() {
				w.Drops = append(w.Drops, NewItemDrop(float64(bx)+0.5, float64(by)+0.5, s))
			}
		}
		delete(w.Chests, key)
	}

	w.SetBlock(bx, by, BAir)

	// Attached blocks above pop off (torches, plants...).
	above := w.Block(bx, by-1)
	if above != BAir && !above.Solid() && !above.Liquid() && above.Info().Hardness >= 0 && above.Info().Hardness < 0.5 {
		g.BreakBlock(bx, by-1, TierNone, false)
	}
	// Sand and gravel fall.
	g.settleFalling(bx, by-1)

	// Drops.
	if bi.NeedsTool && (!toolMatches || tier < bi.MinTier) {
		return
	}
	g.dropBlockItem(bx, by, b)
	g.Player.XP += xpForBlock(b)
}

func xpForBlock(b Block) int {
	switch b {
	case BCoalOre, BDeepCoalOre:
		return 1
	case BRedstoneOre, BDeepRedstoneOre, BLapisOre, BDeepLapisOre:
		return 2
	case BDiamondOre, BDeepDiamondOre, BEmeraldOre, BDeepEmeraldOre:
		return 6
	case BSculk, BSculkSensor:
		return 2
	}
	return 0
}

// dropBlockItem spawns the drop entity for a broken block.
func (g *Game) dropBlockItem(bx, by int, b Block) {
	bi := b.Info()
	var stack ItemStack
	if bi.Drops == INone && bi.DropMin == 0 && bi.DropMax == 0 {
		stack = ItemStack{Item(b), 1} // default: drop itself
	} else if bi.Drops == INone {
		return
	} else {
		n := bi.DropMin
		if bi.DropMax > bi.DropMin {
			n += g.rng.Intn(bi.DropMax - bi.DropMin + 1)
		}
		if n <= 0 {
			return
		}
		stack = ItemStack{bi.Drops, n}
	}
	g.World.Drops = append(g.World.Drops, NewItemDrop(float64(bx)+0.5, float64(by)+0.5, stack))
}

// settleFalling makes sand/gravel columns fall when unsupported.
func (g *Game) settleFalling(bx, by int) {
	w := &g.World
	for y := by; y >= 0; y-- {
		b := w.Block(bx, y)
		if b != BSand && b != BGravel {
			break
		}
		// Fall down as far as possible.
		ty := y + 1
		for ty < WorldH && (w.Block(bx, ty) == BAir || w.Block(bx, ty).Liquid()) {
			ty++
		}
		ty--
		if ty > y {
			w.SetBlock(bx, y, BAir)
			w.SetBlock(bx, ty, b)
		} else {
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
