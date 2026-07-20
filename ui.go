package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const slotPx = 36

// ---- HUD ----

func (g *Game) drawHUD(screen *ebiten.Image) {
	p := &g.Player

	// Hotbar.
	hbX := ScreenW/2 - 9*slotPx/2
	hbY := ScreenH - slotPx - 6
	for i := 0; i < 9; i++ {
		x := hbX + i*slotPx
		g.drawSlot(screen, x, hbY, &p.Inventory[i], i == p.Hotbar)
	}
	// Held item name.
	if !p.Held().Empty() {
		name := p.Held().Item.Info().Name
		ebitenutil.DebugPrintAt(screen, name, ScreenW/2-len(name)*3, hbY-16)
	}

	// Hearts.
	for i := 0; i < 10; i++ {
		x := float32(hbX + i*15)
		y := float32(hbY - 34)
		full := p.HP >= float64((i+1)*2)
		half := !full && p.HP > float64(i*2)
		c := color.RGBA{60, 20, 20, 255}
		if full {
			c = color.RGBA{225, 40, 40, 255}
		} else if half {
			c = color.RGBA{170, 40, 40, 255}
		}
		drawHeart(screen, x, y, c)
	}
	// Hunger.
	for i := 0; i < 10; i++ {
		x := float32(hbX + 9*slotPx - 15 - i*15)
		y := float32(hbY - 34)
		full := p.Hunger >= float64((i+1)*2)
		half := !full && p.Hunger > float64(i*2)
		c := color.RGBA{60, 45, 20, 255}
		if full {
			c = color.RGBA{200, 130, 40, 255}
		} else if half {
			c = color.RGBA{150, 100, 40, 255}
		}
		vector.DrawFilledRect(screen, x, y, 11, 11, c, false)
	}
	// Breath bubbles when underwater.
	if p.Breath < 10 {
		for i := 0; i < int(p.Breath); i++ {
			x := float32(hbX + 9*slotPx - 15 - i*15)
			vector.DrawFilledRect(screen, x, float32(hbY-50), 10, 10, color.RGBA{90, 150, 235, 255}, false)
		}
	}
	// XP.
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("XP %d", p.XP), hbX-52, hbY+10)

	// Info line.
	ebitenutil.DebugPrintAt(screen, fmtCoord(g), 8, 8)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Seed %d  Day %.1f", g.World.Seed, g.World.Time/DayLength+1), 8, 24)
	ebitenutil.DebugPrintAt(screen, "E inventory | LMB mine | RMB place/use | Q drop | Ctrl run", 8, 40)
	if g.msgTimer > 0 {
		ebitenutil.DebugPrintAt(screen, g.saveMsg, ScreenW/2-len(g.saveMsg)*3, 70)
	}
}

func drawHeart(screen *ebiten.Image, x, y float32, c color.RGBA) {
	vector.DrawFilledRect(screen, x, y+2, 11, 6, c, false)
	vector.DrawFilledRect(screen, x+1, y, 4, 3, c, false)
	vector.DrawFilledRect(screen, x+6, y, 4, 3, c, false)
	vector.DrawFilledRect(screen, x+2, y+8, 7, 3, c, false)
	vector.DrawFilledRect(screen, x+4, y+10, 3, 2, c, false)
}

func (g *Game) drawSlot(screen *ebiten.Image, x, y int, s *ItemStack, selected bool) {
	bg := color.RGBA{25, 25, 30, 200}
	border := color.RGBA{90, 90, 100, 255}
	if selected {
		border = color.RGBA{240, 240, 240, 255}
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), slotPx-2, slotPx-2, bg, false)
	vector.StrokeRect(screen, float32(x), float32(y), slotPx-2, slotPx-2, 2, border, false)
	if s != nil && !s.Empty() {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1.6, 1.6)
		op.GeoM.Translate(float64(x)+4, float64(y)+4)
		screen.DrawImage(texForItem(s.Item), op)
		if s.Count > 1 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprint(s.Count), x+slotPx-16, y+slotPx-18)
		}
	}
}

// ---- UI panels ----

