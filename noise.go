package main

import "math"

// Perlin is seeded 1D/2D gradient noise used by world generation.
type Perlin struct {
	perm [512]int
}

func NewPerlin(seed int64) *Perlin {
	p := &Perlin{}
	src := seed
	next := func() int64 {
		src = src*6364136223846793005 + 1442695040888963407
		return src
	}
	base := make([]int, 256)
	for i := range base {
		base[i] = i
	}
	for i := 255; i > 0; i-- {
		j := int(uint64(next()) % uint64(i+1))
		base[i], base[j] = base[j], base[i]
	}
	for i := 0; i < 512; i++ {
		p.perm[i] = base[i&255]
	}
	return p
}

func fade(t float64) float64 { return t * t * t * (t*(t*6-15) + 10) }
func lerp(a, b, t float64) float64 { return a + t*(b-a) }

func grad2(h int, x, y float64) float64 {
	switch h & 7 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	case 3:
		return -x - y
	case 4:
		return x
	case 5:
		return -x
	case 6:
		return y
	default:
		return -y
	}
}

// Noise2 returns 2D Perlin noise in roughly [-1, 1].
func (p *Perlin) Noise2(x, y float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)
	u := fade(xf)
	v := fade(yf)
	aa := p.perm[p.perm[xi]+yi]
	ab := p.perm[p.perm[xi]+yi+1]
	ba := p.perm[p.perm[xi+1]+yi]
	bb := p.perm[p.perm[xi+1]+yi+1]
	return lerp(
		lerp(grad2(aa, xf, yf), grad2(ba, xf-1, yf), u),
		lerp(grad2(ab, xf, yf-1), grad2(bb, xf-1, yf-1), u),
		v)
}

// Octave2 layers several octaves of Noise2.
func (p *Perlin) Octave2(x, y float64, octaves int, persistence float64) float64 {
	total, freq, amp, max := 0.0, 1.0, 1.0, 0.0
	for i := 0; i < octaves; i++ {
		total += p.Noise2(x*freq, y*freq) * amp
		max += amp
		amp *= persistence
		freq *= 2
	}
	return total / max
}

// Noise1 returns 1D noise via a fixed y slice.
func (p *Perlin) Noise1(x float64) float64 { return p.Noise2(x, 37.417) }

// Octave1 layers octaves of 1D noise.
func (p *Perlin) Octave1(x float64, octaves int, persistence float64) float64 {
	total, freq, amp, max := 0.0, 1.0, 1.0, 0.0
	for i := 0; i < octaves; i++ {
		total += p.Noise1(x*freq) * amp
		max += amp
		amp *= persistence
		freq *= 2
	}
	return total / max
}

// hash2 gives a deterministic pseudo-random uint64 for integer coordinates,
// used for per-position decisions (structure placement, decoration).
func hash2(seed int64, x, y int) uint64 {
	h := uint64(seed) ^ 0x9E3779B97F4A7C15
	h ^= uint64(int64(x)) * 0xBF58476D1CE4E5B9
	h = (h ^ (h >> 27)) * 0x94D049BB133111EB
	h ^= uint64(int64(y)) * 0xD6E8FEB86659FD93
	h = (h ^ (h >> 31)) * 0xFF51AFD7ED558CCD
	return h ^ (h >> 33)
}

// hashFloat maps hash2 into [0,1).
func hashFloat(seed int64, x, y int) float64 {
	return float64(hash2(seed, x, y)>>11) / float64(1<<53)
}
