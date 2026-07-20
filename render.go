package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// worldToScreen converts block coordinates to pixels.
func (g *Game) worldToScreen(wx, wy float64) (float64, float64) {
	return (wx-g.CamX)*TileSize + ScreenW/2, (wy-g.CamY)*TileSize + ScreenH/2
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawSky(screen)
	g.drawWorld(screen)
	g.drawEntities(screen)
	g.drawPlayer(screen)
	g.drawCursor(screen)
	g.drawHUD(screen)
	if g.UIOpen() {
		g.drawUI(screen)
	}
	if g.Paused {
		vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 140}, false)
		ebitenutil.DebugPrintAt(screen, "PAUSED - press ESC to resume", ScreenW/2-90, ScreenH/2)
	}
	if g.Player.Dead {
		vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{120, 0, 0, 120}, false)
		ebitenutil.DebugPrintAt(screen, "YOU DIED - respawning...", ScreenW/2-80, ScreenH/2)
	}
}

func (g *Game) drawSky(screen *ebiten.Image) {
	day := g.World.DayFactor()
	// Sky fades with depth underground too.
	depth := g.CamY - float64(SurfaceBaseY)
	depthFade := 1 - math.Min(math.Max(depth/80, 0), 1)

	top := lerpColor(color.RGBA{8, 8, 24, 255}, color.RGBA{92, 148, 236, 255}, day)
	bottom := lerpColor(color.RGBA{16, 16, 40, 255}, color.RGBA{150, 195, 245, 255}, day)
	top = scaleColor(top, 0.15+0.85*depthFade)
	bottom = scaleColor(bottom, 0.15+0.85*depthFade)
	for i := 0; i < 8; i++ {
		c := lerpColor(top, bottom, float64(i)/7)
		vector.DrawFilledRect(screen, 0, float32(i*ScreenH/8), ScreenW, float32(ScreenH/8+1), c, false)
	}
	if depthFade > 0.05 {
		// Sun and moon travel across the sky.
		t := g.World.Time / DayLength
		ang := (t - 0.25) * 2 * math.Pi
		sx := ScreenW/2 + math.Cos(ang)*ScreenW*0.42
		sy := ScreenH*0.55 + math.Sin(ang)*ScreenH*0.5
		vector.DrawFilledRect(screen, float32(sx-14), float32(sy-14), 28, 28, color.RGBA{255, 235, 120, uint8(255 * depthFade)}, false)
		mx := ScreenW/2 + math.Cos(ang+math.Pi)*ScreenW*0.42
		my := ScreenH*0.55 + math.Sin(ang+math.Pi)*ScreenH*0.5
		vector.DrawFilledRect(screen, float32(mx-11), float32(my-11), 22, 22, color.RGBA{220, 220, 235, uint8(255 * depthFade)}, false)
		// Stars at night.
		if day < 0.5 {
			a := uint8((0.5 - day) * 2 * 180 * depthFade)
			for i := 0; i < 60; i++ {
				x := float32(hash2(1234, i, 0) % ScreenW)
				y := float32(hash2(4321, i, 1) % (ScreenH * 2 / 3))
				vector.DrawFilledRect(screen, x, y, 2, 2, color.RGBA{255, 255, 255, a}, false)
			}
		}
	}
}

