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
	g.drawGlows(screen)
	g.drawCursor(screen)

	// Depth vignette: subtle on the surface, heavy underground.
	depth := g.CamY - float64(SurfaceBaseY)
	vig := 0.25 + 0.45*math.Min(math.Max(depth/60, 0), 1)
	drawVignette(screen, vig, color.RGBA{0, 0, 0, 255})
	// Hurt flash.
	if g.Player.HurtTimer > 0.25 {
		drawVignette(screen, (g.Player.HurtTimer-0.25)*2.4, color.RGBA{190, 20, 20, 255})
	}
	// Low health pulse.
	if !g.Player.Dead && g.Player.HP <= 6 {
		pulse := 0.18 + 0.10*math.Sin(g.World.Time*6)
		drawVignette(screen, pulse, color.RGBA{160, 10, 10, 255})
	}

	g.drawHUD(screen)
	if g.UIOpen() {
		g.drawUI(screen)
	}
	if g.Paused {
		vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 140}, false)
		ebitenutil.DebugPrintAt(screen, "PAUSED - press ESC to resume", ScreenW/2-90, ScreenH/2)
	}
	if g.Player.Dead {
		drawVignette(screen, 0.9, color.RGBA{120, 0, 0, 255})
		vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{80, 0, 0, 90}, false)
		ebitenutil.DebugPrintAt(screen, "YOU DIED - respawning...", ScreenW/2-80, ScreenH/2)
	}
}

// gradRect fills a rect with a vertical gradient using vertex colors.
// Colors are straight (non-premultiplied); alpha is respected.
func gradRect(dst *ebiten.Image, x, y, w, h float32, top, bottom color.RGBA) {
	cr := func(c color.RGBA) (float32, float32, float32, float32) {
		return float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A) / 255
	}
	tr, tg, tb, ta := cr(top)
	br, bg, bb, ba := cr(bottom)
	vs := []ebiten.Vertex{
		{DstX: x, DstY: y, SrcX: 1, SrcY: 1, ColorR: tr, ColorG: tg, ColorB: tb, ColorA: ta},
		{DstX: x + w, DstY: y, SrcX: 2, SrcY: 1, ColorR: tr, ColorG: tg, ColorB: tb, ColorA: ta},
		{DstX: x, DstY: y + h, SrcX: 1, SrcY: 2, ColorR: br, ColorG: bg, ColorB: bb, ColorA: ba},
		{DstX: x + w, DstY: y + h, SrcX: 2, SrcY: 2, ColorR: br, ColorG: bg, ColorB: bb, ColorA: ba},
	}
	dst.DrawTriangles(vs, []uint16{0, 1, 2, 1, 2, 3}, whiteTex, nil)
}

