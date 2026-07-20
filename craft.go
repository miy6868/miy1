package main

// Station is where a recipe can be crafted.
type Station int

const (
	StationNone    Station = iota // 2x2 inventory crafting
	StationTable                  // requires crafting table
	StationFurnace                // requires furnace (smelting)
)

// Recipe converts ingredients into an output stack.
type Recipe struct {
	Out     ItemStack
	In      []ItemStack
	Station Station
}

var recipes = []Recipe{
	// Basics.
	{Out: ItemStack{IPlanks, 4}, In: []ItemStack{{ILogItem, 1}}},
	{Out: ItemStack{IStick, 4}, In: []ItemStack{{IPlanks, 2}}},
	{Out: ItemStack{ICraftTable, 1}, In: []ItemStack{{IPlanks, 4}}},
	{Out: ItemStack{ITorch, 4}, In: []ItemStack{{ICoal, 1}, {IStick, 1}}},
	{Out: ItemStack{IChestItem, 1}, In: []ItemStack{{IPlanks, 8}}, Station: StationTable},
	{Out: ItemStack{IFurnaceItem, 1}, In: []ItemStack{{ICobble, 8}}, Station: StationTable},
	{Out: ItemStack{ILadderItem, 3}, In: []ItemStack{{IStick, 7}}, Station: StationTable},
	{Out: ItemStack{IFenceItem, 3}, In: []ItemStack{{IStick, 4}, {IPlanks, 2}}, Station: StationTable},
	{Out: ItemStack{IDoorItem, 1}, In: []ItemStack{{IPlanks, 6}}, Station: StationTable},
	{Out: ItemStack{IStoneBricksI, 4}, In: []ItemStack{{IStoneItem, 4}}, Station: StationTable},
	{Out: ItemStack{Item(BSandstone), 1}, In: []ItemStack{{ISandItem, 4}}, Station: StationTable},
	{Out: ItemStack{IRailItem, 8}, In: []ItemStack{{IIronIngot, 3}, {IStick, 1}}, Station: StationTable},
	{Out: ItemStack{IGlowstoneBlk, 1}, In: []ItemStack{{IGlowstoneDust, 4}}},
	{Out: ItemStack{Item(BWool), 1}, In: []ItemStack{{IString, 4}}},
	{Out: ItemStack{IBookshelfI, 1}, In: []ItemStack{{IPlanks, 6}, {ILeather, 3}}, Station: StationTable},
	{Out: ItemStack{ITNTItem, 1}, In: []ItemStack{{IGunpowder, 5}, {ISandItem, 4}}, Station: StationTable},
	{Out: ItemStack{IBucket, 1}, In: []ItemStack{{IIronIngot, 3}}, Station: StationTable},
	{Out: ItemStack{IFlintSteel, 1}, In: []ItemStack{{IIronIngot, 1}, {IGravelI, 1}}, Station: StationTable},
	{Out: ItemStack{ICompass, 1}, In: []ItemStack{{IIronIngot, 4}, {IRedstone, 1}}, Station: StationTable},
	{Out: ItemStack{IBow, 1}, In: []ItemStack{{IStick, 3}, {IString, 3}}, Station: StationTable},
	{Out: ItemStack{IArrow, 4}, In: []ItemStack{{IStick, 1}, {IGravelI, 1}, {IFeather, 1}}, Station: StationTable},
	{Out: ItemStack{IMushroomStew, 1}, In: []ItemStack{{IBrownMushroomItem, 1}, {IRedMushroomItem, 1}}},
	{Out: ItemStack{IBread, 1}, In: []ItemStack{{IWheatSeeds, 3}}},
	{Out: ItemStack{IGoldenApple, 1}, In: []ItemStack{{IApple, 1}, {IGoldIngot, 4}}, Station: StationTable},

	// Tools: wood.
	{Out: ItemStack{IWoodPickaxe, 1}, In: []ItemStack{{IPlanks, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IWoodAxe, 1}, In: []ItemStack{{IPlanks, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IWoodShovel, 1}, In: []ItemStack{{IPlanks, 1}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IWoodSword, 1}, In: []ItemStack{{IPlanks, 2}, {IStick, 1}}, Station: StationTable},
	// Stone.
	{Out: ItemStack{IStonePickaxe, 1}, In: []ItemStack{{ICobble, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IStoneAxe, 1}, In: []ItemStack{{ICobble, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IStoneShovel, 1}, In: []ItemStack{{ICobble, 1}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IStoneSword, 1}, In: []ItemStack{{ICobble, 2}, {IStick, 1}}, Station: StationTable},
	// Iron.
	{Out: ItemStack{IIronPickaxe, 1}, In: []ItemStack{{IIronIngot, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IIronAxe, 1}, In: []ItemStack{{IIronIngot, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IIronShovel, 1}, In: []ItemStack{{IIronIngot, 1}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IIronSword, 1}, In: []ItemStack{{IIronIngot, 2}, {IStick, 1}}, Station: StationTable},
	// Gold.
	{Out: ItemStack{IGoldPickaxe, 1}, In: []ItemStack{{IGoldIngot, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IGoldSword, 1}, In: []ItemStack{{IGoldIngot, 2}, {IStick, 1}}, Station: StationTable},
	// Diamond.
	{Out: ItemStack{IDiamondPickaxe, 1}, In: []ItemStack{{IDiamond, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IDiamondAxe, 1}, In: []ItemStack{{IDiamond, 3}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IDiamondShovel, 1}, In: []ItemStack{{IDiamond, 1}, {IStick, 2}}, Station: StationTable},
	{Out: ItemStack{IDiamondSword, 1}, In: []ItemStack{{IDiamond, 2}, {IStick, 1}}, Station: StationTable},
	// Armor.
	{Out: ItemStack{ILeatherArmor, 1}, In: []ItemStack{{ILeather, 8}}, Station: StationTable},
	{Out: ItemStack{IIronArmor, 1}, In: []ItemStack{{IIronIngot, 8}}, Station: StationTable},
	{Out: ItemStack{IDiamondArmor, 1}, In: []ItemStack{{IDiamond, 8}}, Station: StationTable},

	// Smelting (furnace consumes 1 coal per craft, added at lookup).
	{Out: ItemStack{IIronIngot, 1}, In: []ItemStack{{IRawIron, 1}}, Station: StationFurnace},
	{Out: ItemStack{IGoldIngot, 1}, In: []ItemStack{{IRawGold, 1}}, Station: StationFurnace},
	{Out: ItemStack{ICopperIngot, 1}, In: []ItemStack{{IRawCopper, 1}}, Station: StationFurnace},
	{Out: ItemStack{IGlassItem, 1}, In: []ItemStack{{ISandItem, 1}}, Station: StationFurnace},
	{Out: ItemStack{IStoneItem, 1}, In: []ItemStack{{ICobble, 1}}, Station: StationFurnace},
	{Out: ItemStack{IBrick, 1}, In: []ItemStack{{IClayBall, 1}}, Station: StationFurnace},
	{Out: ItemStack{ISteak, 1}, In: []ItemStack{{IBeef, 1}}, Station: StationFurnace},
	{Out: ItemStack{ICookedPorkchop, 1}, In: []ItemStack{{IPorkchop, 1}}, Station: StationFurnace},
	{Out: ItemStack{ICookedChicken, 1}, In: []ItemStack{{IChicken, 1}}, Station: StationFurnace},
}

// IGravelI is the gravel block item (used as flint source in recipes).
const IGravelI = Item(BGravel)

// CanCraft reports whether the player has the ingredients (and fuel for
// furnace recipes).
func (p *Player) CanCraft(r *Recipe) bool {
	for _, in := range r.In {
		if p.Count(in.Item) < in.Count {
			return false
		}
	}
	if r.Station == StationFurnace && p.Count(ICoal) < 1 {
		return false
	}
	return true
}

// Craft consumes ingredients and gives the output.
func (p *Player) Craft(r *Recipe) bool {
	if !p.CanCraft(r) {
		return false
	}
	for _, in := range r.In {
		p.Consume(in.Item, in.Count)
	}
	if r.Station == StationFurnace {
		p.Consume(ICoal, 1)
	}
	given := p.Give(r.Out)
	if given < r.Out.Count {
		return true // partial fit: remainder is lost overboard, keep it simple
	}
	return true
}
