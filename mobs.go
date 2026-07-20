package main

import (
	"math"
	"math/rand"
)

// MobKind enumerates mob types.
type MobKind int

const (
	MobZombie MobKind = iota
	MobSkeleton
	MobCreeper
	MobSpider
	MobSlime
	MobBat
	MobWarden // deep dark guardian
	MobPig
	MobCow
	MobSheep
	MobChicken
	MobKindCount
)

type MobInfo struct {
	Name    string
	MaxHP   float64
	Damage  float64
	Speed   float64
	W, H    float64
	Hostile bool
	Flying  bool
	Drops   []lootEntry
	XP      int
}

var mobInfo = [MobKindCount]MobInfo{
	MobZombie:   {Name: "Zombie", MaxHP: 20, Damage: 3, Speed: 2.2, W: 0.7, H: 1.8, Hostile: true, XP: 5, Drops: []lootEntry{{IRottenFlesh, 1, 2, 0.9}, {IIronIngot, 1, 1, 0.05}}},
	MobSkeleton: {Name: "Skeleton", MaxHP: 20, Damage: 3, Speed: 2.4, W: 0.7, H: 1.8, Hostile: true, XP: 5, Drops: []lootEntry{{IBone, 1, 2, 0.9}, {IArrow, 1, 3, 0.6}}},
	MobCreeper:  {Name: "Creeper", MaxHP: 20, Damage: 12, Speed: 2.0, W: 0.7, H: 1.6, Hostile: true, XP: 5, Drops: []lootEntry{{IGunpowder, 1, 2, 0.9}}},
	MobSpider:   {Name: "Spider", MaxHP: 16, Damage: 2, Speed: 3.0, W: 1.2, H: 0.8, Hostile: true, XP: 5, Drops: []lootEntry{{IString, 1, 2, 0.9}, {ISpiderEye, 1, 1, 0.3}}},
	MobSlime:    {Name: "Slime", MaxHP: 12, Damage: 2, Speed: 1.6, W: 1.0, H: 1.0, Hostile: true, XP: 3, Drops: []lootEntry{{ISlimeball, 1, 2, 0.9}}},
	MobBat:      {Name: "Bat", MaxHP: 6, Damage: 0, Speed: 3.5, W: 0.5, H: 0.5, Flying: true, XP: 1},
	MobWarden:   {Name: "Warden", MaxHP: 80, Damage: 8, Speed: 2.8, W: 0.9, H: 2.4, Hostile: true, XP: 40, Drops: []lootEntry{{IXPGem, 3, 6, 1.0}, {IDiamond, 1, 2, 0.5}}},
	MobPig:      {Name: "Pig", MaxHP: 10, Speed: 1.6, W: 0.9, H: 0.9, XP: 2, Drops: []lootEntry{{IPorkchop, 1, 3, 1.0}}},
	MobCow:      {Name: "Cow", MaxHP: 10, Speed: 1.5, W: 0.9, H: 1.3, XP: 2, Drops: []lootEntry{{IBeef, 1, 3, 1.0}, {ILeather, 1, 2, 0.8}}},
	MobSheep:    {Name: "Sheep", MaxHP: 8, Speed: 1.5, W: 0.9, H: 1.2, XP: 2, Drops: []lootEntry{{IWoolItem, 1, 2, 1.0}}},
	MobChicken:  {Name: "Chicken", MaxHP: 4, Speed: 1.4, W: 0.6, H: 0.7, XP: 1, Drops: []lootEntry{{IChicken, 1, 1, 1.0}, {IFeather, 1, 2, 0.7}}},
}

// Mob is a live creature.
type Mob struct {
	Entity
	Kind      MobKind
	HP        float64
	Dir       float64 // wander direction
	DirTimer  float64
	AttackCD  float64
	HurtTimer float64
	Fuse      float64 // creeper fuse
	ShootCD   float64 // skeleton
}

func NewMob(k MobKind, x, y float64) *Mob {
	inf := &mobInfo[k]
	return &Mob{
		Entity: Entity{X: x - inf.W/2, Y: y - inf.H, W: inf.W, H: inf.H, StepUp: !inf.Flying},
		Kind:   k,
		HP:     inf.MaxHP,
		Dir:    1,
	}
}

func (m *Mob) Info() *MobInfo { return &mobInfo[m.Kind] }