func (g *Game) drawSky(screen *ebiten.Image) {
	day := g.World.DayFactor()
	depth := g.CamY - float64(SurfaceBaseY)
	depthFade := 1 - math.Min(math.Max(depth/80, 0), 1)

	top := lerpColor(color.RGBA{6, 7, 22, 255}, color.RGBA{88, 146, 238, 255}, day)
	bottom := lerpColor(color.RGBA{18, 18, 44, 255}, color.RGBA{168, 208, 248, 255}, day)
	top = scaleColor(top, 0.12+0.88*depthFade)
	bottom = scaleColor(bottom, 0.12+0.88*depthFade)
	gradRect(screen, 0, 0, ScreenW, ScreenH, top, bottom)

	// Dawn/dusk horizon glow.
	t := g.World.Time / DayLength
	dawn := math.Max(0, 1-math.Abs(t-0.0)/0.06) + math.Max(0, 1-math.Abs(t-1.0)/0.06)
	dusk := math.Max(0, 1-math.Abs(t-0.5)/0.06)
	if s := (dawn + dusk) * depthFade; s > 0.01 {
		a := uint8(120 * s)
		gradRect(screen, 0, ScreenH*0.35, ScreenW, ScreenH*0.65,
			color.RGBA{0, 0, 0, 0}, color.RGBA{uint8(240 * s), uint8(120 * s), uint8(50 * s), a})
	}

	if depthFade <= 0.05 {
		return
	}
	// Sun and moon travel across the sky, with soft halos.
	ang := (t - 0.25) * 2 * math.Pi
	sx := ScreenW/2 + math.Cos(ang)*ScreenW*0.42
	sy := ScreenH*0.55 + math.Sin(ang)*ScreenH*0.5
	drawGlow(screen, sx, sy, 90, color.RGBA{255, 210, 90, 255}, 0.55*depthFade)
	vector.DrawFilledCircle(screen, float32(sx), float32(sy), 15, color.NRGBA{255, 240, 170, uint8(255 * depthFade)}, true)
	vector.DrawFilledCircle(screen, float32(sx), float32(sy), 11, color.NRGBA{255, 250, 220, uint8(255 * depthFade)}, true)

	mx := ScreenW/2 + math.Cos(ang+math.Pi)*ScreenW*0.42
	my := ScreenH*0.55 + math.Sin(ang+math.Pi)*ScreenH*0.5
	drawGlow(screen, mx, my, 55, color.RGBA{170, 190, 235, 255}, 0.35*depthFade)
	vector.DrawFilledCircle(screen, float32(mx), float32(my), 11, color.NRGBA{225, 230, 240, uint8(255 * depthFade)}, true)
	vector.DrawFilledCircle(screen, float32(mx)+4, float32(my)-3, 9, scaleColor(top, 1.1), true)

	// Twinkling stars at night.
	if day < 0.55 {
		base := (0.55 - day) / 0.55 * depthFade
		for i := 0; i < 90; i++ {
			x := float32(hash2(1234, i, 0) % ScreenW)
			y := float32(hash2(4321, i, 1) % (ScreenH * 2 / 3))
			tw := 0.6 + 0.4*math.Sin(g.World.Time*2+float64(i))
			a := uint8(200 * base * tw)
			sz := float32(1 + i%2)
			vector.DrawFilledRect(screen, x, y, sz, sz, color.NRGBA{255, 255, 255, a}, false)
		}
	}

	// Parallax hill silhouettes behind the terrain.
	g.drawHills(screen, 0.25, 30, scaleColor(lerpColor(color.RGBA{30, 34, 60, 255}, color.RGBA{116, 164, 190, 255}, day), 0.9), depthFade)
	g.drawHills(screen, 0.45, 12, scaleColor(lerpColor(color.RGBA{22, 26, 48, 255}, color.RGBA{88, 138, 128, 255}, day), 0.95), depthFade)

	// Drifting clouds.
	for i := 0; i < 7; i++ {
		ct := cloudTex[i%3]
		w := float64(ct.Bounds().Dx())
		speed := 6.0 + float64(i%3)*3
		par := 0.12 + 0.05*float64(i%3)
		x := math.Mod(float64(hash2(88, i, 0)%3000)-g.CamX*TileSize*par+g.World.Time*speed, ScreenW+w) - w
		y := 30 + float64(hash2(99, i, 1)%140)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		br := float32(0.35 + 0.65*day)
		op.ColorScale.Scale(br, br, br, float32(0.85*depthFade))
		screen.DrawImage(ct, op)
	}
}

// drawHills draws a distant terrain silhouette with parallax.
func (g *Game) drawHills(screen *ebiten.Image, parallax float64, amp float64, c color.RGBA, alpha float64) {
	if alpha <= 0.05 {
		return
	}
	nc := color.NRGBA{c.R, c.G, c.B, uint8(255 * alpha)}
	base := ScreenH*0.55 - (g.CamY-float64(SurfaceBaseY))*TileSize*parallax*0.4
	step := 8
	for px := 0; px < ScreenW; px += step {
		wx := g.CamX*parallax + float64(px)*0.08
		h := g.World.Gen.terrain.Octave1(wx*0.05+parallax*100, 3, 0.5) * amp * TileSize * 0.25
		y := float32(base - h*4)
		if y < 0 {
			y = 0
		}
		vector.DrawFilledRect(screen, float32(px), y, float32(step), float32(ScreenH)-y, nc, false)
	}
}

