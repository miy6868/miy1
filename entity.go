package main

import "math"

// Entity coordinates are in blocks (floats); W/H are the hitbox size in blocks.
type Entity struct {
	X, Y   float64 // top-left corner of hitbox
	W, H   float64
	VX, VY float64
	OnGround bool
	InWater  bool
	InLava   bool
}

const gravity = 32.0 // blocks/s^2

// solidAt reports whether a solid block occupies (x, y).
func solidAt(w *World, x, y float64) bool {
	return w.Block(int(math.Floor(x)), int(math.Floor(y))).Solid()
}

// collides checks the entity box against solid blocks.
func (e *Entity) collides(w *World) bool {
	x0 := int(math.Floor(e.X))
	y0 := int(math.Floor(e.Y))
	x1 := int(math.Floor(e.X + e.W - 1e-6))
	y1 := int(math.Floor(e.Y + e.H - 1e-6))
	for bx := x0; bx <= x1; bx++ {
		for by := y0; by <= y1; by++ {
			if w.Block(bx, by).Solid() {
				return true
			}
		}
	}
	return false
}

// MoveAndCollide integrates velocity with axis-separated collision.
func (e *Entity) MoveAndCollide(w *World, dt float64) {
	// Liquids state (sampled at center).
	cx, cy := e.X+e.W/2, e.Y+e.H/2
	b := w.Block(int(math.Floor(cx)), int(math.Floor(cy)))
	e.InWater = b == BWater
	e.InLava = b == BLava

	drag := 1.0
	if e.InWater {
		drag = 0.5
	}
	if e.InLava {
		drag = 0.35
	}

	// X axis.
	step := e.VX * dt * drag
	e.X += step
	if e.collides(w) {
		if step > 0 {
			e.X = math.Floor(e.X+e.W) - e.W - 1e-4
		} else {
			e.X = math.Floor(e.X) + 1 + 1e-4
		}
		e.VX = 0
	}

	// Y axis.
	e.VY += gravity * dt
	maxFall := 60.0
	if e.InWater || e.InLava {
		maxFall = 5.0
	}
	if e.VY > maxFall {
		e.VY = maxFall
	}
	step = e.VY * dt * drag
	e.OnGround = false
	e.Y += step
	if e.collides(w) {
		if step > 0 {
			e.Y = math.Floor(e.Y+e.H) - e.H - 1e-4
			e.OnGround = true
		} else {
			e.Y = math.Floor(e.Y) + 1 + 1e-4
		}
		e.VY = 0
	}
}

// ItemDrop is an item lying in the world.
type ItemDrop struct {
	Entity
	Stack  ItemStack
	Age    float64
	Pickup float64 // delay before it can be picked up
}

func NewItemDrop(x, y float64, s ItemStack) *ItemDrop {
	return &ItemDrop{
		Entity: Entity{X: x - 0.25, Y: y - 0.25, W: 0.5, H: 0.5, VX: 0, VY: -3},
		Stack:  s,
		Pickup: 0.5,
	}
}

// Particle is a short-lived visual effect (block break debris, etc.).
type Particle struct {
	X, Y, VX, VY float64
	Life         float64
	R, G, B      uint8
}

func (w *World) SpawnBreakParticles(bx, by int, b Block) {
	r, g, bl := blockAvgColor(b)
	for i := 0; i < 6; i++ {
		f := hashFloat(w.Seed, bx*7+i, by*13+i)
		w.Particles = append(w.Particles, &Particle{
			X: float64(bx) + 0.2 + 0.6*f,
			Y: float64(by) + 0.2 + 0.6*hashFloat(w.Seed, by+i, bx-i),
			VX: (f - 0.5) * 6,
			VY: -2 - 3*f,
			Life: 0.4 + 0.3*f,
			R: r, G: g, B: bl,
		})
	}
}