// Update runs one tick of mob AI and physics.
func (m *Mob) Update(g *Game, dt float64) {
	inf := m.Info()
	p := &g.Player
	px := p.X + p.W/2
	py := p.Y + p.H/2
	mx := m.X + m.W/2
	my := m.Y + m.H/2
	dx := px - mx
	dy := py - my
	dist := math.Hypot(dx, dy)

	m.DirTimer -= dt
	m.AttackCD -= dt
	m.HurtTimer -= dt
	m.ShootCD -= dt

	chase := inf.Hostile && dist < 24 && !p.Dead
	if inf.Flying {
		// Bats flit around.
		if m.DirTimer <= 0 {
			m.DirTimer = 0.5 + g.rng.Float64()
			m.VX = (g.rng.Float64()*2 - 1) * inf.Speed
			m.VY = (g.rng.Float64()*2 - 1) * inf.Speed
		}
		m.X += m.VX * dt
		m.Y += m.VY * dt
		if m.collides(&g.World) || m.Y < 0 {
			m.X -= m.VX * dt
			m.Y -= m.VY * dt
			m.DirTimer = 0
		}
		return
	}

	w := &g.World
	if chase {
		// Steer toward the player with a dead zone so the mob doesn't
		// jitter back and forth when nearly aligned.
		switch {
		case dx > 0.5:
			m.Dir = 1
		case dx < -0.5:
			m.Dir = -1
		default:
			m.Dir = 0
		}
		// Skeletons kite: back away when the player gets too close so
		// they keep room to shoot.
		if m.Kind == MobSkeleton && dist < 5 {
			m.Dir = -m.Dir
		}
	} else if m.DirTimer <= 0 {
		m.DirTimer = 2 + g.rng.Float64()*3
		r := g.rng.Float64()
		switch {
		case r < 0.45:
			m.Dir = 0
		case r < 0.72:
			m.Dir = 1
		default:
			m.Dir = -1
		}
	}

	// Ledge avoidance: idle/wandering mobs turn back at cliffs instead of
	// marching off. Chasers commit so they can still pursue down drops.
	if !chase && m.OnGround && m.Dir != 0 {
		aheadX := mx + m.Dir*(m.W/2+0.25)
		footY := m.Y + m.H
		if !solidAt(w, aheadX, footY+0.3) && !solidAt(w, aheadX, footY+1.3) {
			m.Dir = -m.Dir
			m.DirTimer = 1 + g.rng.Float64()
		}
	}

	speed := inf.Speed
	if !chase {
		speed *= 0.45
	}
	// Ease toward the target speed rather than snapping, so motion and
	// knockback recovery look smooth.
	target := m.Dir * speed
	m.VX += (target - m.VX) * math.Min(1, dt*12)

	// Jumping. Auto step-up already clears single-block ledges, so only
	// jump for taller walls, to climb toward a player above, or to bob
	// out of water.
	if (m.OnGround || m.InWater) && m.Dir != 0 {
		frontX := mx + m.Dir*(m.W/2+0.25)
		lowBlocked := solidAt(w, frontX, m.Y+m.H-0.3)
		highBlocked := solidAt(w, frontX, m.Y+m.H-1.4)
		needClimb := chase && dy < -1.3 && solidAt(w, mx, m.Y+m.H+0.4)
		if (lowBlocked && highBlocked) || needClimb {
			if !solidAt(w, mx, m.Y-0.6) { // headroom before committing
				m.VY = -11.5
			}
		}
	}
	if m.InWater {
		// Swim up toward the surface (or the player if above).
		if chase && dy < 0 {
			m.VY = -6
		} else if m.VY > 1 {
			m.VY = 1
		}
	}
	// Slimes hop along the ground.
	if m.Kind == MobSlime && m.OnGround && m.Dir != 0 {
		m.VY = -8
	}
	m.MoveAndCollide(w, dt)
	if m.InLava {
		m.Hurt(g, 4*dt, 0)
	}

	// Creeper: fuse when close.
	if m.Kind == MobCreeper {
		if dist < 2.2 && !p.Dead {
			m.Fuse += dt
			if m.Fuse > 1.2 {
				g.Explode(mx, my, 3.2, 10)
				m.HP = 0
				return
			}
		} else {
			m.Fuse = math.Max(0, m.Fuse-dt*2)
		}
		return
	}

	// Skeleton: shoot arrows from range.
	if m.Kind == MobSkeleton && chase && dist > 3 && dist < 14 && m.ShootCD <= 0 {
		m.ShootCD = 1.6
		g.SpawnArrow(mx, my, dx/dist*18, dy/dist*18-1.5, inf.Damage, false)
	}

	// Melee contact damage.
	if inf.Hostile && inf.Damage > 0 && m.AttackCD <= 0 && !p.Dead {
		if rectsOverlap(m.X, m.Y, m.W, m.H, p.X, p.Y, p.W, p.H) {
			p.TakeDamage(g, inf.Damage)
			m.AttackCD = 1.0
			// Knock the player back.
			kb := 6.0
			if dx < 0 {
				kb = -kb
			}
			p.VX += kb
			p.VY -= 4
		}
	}
}