func (g *Game) drawWorld(screen *ebiten.Image) {
	w := &g.World
	x0 := int(math.Floor(g.CamX - ScreenW/TileSize/2 - 1))
	x1 := int(math.Ceil(g.CamX + ScreenW/TileSize/2 + 1))
	y0 := int(math.Floor(g.CamY - ScreenH/TileSize/2 - 1))
	y1 := int(math.Ceil(g.CamY + ScreenH/TileSize/2 + 1))
	if y0 < 0 {
		y0 = 0
	}
	if y1 >= WorldH {
		y1 = WorldH - 1
	}

	day := w.DayFactor()
	for x := x0; x <= x1; x++ {
		c := w.ChunkAt(floorDiv(x, ChunkW))
		lx := mod(x, ChunkW)
		surface := c.Height[lx]
		for y := y0; y <= y1; y++ {
			b := c.At(lx, y)
			sx, sy := g.worldToScreen(float64(x), float64(y))

			sky := float64(c.SkyLight[y*ChunkW+lx]) * (0.25 + 0.75*day)
			blk := float64(c.BlockLight[y*ChunkW+lx])
			light := sky
			if blk > light {
				light = blk
			}
			br := 0.10 + 0.90*light/15

			// Cave background wall behind transparent blocks underground.
			if !b.Opaque() && y > surface {
				wall := BStone
				if y >= DeepslateY {
					wall = BDeepslate
				}
				if y <= w.SurfaceY(x)+4 {
					wall = BDirt
				}
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(sx, sy)
				f := float32(br * 0.45)
				op.ColorScale.Scale(f, f, f, 1)
				screen.DrawImage(blockTex[wall], op)
			}

			if b == BAir {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(sx, sy)
			f := float32(br)
			if b.Liquid() {
				op.ColorScale.Scale(f, f, f, 0.82)
			} else {
				op.ColorScale.Scale(f, f, f, 1)
			}
			screen.DrawImage(blockTex[b], op)
		}
	}

	// Mining crack overlay.
	p := &g.Player
	if p.MineProgress > 0 {
		b := w.Block(p.MineX, p.MineY)
		if b != BAir && b.Info().Hardness > 0 {
			frac := p.MineProgress / b.Info().Hardness
			sx, sy := g.worldToScreen(float64(p.MineX), float64(p.MineY))
			n := int(frac * 8)
			for i := 0; i < n; i++ {
				cx := float32(sx) + float32(hash2(9, i, p.MineX)%TileSize)
				cy := float32(sy) + float32(hash2(7, i, p.MineY)%TileSize)
				vector.StrokeLine(screen, cx-3, cy-3, cx+3, cy+3, 1, color.RGBA{0, 0, 0, 200}, false)
				vector.StrokeLine(screen, cx+3, cy-3, cx-3, cy+3, 1, color.RGBA{0, 0, 0, 140}, false)
			}
		}
	}
}

func (g *Game) drawEntities(screen *ebiten.Image) {
	w := &g.World
	// Item drops.
	for _, d := range w.Drops {
		sx, sy := g.worldToScreen(d.X, d.Y)
		bob := math.Sin(d.Age*4) * 2
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(0.6, 0.6)
		op.GeoM.Translate(sx, sy+bob)
		l := float32(0.3 + 0.7*float64(g.lightLevelAt(int(d.X), int(d.Y)))/15)
		op.ColorScale.Scale(l, l, l, 1)
		screen.DrawImage(texForItem(d.Stack.Item), op)
		if d.Stack.Count > 1 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprint(d.Stack.Count), int(sx)+4, int(sy)+2)
		}
	}
	// Arrows.
	for _, a := range g.Arrows {
		sx, sy := g.worldToScreen(a.X, a.Y)
		ang := math.Atan2(a.VY, a.VX)
		ex := sx + math.Cos(ang)*8
		ey := sy + math.Sin(ang)*8
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(ex), float32(ey), 2, color.RGBA{200, 190, 170, 255}, false)
	}
	// Mobs.
	for _, m := range w.Mobs {
		g.drawMob(screen, m)
	}
	// Particles.
	for _, p := range w.Particles {
		sx, sy := g.worldToScreen(p.X, p.Y)
		vector.DrawFilledRect(screen, float32(sx), float32(sy), 3, 3, color.RGBA{p.R, p.G, p.B, 255}, false)
	}
}

var mobColors = [MobKindCount][2]color.RGBA{
	MobZombie:   {{60, 140, 70, 255}, {40, 90, 140, 255}},
	MobSkeleton: {{215, 215, 205, 255}, {160, 160, 150, 255}},
	MobCreeper:  {{80, 180, 70, 255}, {60, 140, 55, 255}},
	MobSpider:   {{40, 35, 35, 255}, {60, 50, 50, 255}},
	MobSlime:    {{90, 200, 80, 255}, {120, 220, 110, 255}},
	MobBat:      {{70, 60, 55, 255}, {90, 80, 70, 255}},
	MobWarden:   {{20, 50, 65, 255}, {40, 190, 210, 255}},
	MobPig:      {{240, 170, 170, 255}, {220, 140, 140, 255}},
	MobCow:      {{110, 75, 55, 255}, {235, 235, 235, 255}},
	MobSheep:    {{230, 230, 230, 255}, {200, 170, 150, 255}},
	MobChicken:  {{240, 240, 235, 255}, {220, 60, 50, 255}},
}

