package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// glowTex is a soft radial gradient used for bloom-style light glow and the
// underground vignette. It is white with alpha fading from center to edge.
var glowTex *ebiten.Image

const glowTexSize = 128

func buildGlowTex() {
	glowTex = ebiten.NewImage(glowTexSize, glowTexSize)
	c := float64(glowTexSize) / 2
	for y := 0; y < glowTexSize; y++ {
		for x := 0; x < glowTexSize; x++ {
			dx, dy := float64(x)-c, float64(y)-c
			d := math.Sqrt(dx*dx+dy*dy) / c
			a := 1 - d
			if a < 0 {
				a = 0
			}
			a = a * a // steeper falloff, softer edge
			glowTex.Set(x, y, color.NRGBA{255, 255, 255, uint8(a * 255)})
		}
	}
}

// vignetteTex darkens screen edges; drawn stronger the deeper you are.
var vignetteTex *ebiten.Image

func buildVignette() {
	const vw, vh = ScreenW / 4, ScreenH / 4 // low-res is fine, it's smooth
	img := ebiten.NewImage(vw, vh)
	for y := 0; y < vh; y++ {
		for x := 0; x < vw; x++ {
			dx := (float64(x)/vw - 0.5) * 2
			dy := (float64(y)/vh - 0.5) * 2
			d := math.Sqrt(dx*dx*0.9 + dy*dy*1.3)
			a := (d - 0.55) / 0.65
			if a < 0 {
				a = 0
			}
			if a > 1 {
				a = 1
			}
			a = a * a
			// White with alpha so a draw-time tint can recolor it.
			img.Set(x, y, color.NRGBA{255, 255, 255, uint8(a * 255)})
		}
	}
	vignetteTex = img
}

// drawVignette overlays the vignette scaled to the screen at given strength.
func drawVignette(screen *ebiten.Image, alpha float64, tint color.RGBA) {
	if alpha <= 0.01 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	b := vignetteTex.Bounds()
	op.GeoM.Scale(float64(ScreenW)/float64(b.Dx()), float64(ScreenH)/float64(b.Dy()))
	op.ColorScale.ScaleWithColor(tint)
	op.ColorScale.ScaleAlpha(float32(alpha))
	screen.DrawImage(vignetteTex, op)
}

// glowColorFor returns the tint color used for a light-emitting block's bloom.
func glowColorFor(b Block) color.RGBA {
	switch b {
	case BTorch:
		return color.RGBA{255, 170, 80, 255}
	case BLava:
		return color.RGBA{255, 110, 40, 255}
	case BGlowstone:
		return color.RGBA{255, 225, 140, 255}
	case BGlowBerries:
		return color.RGBA{255, 200, 90, 255}
	case BSculkSensor:
		return color.RGBA{80, 220, 235, 255}
	case BAmethystCluster:
		return color.RGBA{190, 140, 240, 255}
	case BRedstoneOre, BDeepRedstoneOre:
		return color.RGBA{230, 60, 60, 255}
	case BSpawner:
		return color.RGBA{110, 150, 220, 255}
	default:
		return color.RGBA{255, 220, 170, 255}
	}
}

// drawGlow renders an additive soft-light bloom centered on a world point.
func drawGlow(screen *ebiten.Image, sx, sy, radiusPx float64, tint color.RGBA, alpha float64) {
	op := &ebiten.DrawImageOptions{}
	scale := radiusPx * 2 / glowTexSize
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(sx-radiusPx, sy-radiusPx)
	r := float32(tint.R) / 255
	g := float32(tint.G) / 255
	bl := float32(tint.B) / 255
	op.ColorScale.Scale(r, g, bl, float32(alpha))
	op.Blend = ebiten.BlendLighter
	screen.DrawImage(glowTex, op)
}