// Hurt applies damage and knockback; drops loot on death.
func (m *Mob) Hurt(g *Game, dmg, knockX float64) {
	if m.HurtTimer > 0 {
		return
	}
	m.HP -= dmg
	m.HurtTimer = 0.4
	m.VX += knockX
	m.VY -= 5
	if m.HP <= 0 {
		inf := m.Info()
		rng := rand.New(rand.NewSource(g.rng.Int63()))
		for _, e := range inf.Drops {
			if rng.Float64() < e.chance {
				n := e.min
				if e.max > e.min {
					n += rng.Intn(e.max - e.min + 1)
				}
				g.World.Drops = append(g.World.Drops,
					NewItemDrop(m.X+m.W/2, m.Y+m.H/2, ItemStack{e.item, n}))
			}
		}
		g.Player.XP += inf.XP
	}
}

func rectsOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && bx < ax+aw && ay < by+bh && by < ay+ah
}

// trySpawnMobs occasionally spawns mobs around the player. Hostiles need
// darkness (night surface or unlit caves); passives need daylight grass.
func (g *Game) trySpawnMobs() {
	w := &g.World
	hostiles, passives := 0, 0
	for _, m := range w.Mobs {
		if m.Info().Hostile {
			hostiles++
		} else {
			passives++
		}
	}
	px := int(g.Player.X)
	py := int(g.Player.Y)

	for try := 0; try < 8; try++ {
		x := px + g.rng.Intn(60) - 30
		y := py + g.rng.Intn(40) - 20
		if y < 1 || y >= WorldH-2 {
			continue
		}
		dx, dy := float64(x-px), float64(y-py)
		if dx*dx+dy*dy < 12*12 { // not too close
			continue
		}
		if w.Block(x, y).Solid() || w.Block(x, y-1).Solid() || !w.Block(x, y+1).Solid() {
			continue
		}
		light := g.lightLevelAt(x, y)
		surface := y <= w.SurfaceY(x)

		if light <= 4 && hostiles < 24 {
			kind := MobZombie
			r := g.rng.Float64()
			depth := y - w.SurfaceY(x)
			switch {
			case y > DeepDarkY && w.nearSculk(x, y) && r < 0.25:
				kind = MobWarden
			case r < 0.3:
				kind = MobZombie
			case r < 0.5:
				kind = MobSkeleton
			case r < 0.65:
				kind = MobSpider
			case r < 0.8:
				kind = MobCreeper
			case depth > 40 && r < 0.92:
				kind = MobSlime
			default:
				kind = MobBat
			}
			w.Mobs = append(w.Mobs, NewMob(kind, float64(x)+0.5, float64(y)+1))
			hostiles++
		} else if surface && light > 8 && !g.World.IsNight() && passives < 8 {
			if w.Block(x, y+1) != BGrass {
				continue
			}
			kinds := []MobKind{MobPig, MobCow, MobSheep, MobChicken}
			kind := kinds[g.rng.Intn(len(kinds))]
			w.Mobs = append(w.Mobs, NewMob(kind, float64(x)+0.5, float64(y)+1))
			passives++
		}
	}
}

func (w *World) nearSculk(x, y int) bool {
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			if b := w.Block(x+dx, y+dy); b == BSculk || b == BSculkSensor {
				return true
			}
		}
	}
	return false
}

// Arrow projectile (skeleton and player bow).
type Arrow struct {
	X, Y, VX, VY float64
	Damage       float64
	FromPlayer   bool
	Life         float64
}

func (g *Game) SpawnArrow(x, y, vx, vy, dmg float64, fromPlayer bool) {
	g.Arrows = append(g.Arrows, &Arrow{X: x, Y: y, VX: vx, VY: vy, Damage: dmg, FromPlayer: fromPlayer, Life: 4})
}

// Explode destroys terrain in a radius and damages entities. Used by
// creepers and TNT.
func (g *Game) Explode(cx, cy, radius, dmg float64) {
	w := &g.World
	r := int(radius) + 1
	for bx := int(cx) - r; bx <= int(cx)+r; bx++ {
		for by := int(cy) - r; by <= int(cy)+r; by++ {
			dx, dy := float64(bx)+0.5-cx, float64(by)+0.5-cy
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			b := w.Block(bx, by)
			if b == BAir || b == BBedrock || b == BObsidian || b.Liquid() {
				continue
			}
			if b == BTNT {
				w.SetBlock(bx, by, BAir)
				g.Explode(float64(bx)+0.5, float64(by)+0.5, 3.5, 10)
				continue
			}
			w.SpawnBreakParticles(bx, by, b)
			// Chance to drop the block.
			if g.rng.Float64() < 0.3 {
				g.dropBlockItem(bx, by, b)
			}
			w.SetBlock(bx, by, BAir)
		}
	}
	// Damage player and mobs by proximity.
	p := &g.Player
	pd := math.Hypot(p.X+p.W/2-cx, p.Y+p.H/2-cy)
	if pd < radius*1.5 {
		p.TakeDamage(g, dmg*(1-pd/(radius*1.5)))
	}
	for _, m := range w.Mobs {
		md := math.Hypot(m.X+m.W/2-cx, m.Y+m.H/2-cy)
		if md < radius*1.5 {
			m.Hurt(g, dmg*(1-md/(radius*1.5)), 0)
		}
	}
}