// invSlotAt maps a screen point to an inventory slot index, or -1.
func (g *Game) invSlotAt(mx, my int) int {
	ox, oy := g.invOrigin()
	for i := 0; i < InvSlots; i++ {
		x := ox + (i%InvCols)*slotPx
		y := oy + (i/InvCols)*slotPx
		if mx >= x && mx < x+slotPx-2 && my >= y && my < y+slotPx-2 {
			return i
		}
	}
	return -1
}

func (g *Game) invOrigin() (int, int) {
	return ScreenW/2 - InvCols*slotPx/2, ScreenH - 4*slotPx - 60
}

// visibleRecipes returns craftable-list for the current station.
func (g *Game) visibleRecipes() []*Recipe {
	var out []*Recipe
	for i := range recipes {
		r := &recipes[i]
		switch g.Mode {
		case UIInventory:
			if r.Station == StationNone {
				out = append(out, r)
			}
		case UICraftTable:
			if r.Station == StationNone || r.Station == StationTable {
				out = append(out, r)
			}
		case UIFurnace:
			if r.Station == StationFurnace {
				out = append(out, r)
			}
		}
	}
	return out
}

const recipeRowH = 26
const recipeRows = 10

func (g *Game) updateUI() {
	mx, my := ebiten.CursorPosition()
	click := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	p := &g.Player

	// Scroll the recipe list.
	_, wy := ebiten.Wheel()
	if wy < 0 {
		g.craftScroll++
	} else if wy > 0 && g.craftScroll > 0 {
		g.craftScroll--
	}

	if !click {
		return
	}

	// Inventory slot interaction (all panels show the inventory).
	if idx := g.invSlotAt(mx, my); idx >= 0 {
		s := &p.Inventory[idx]
		if g.Mode == UIChest && g.Carried.Empty() && !s.Empty() {
			// Deposit into chest.
			ch := g.World.Chests[g.ChestKey]
			ch = append(ch, *s)
			g.World.Chests[g.ChestKey] = ch
			*s = ItemStack{}
			return
		}
		if g.Carried.Empty() {
			g.Carried = *s
			*s = ItemStack{}
		} else if s.Empty() {
			*s = g.Carried
			g.Carried = ItemStack{}
		} else if s.Item == g.Carried.Item {
			max := s.Item.Info().MaxStack
			take := min(g.Carried.Count, max-s.Count)
			s.Count += take
			g.Carried.Count -= take
			if g.Carried.Count == 0 {
				g.Carried = ItemStack{}
			}
		} else {
			*s, g.Carried = g.Carried, *s
		}
		return
	}

	// Armor slot.
	ax, ay := g.armorSlotPos()
	if mx >= ax && mx < ax+slotPx-2 && my >= ay && my < ay+slotPx-2 {
		if g.Carried.Empty() {
			g.Carried = p.Armor
			p.Armor = ItemStack{}
		} else if g.Carried.Item.Info().Armor > 0 && g.Carried.Count == 1 {
			p.Armor, g.Carried = g.Carried, p.Armor
		}
		return
	}

	// Chest contents.
	if g.Mode == UIChest {
		ch := g.World.Chests[g.ChestKey]
		ox, oy := g.chestOrigin()
		for i := range ch {
			x := ox + (i%InvCols)*slotPx
			y := oy + (i/InvCols)*slotPx
			if mx >= x && mx < x+slotPx-2 && my >= y && my < y+slotPx-2 {
				if !ch[i].Empty() {
					got := p.Give(ch[i])
					ch[i].Count -= got
					if ch[i].Count <= 0 {
						ch = append(ch[:i], ch[i+1:]...)
					}
					g.World.Chests[g.ChestKey] = ch
				}
				return
			}
		}
		return
	}

	// Recipe list.
	rs := g.visibleRecipes()
	rx, ry := g.recipeOrigin()
	for row := 0; row < recipeRows; row++ {
		i := row + g.craftScroll
		if i >= len(rs) {
			break
		}
		y := ry + row*recipeRowH
		if mx >= rx && mx < rx+300 && my >= y && my < y+recipeRowH-2 {
			if p.CanCraft(rs[i]) {
				p.Craft(rs[i])
			}
			return
		}
	}
}

