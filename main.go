// A 2D Minecraft-like survival game with an underground focus.
//
// The world is procedurally generated from a random seed and extends
// (practically) infinitely to the left and right. Most of the content lies
// below the surface: cave systems, ore veins, lush caves, mushroom caves,
// the deep dark, amethyst geodes, abandoned mineshafts, dungeons and
// ancient vaults full of loot.
package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	seedFlag := flag.Int64("seed", 0, "world seed (0 = random)")
	deepFlag := flag.Bool("deep", false, "debug: spawn deep underground")
	flag.Parse()

	seed := *seedFlag
	if seed == 0 {
		var b [8]byte
		if _, err := rand.Read(b[:]); err != nil {
			log.Fatal(err)
		}
		seed = int64(binary.LittleEndian.Uint64(b[:]))
	}
	fmt.Println("world seed:", seed)

	buildTextures()
	g := NewGame(seed)
	if *deepFlag {
		g.TeleportToCave(220)
	}

	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle(fmt.Sprintf("Minecraft 2D (seed %d)", seed))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