// glowPoint is a queued bloom light for this frame.
type glowPoint struct {
	x, y  float64
	block Block
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
	// Cell brightness grid with a one-cell border for corner interpolation.
	gw := x1 - x0 + 3
	gh := y1 - y0 + 3
	cell := make([]float32, gw*gh)
	for x := x0 - 1; x <= x1+1; x++ {
		c := w.ChunkAt(floorDiv(x, ChunkW))
		lx := mod(x, ChunkW)
		for y := y0 - 1; y <= y1+1; y++ {
			var br float64
			if y < 0 {
				br = 0.10 + 0.90*(0.25+0.75*day)
			} else if y >= WorldH {
				br = 0.10
			} else {
				sky := float64(c.SkyLight[y*ChunkW+lx]) * (0.25 + 0.75*day)
				blk := float64(c.BlockLight[y*ChunkW+lx])
				l := math.Max(sky, blk)
				br = 0.08 + 0.92*l/15
			}
			cell[(y-y0+1)*gw+(x-x0+1)] = float32(br)
		}
	}
	// Corner light = max of the 4 cells that share the corner; max avoids
	// dark seams creeping over lit faces.
	corner := func(x, y int) float32 {
		i := (y - y0 + 1) * gw
		j := x - x0 + 1
		a := cell[i-gw+j-1]
		b := cell[i-gw+j]
		c2 := cell[i+j-1]
		d := cell[i+j]
		m := a
		if b > m {
			m = b
		}
		if c2 > m {
			m = c2
		}
		if d > m {
			m = d
		}
		return m
	}

	g.glows = g.glows[:0]
	for x := x0; x <= x1; x++ {
		c := w.ChunkAt(floorDiv(x, ChunkW))
		lx := mod(x, ChunkW)
		surface := c.Height[lx]
		for y := y0; y <= y1; y++ {
			b := c.At(lx, y)
			sx, sy := g.worldToScreen(float64(x), float64(y))
			c00 := corner(x, y)
			c10 := corner(x+1, y)
			c01 := corner(x, y+1)
			c11 := corner(x+1, y+1)

			// Cave background wall behind transparent blocks underground.
			if !b.Opaque() && y > surface {
				wall := BStone
				if y >= DeepslateY {
					wall = BDeepslate
				}
				if y <= w.SurfaceY(x)+4 {
					wall = BDirt
				}
				drawTileLit(screen, blockTex[wall], sx, sy, c00*0.45, c10*0.45, c01*0.45, c11*0.45, 1)
			}

			if b == BAir {
				continue
			}
			if e := b.Info().LightEmit; e >= 8 && len(g.glows) < 220 {
				// Only glow lava surface tiles, not whole lakes.
				if b != BLava || !w.Block(x, y-1).Liquid() {
					g.glows = append(g.glows, glowPoint{float64(x) + 0.5, float64(y) + 0.5, b})
				}
			}
			tex := blockTex[b]
			alpha := float32(1)
			if b.Liquid() {
				tex = liquidTexFor(b, w.Time, x, y)
				alpha = 0.85
				if b == BLava {
					// Lava is self-lit.
					c00, c10, c01, c11 = 1, 1, 1, 1
					alpha = 1
				}
			}
			drawTileLit(screen, tex, sx, sy, c00, c10, c01, c11, alpha)
		}
	}

	// Mining crack overlay.
	p := &g.Player
	if p.MineProgress > 0 {
		b := w.Block(p.MineX, p.MineY)
		if b != BAir && b.Info().Hardness > 0 {
			frac := p.MineProgress / b.Info().Hardness
			sx, sy := g.worldToScreen(float64(p.MineX), float64(p.MineY))
			vector.DrawFilledRect(screen, float32(sx), float32(sy), TileSize, TileSize,
				color.RGBA{0, 0, 0, uint8(70 * frac)}, false)
			n := int(frac * 9)
			for i := 0; i < n; i++ {
				cx := float32(sx) + float32(hash2(9, i, p.MineX)%TileSize)
				cy := float32(sy) + float32(hash2(7, i, p.MineY)%TileSize)
				vector.StrokeLine(screen, cx-3, cy-3, cx+3, cy+3, 1, color.RGBA{0, 0, 0, 200}, true)
				vector.StrokeLine(screen, cx+3, cy-3, cx-3, cy+3, 1, color.RGBA{0, 0, 0, 140}, true)
			}
		}
	}
}