func (g *Game) armorSlotPos() (int, int) {
	ox, oy := g.invOrigin()
	return ox - slotPx - 10, oy
}

func (g *Game) chestOrigin() (int, int) {
	ox, oy := g.invOrigin()
	return ox, oy - 3*slotPx - 30
}

func (g *Game) recipeOrigin() (int, int) {
	ox, oy := g.invOrigin()
	return ox + InvCols*slotPx + 20, oy - 3*slotPx - 30
}

func (g *Game) drawUI(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 120}, false)
	p := &g.Player
	ox, oy := g.invOrigin()

	title := map[UIMode]string{
		UIInventory:  "Inventory & Crafting",
		UICraftTable: "Crafting Table",
		UIFurnace:    "Furnace (uses 1 coal per smelt)",
		UIChest:      "Chest (click items to take, click inventory to deposit)",
	}[g.Mode]
	ebitenutil.DebugPrintAt(screen, title, ox, oy-3*slotPx-50)

	// Inventory grid.
	for i := 0; i < InvSlots; i++ {
		x := ox + (i%InvCols)*slotPx
		y := oy + (i/InvCols)*slotPx
		g.drawSlot(screen, x, y, &p.Inventory[i], i < 9 && i == p.Hotbar)
	}
	// Armor slot.
	ax, ay := g.armorSlotPos()
	g.drawSlot(screen, ax, ay, &p.Armor, false)
	ebitenutil.DebugPrintAt(screen, "Armor", ax-4, ay+slotPx)

	// Chest contents.
	if g.Mode == UIChest {
		ch := g.World.Chests[g.ChestKey]
		cx, cy := g.chestOrigin()
		n := len(ch)
		if n < 9 {
			n = 9
		}
		for i := 0; i < n; i++ {
			x := cx + (i%InvCols)*slotPx
			y := cy + (i/InvCols)*slotPx
			var s *ItemStack
			if i < len(ch) {
				s = &ch[i]
			}
			g.drawSlot(screen, x, y, s, false)
		}
	} else {
		// Recipe list.
		rs := g.visibleRecipes()
		rx, ry := g.recipeOrigin()
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Recipes (%d) - scroll wheel", len(rs)), rx, ry-16)
		for row := 0; row < recipeRows; row++ {
			i := row + g.craftScroll
			if i >= len(rs) {
				break
			}
			r := rs[i]
			y := ry + row*recipeRowH
			can := p.CanCraft(r)
			bg := color.RGBA{40, 20, 20, 220}
			if can {
				bg = color.RGBA{20, 50, 25, 220}
			}
			vector.DrawFilledRect(screen, float32(rx), float32(y), 300, recipeRowH-3, bg, false)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(1.2, 1.2)
			op.GeoM.Translate(float64(rx)+2, float64(y)+2)
			screen.DrawImage(texForItem(r.Out.Item), op)
			label := r.Out.Item.Info().Name
			if r.Out.Count > 1 {
				label = fmt.Sprintf("%s x%d", label, r.Out.Count)
			}
			ebitenutil.DebugPrintAt(screen, label, rx+26, y+4)
			// Ingredient icons on the right.
			ix := rx + 300 - len(r.In)*30
			for _, in := range r.In {
				iop := &ebiten.DrawImageOptions{}
				iop.GeoM.Translate(float64(ix), float64(y)+4)
				if p.Count(in.Item) < in.Count {
					iop.ColorScale.Scale(1, 0.4, 0.4, 1)
				}
				screen.DrawImage(texForItem(in.Item), iop)
				ebitenutil.DebugPrintAt(screen, fmt.Sprint(in.Count), ix+8, y+12)
				ix += 30
			}
		}
	}

	// Carried stack follows the mouse.
	if !g.Carried.Empty() {
		mx, my := ebiten.CursorPosition()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1.6, 1.6)
		op.GeoM.Translate(float64(mx)-12, float64(my)-12)
		screen.DrawImage(texForItem(g.Carried.Item), op)
		if g.Carried.Count > 1 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprint(g.Carried.Count), mx+4, my+4)
		}
	}
}