func (g *Game) drawMob(screen *ebiten.Image, m *Mob) {
	sx, sy := g.worldToScreen(m.X, m.Y)
	wpx := float32(m.W * TileSize)
	hpx := float32(m.H * TileSize)
	l := float32(0.25 + 0.75*float64(g.lightLevelAt(int(m.X+m.W/2), int(m.Y+m.H/2)))/15)
	c := mobColors[m.Kind]
	body := c[0]
	if m.HurtTimer > 0.2 {
		body = color.RGBA{230, 80, 80, 255}
	}
	if m.Kind == MobCreeper && m.Fuse > 0 && int(m.Fuse*10)%2 == 0 {
		body = color.RGBA{255, 255, 255, 255}
	}
	sc := func(cc color.RGBA) color.RGBA {
		return color.RGBA{uint8(float32(cc.R) * l), uint8(float32(cc.G) * l), uint8(float32(cc.B) * l), cc.A}
	}
	// Body block + head/accent.
	vector.DrawFilledRect(screen, float32(sx), float32(sy)+hpx*0.3, wpx, hpx*0.7, sc(c[1]), false)
	vector.DrawFilledRect(screen, float32(sx), float32(sy), wpx, hpx*0.35, sc(body), false)
	// Eyes.
	eye := color.RGBA{0, 0, 0, 255}
	if m.Kind == MobSpider || m.Kind == MobWarden {
		eye = color.RGBA{200, 40, 40, 255}
	}
	if m.Kind == MobWarden {
		eye = color.RGBA{60, 220, 235, 255}
	}
	ex := float32(sx) + wpx*0.2
	if g.Player.X < m.X {
		ex = float32(sx) + wpx*0.15
	}
	vector.DrawFilledRect(screen, ex, float32(sy)+hpx*0.12, 3, 3, eye, false)
	vector.DrawFilledRect(screen, ex+wpx*0.4, float32(sy)+hpx*0.12, 3, 3, eye, false)
	// HP bar when damaged.
	if m.HP < m.Info().MaxHP {
		frac := float32(m.HP / m.Info().MaxHP)
		vector.DrawFilledRect(screen, float32(sx), float32(sy)-6, wpx, 3, color.RGBA{40, 40, 40, 200}, false)
		vector.DrawFilledRect(screen, float32(sx), float32(sy)-6, wpx*frac, 3, color.RGBA{220, 50, 50, 230}, false)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := &g.Player
	if p.Dead {
		return
	}
	sx, sy := g.worldToScreen(p.X, p.Y)
	wpx := float32(p.W * TileSize)
	hpx := float32(p.H * TileSize)
	l := float32(0.35 + 0.65*float64(g.lightLevelAt(int(p.X+p.W/2), int(p.Y+p.H/2)))/15)
	sc := func(cc color.RGBA) color.RGBA {
		if p.HurtTimer > 0.3 {
			return color.RGBA{230, 90, 90, cc.A}
		}
		return color.RGBA{uint8(float32(cc.R) * l), uint8(float32(cc.G) * l), uint8(float32(cc.B) * l), cc.A}
	}
	// Legs, torso, head.
	vector.DrawFilledRect(screen, float32(sx), float32(sy)+hpx*0.62, wpx, hpx*0.38, sc(color.RGBA{55, 70, 140, 255}), false)
	vector.DrawFilledRect(screen, float32(sx), float32(sy)+hpx*0.28, wpx, hpx*0.36, sc(color.RGBA{60, 160, 160, 255}), false)
	vector.DrawFilledRect(screen, float32(sx)+wpx*0.08, float32(sy), wpx*0.84, hpx*0.28, sc(color.RGBA{225, 180, 140, 255}), false)
	// Eye in facing direction.
	ex := float32(sx) + wpx*0.62
	if p.Facing < 0 {
		ex = float32(sx) + wpx*0.22
	}
	vector.DrawFilledRect(screen, ex, float32(sy)+hpx*0.10, 3, 3, color.RGBA{40, 40, 60, 255}, false)
	// Held item in hand.
	if !p.Held().Empty() {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(0.7, 0.7)
		swing := p.SwingTimer * 40
		hx := sx + float64(wpx)*0.7
		if p.Facing < 0 {
			hx = sx - 8
		}
		op.GeoM.Translate(hx, sy+float64(hpx)*0.35-swing)
		screen.DrawImage(texForItem(p.Held().Item), op)
	}
}

func (g *Game) drawCursor(screen *ebiten.Image) {
	if g.UIOpen() {
		return
	}
	sx, sy := g.worldToScreen(float64(g.CursorX), float64(g.CursorY))
	c := color.RGBA{255, 255, 255, 180}
	if !g.CursorReach {
		c = color.RGBA{255, 255, 255, 50}
	}
	vector.StrokeRect(screen, float32(sx), float32(sy), TileSize, TileSize, 2, c, false)
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		255,
	}
}

func scaleColor(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * f), uint8(float64(c.G) * f), uint8(float64(c.B) * f), 255}
}