// drawTileLit draws one tile with per-corner brightness (smooth lighting).
func drawTileLit(dst *ebiten.Image, tex *ebiten.Image, sx, sy float64, l00, l10, l01, l11, alpha float32) {
	x := float32(sx)
	y := float32(sy)
	vs := []ebiten.Vertex{
		{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: l00, ColorG: l00, ColorB: l00, ColorA: alpha},
		{DstX: x + TileSize, DstY: y, SrcX: TileSize, SrcY: 0, ColorR: l10, ColorG: l10, ColorB: l10, ColorA: alpha},
		{DstX: x, DstY: y + TileSize, SrcX: 0, SrcY: TileSize, ColorR: l01, ColorG: l01, ColorB: l01, ColorA: alpha},
		{DstX: x + TileSize, DstY: y + TileSize, SrcX: TileSize, SrcY: TileSize, ColorR: l11, ColorG: l11, ColorB: l11, ColorA: alpha},
	}
	dst.DrawTriangles(vs, tileIndices, tex, nil)
}

var tileIndices = []uint16{0, 1, 2, 1, 2, 3}

// drawGlows renders queued light blooms additively.
func (g *Game) drawGlows(screen *ebiten.Image) {
	for _, gp := range g.glows {
		sx, sy := g.worldToScreen(gp.x, gp.y)
		if sx < -160 || sx > ScreenW+160 || sy < -160 || sy > ScreenH+160 {
			continue
		}
		e := gp.block.Info().LightEmit
		r := float64(e) * 7
		a := 0.20
		// Torches and lava flicker.
		switch gp.block {
		case BTorch:
			f := math.Sin(g.World.Time*9+float64(int(gp.x)*7+int(gp.y)*13)) * 0.12
			r *= 1 + f
			a = 0.26 + f*0.3
		case BLava:
			a = 0.16 + 0.05*math.Sin(g.World.Time*3+gp.x*0.7)
		case BGlowstone:
			a = 0.25
		}
		drawGlow(screen, sx, sy, r, glowColorFor(gp.block), a)
	}
}

// drawShadow draws a soft contact shadow under an entity.
func (g *Game) drawShadow(screen *ebiten.Image, e *Entity) {
	if !e.OnGround {
		return
	}
	sx, sy := g.worldToScreen(e.X+e.W/2, e.Y+e.H)
	op := &ebiten.DrawImageOptions{}
	w := e.W * TileSize * 1.1
	h := w * 0.32
	op.GeoM.Scale(w/glowTexSize, h/glowTexSize)
	op.GeoM.Translate(sx-w/2, sy-h/2)
	op.ColorScale.Scale(0, 0, 0, 0.30)
	screen.DrawImage(glowTex, op)
}

func (g *Game) drawEntities(screen *ebiten.Image) {
	w := &g.World
	// Item drops.
	for _, d := range w.Drops {
		g.drawShadow(screen, &d.Entity)
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
		ex := sx + math.Cos(ang)*9
		ey := sy + math.Sin(ang)*9
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(ex), float32(ey), 2, color.RGBA{210, 200, 180, 255}, true)
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(sx-math.Cos(ang)*3), float32(sy-math.Sin(ang)*3), 3, color.RGBA{240, 240, 245, 255}, true)
	}
	// Mobs.
	for _, m := range w.Mobs {
		g.drawShadow(screen, &m.Entity)
		g.drawMob(screen, m)
	}
	// Particles.
	for _, p := range w.Particles {
		sx, sy := g.worldToScreen(p.X, p.Y)
		a := uint8(math.Min(1, p.Life*3) * 255)
		sz := float32(2 + math.Min(p.Life*2, 1.5))
		vector.DrawFilledRect(screen, float32(sx), float32(sy), sz, sz, color.RGBA{p.R, p.G, p.B, a}, false)
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
	head := c[0]
	if m.HurtTimer > 0.2 {
		head = color.RGBA{230, 80, 80, 255}
	}
	if m.Kind == MobCreeper && m.Fuse > 0 && int(m.Fuse*10)%2 == 0 {
		head = color.RGBA{255, 255, 255, 255}
	}
	sc := func(cc color.RGBA) color.RGBA {
		return color.RGBA{uint8(float32(cc.R) * l), uint8(float32(cc.G) * l), uint8(float32(cc.B) * l), cc.A}
	}
	x, y := float32(sx), float32(sy)
	bob := float32(math.Sin(float64(m.X)*4)) * 1.5
	if m.VX == 0 {
		bob = 0
	}

	// Outline for readability against dark caves.
	outline := color.RGBA{0, 0, 0, 110}
	vector.StrokeRect(screen, x-1, y-1, wpx+2, hpx+2, 1.5, outline, true)

	// Legs (animated), body, head with top highlight.
	legC := sc(scaleColor(c[1], 0.75))
	legW := wpx * 0.26
	legH := hpx * 0.22
	vector.DrawFilledRect(screen, x+wpx*0.12, y+hpx-legH+bob/2, legW, legH, legC, false)
	vector.DrawFilledRect(screen, x+wpx*0.62, y+hpx-legH-bob/2, legW, legH, legC, false)
	vector.DrawFilledRect(screen, x, y+hpx*0.32, wpx, hpx*0.48, sc(c[1]), false)
	vector.DrawFilledRect(screen, x, y+hpx*0.32, wpx, hpx*0.07, sc(scaleColor(c[1], 1.25)), false)
	vector.DrawFilledRect(screen, x, y, wpx, hpx*0.35, sc(head), false)
	vector.DrawFilledRect(screen, x, y, wpx, hpx*0.06, sc(scaleColor(head, 1.3)), false)

	// Faces.
	eye := color.RGBA{15, 15, 20, 255}
	glowEye := false
	switch m.Kind {
	case MobSpider:
		eye = color.RGBA{210, 40, 40, 255}
		glowEye = true
	case MobWarden:
		eye = color.RGBA{70, 230, 245, 255}
		glowEye = true
	case MobCreeper:
		eye = color.RGBA{20, 30, 20, 255}
	}
	ex := x + wpx*0.18
	fw := wpx * 0.42
	if g.Player.X < m.X {
		ex = x + wpx*0.12
	}
	eyeY := y + hpx*0.10
	vector.DrawFilledRect(screen, ex, eyeY, 3, 3, eye, false)
	vector.DrawFilledRect(screen, ex+fw, eyeY, 3, 3, eye, false)
	if m.Kind == MobCreeper {
		// Iconic frowning mouth.
		mx := ex + fw/2 - 1
		vector.DrawFilledRect(screen, mx, eyeY+4, 5, 4, eye, false)
		vector.DrawFilledRect(screen, mx-2, eyeY+6, 3, 4, eye, false)
		vector.DrawFilledRect(screen, mx+4, eyeY+6, 3, 4, eye, false)
	}
	if glowEye {
		esx, esy := float64(ex), float64(eyeY)
		drawGlow(screen, esx+1, esy+1, 7, eye, 0.5)
		drawGlow(screen, esx+float64(fw)+1, esy+1, 7, eye, 0.5)
	}

	// HP bar when damaged.
	if m.HP < m.Info().MaxHP {
		frac := float32(m.HP / m.Info().MaxHP)
		vector.DrawFilledRect(screen, x, y-7, wpx, 3.5, color.RGBA{20, 20, 25, 200}, false)
		vector.DrawFilledRect(screen, x, y-7, wpx*frac, 3.5, color.RGBA{225, 60, 60, 235}, false)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := &g.Player
	if p.Dead {
		return
	}
	g.drawShadow(screen, &p.Entity)
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
	x, y := float32(sx), float32(sy)
	walk := float32(0)
	if math.Abs(p.VX) > 0.5 && p.OnGround {
		walk = float32(math.Sin(p.X*4)) * 3
	}

	vector.StrokeRect(screen, x-1, y-1, wpx+2, hpx+2, 1.5, color.RGBA{0, 0, 0, 90}, true)

	// Legs with walk swing.
	legC := sc(color.RGBA{48, 60, 120, 255})
	legW := wpx * 0.34
	legY := y + hpx*0.64
	legH := hpx * 0.36
	vector.DrawFilledRect(screen, x+wpx*0.08+walk/2, legY, legW, legH, legC, false)
	vector.DrawFilledRect(screen, x+wpx*0.58-walk/2, legY, legW, legH, sc(color.RGBA{55, 70, 140, 255}), false)
	// Torso with shading.
	vector.DrawFilledRect(screen, x, y+hpx*0.28, wpx, hpx*0.38, sc(color.RGBA{58, 158, 158, 255}), false)
	vector.DrawFilledRect(screen, x, y+hpx*0.28, wpx, hpx*0.05, sc(color.RGBA{80, 190, 190, 255}), false)
	// Arm toward facing (behind held item).
	armX := x + wpx*0.65
	if p.Facing < 0 {
		armX = x + wpx*0.05
	}
	vector.DrawFilledRect(screen, armX, y+hpx*0.30-walk/3, wpx*0.30, hpx*0.30, sc(color.RGBA{225, 180, 140, 255}), false)
	// Head with hair.
	vector.DrawFilledRect(screen, x+wpx*0.08, y, wpx*0.84, hpx*0.28, sc(color.RGBA{228, 184, 144, 255}), false)
	vector.DrawFilledRect(screen, x+wpx*0.08, y, wpx*0.84, hpx*0.08, sc(color.RGBA{95, 65, 40, 255}), false)
	// Eye in facing direction.
	ex := x + wpx*0.60
	if p.Facing < 0 {
		ex = x + wpx*0.24
	}
	vector.DrawFilledRect(screen, ex, y+hpx*0.12, 3, 3, color.RGBA{45, 45, 70, 255}, false)
	// Held item swings when used.
	if !p.Held().Empty() {
		op := &ebiten.DrawImageOptions{}
		swing := p.SwingTimer / 0.25 // 1 -> 0
		ang := -0.5 + swing*1.4
		if p.Facing < 0 {
			ang = -ang
		}
		op.GeoM.Translate(-8, -8)
		op.GeoM.Rotate(ang)
		op.GeoM.Scale(0.8, 0.8)
		hx := sx + float64(wpx)*0.95
		if p.Facing < 0 {
			hx = sx - 4
		}
		op.GeoM.Translate(hx, sy+float64(hpx)*0.42)
		op.ColorScale.Scale(l, l, l, 1)
		screen.DrawImage(texForItem(p.Held().Item), op)
	}
}

func (g *Game) drawCursor(screen *ebiten.Image) {
	if g.UIOpen() {
		return
	}
	sx, sy := g.worldToScreen(float64(g.CursorX), float64(g.CursorY))
	pulse := 0.7 + 0.3*math.Sin(g.World.Time*5)
	a := uint8(200 * pulse)
	if !g.CursorReach {
		a = 45
	}
	c := color.NRGBA{255, 255, 255, a}
	x, y := float32(sx), float32(sy)
	const ln = 5
	// Corner ticks read cleaner than a full box.
	for _, p := range [4][4]float32{
		{x, y, 1, 1}, {x + TileSize, y, -1, 1},
		{x, y + TileSize, 1, -1}, {x + TileSize, y + TileSize, -1, -1},
	} {
		vector.StrokeLine(screen, p[0], p[1], p[0]+ln*p[2], p[1], 2, c, true)
		vector.StrokeLine(screen, p[0], p[1], p[0], p[1]+ln*p[3], 2, c, true)
	}
	if g.CursorReach {
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.NRGBA{255, 255, 255, 14}, false)
	}
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
	cl := func(v float64) uint8 {
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.RGBA{cl(float64(c.R) * f), cl(float64(c.G) * f), cl(float64(c.B) * f), 255}
}
