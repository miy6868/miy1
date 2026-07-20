package main

// Item identifies an inventory item. Values below 256 are block items
// (Item(b) places Block(b)); values from 256 up are non-block items.
type Item uint16

const (
	INone Item = 0 // same as BAir: empty slot
)

const (
	IStick Item = 256 + iota
	ICoal
	IRawIron
	IIronIngot
	IRawGold
	IGoldIngot
	IRawCopper
	ICopperIngot
	IDiamond
	IEmerald
	IRedstone
	ILapis
	IAmethystShard
	IGlowstoneDust
	IXPGem
	IClayBall
	IBrick
	ISnowball
	IGlowBerry
	IApple
	IGoldenApple
	IBread
	IWheatSeeds
	IPorkchop
	ICookedPorkchop
	IBeef
	ISteak
	IChicken
	ICookedChicken
	IBrownMushroomItem
	IRedMushroomItem
	IMushroomStew
	IBone
	IString
	IGunpowder
	IRottenFlesh
	ISpiderEye
	IEnderPearl
	ISlimeball
	IFeather
	ILeather
	IArrow
	IBow
	IBucket
	IWaterBucket
	ILavaBucket
	IFlintSteel
	ICompass
	ICobbledDeepslate
	IWoodPickaxe
	IStonePickaxe
	IIronPickaxe
	IGoldPickaxe
	IDiamondPickaxe
	IWoodAxe
	IStoneAxe
	IIronAxe
	IGoldAxe
	IDiamondAxe
	IWoodShovel
	IStoneShovel
	IIronShovel
	IGoldShovel
	IDiamondShovel
	IWoodSword
	IStoneSword
	IIronSword
	IGoldSword
	IDiamondSword
	ILeatherArmor
	IIronArmor
	IDiamondArmor
	IItemCount
	INone2 = INone // marker used in BlockInfo.Drops for "drops nothing"
)

// Block item aliases for readability in recipes/drops.
const (
	IDirt         = Item(BDirt)
	ICobble       = Item(BCobble)
	IPlanks       = Item(BPlanks)
	ILogItem      = Item(BLog)
	ITorch        = Item(BTorch)
	IStoneItem    = Item(BStone)
	ISandItem     = Item(BSand)
	IGlassItem    = Item(BGlass)
	ICraftTable   = Item(BCraftTable)
	IFurnaceItem  = Item(BFurnace)
	IChestItem    = Item(BChest)
	ILadderItem   = Item(BLadder)
	IFenceItem    = Item(BFence)
	IRailItem     = Item(BRail)
	IGlowstoneBlk = Item(BGlowstone)
	IStoneBricksI = Item(BStoneBricks)
	IBookshelfI   = Item(BBookshelf)
	ITNTItem      = Item(BTNT)
	IDoorItem     = Item(BDoor)
	IObsidianItem = Item(BObsidian)
	ISandstoneI   = Item(BSandstone)
	IWoolItem     = Item(BWool)
)

// ItemInfo describes static item properties.
type ItemInfo struct {
	Name      string
	MaxStack  int
	Tool      ToolKind
	Tier      int
	Damage    float64 // melee damage (hearts*2)
	Food      int     // hunger points restored
	Heal      float64 // instant health restored
	Armor     float64 // damage reduction fraction when in armor slot
	FuelSmelt int     // number of items one unit smelts as fuel
}

var itemInfo = map[Item]ItemInfo{}

func init() {
	reg := func(i Item, inf ItemInfo) {
		if inf.MaxStack == 0 {
			inf.MaxStack = 64
		}
		itemInfo[i] = inf
	}
	reg(IStick, ItemInfo{Name: "Stick"})
	reg(ICoal, ItemInfo{Name: "Coal", FuelSmelt: 8})
	reg(IRawIron, ItemInfo{Name: "Raw Iron"})
	reg(IIronIngot, ItemInfo{Name: "Iron Ingot"})
	reg(IRawGold, ItemInfo{Name: "Raw Gold"})
	reg(IGoldIngot, ItemInfo{Name: "Gold Ingot"})
	reg(IRawCopper, ItemInfo{Name: "Raw Copper"})
	reg(ICopperIngot, ItemInfo{Name: "Copper Ingot"})
	reg(IDiamond, ItemInfo{Name: "Diamond"})
	reg(IEmerald, ItemInfo{Name: "Emerald"})
	reg(IRedstone, ItemInfo{Name: "Redstone Dust"})
	reg(ILapis, ItemInfo{Name: "Lapis Lazuli"})
	reg(IAmethystShard, ItemInfo{Name: "Amethyst Shard"})
	reg(IGlowstoneDust, ItemInfo{Name: "Glowstone Dust"})
	reg(IXPGem, ItemInfo{Name: "Echo Shard"})
	reg(IClayBall, ItemInfo{Name: "Clay Ball"})
	reg(IBrick, ItemInfo{Name: "Brick"})
	reg(ISnowball, ItemInfo{Name: "Snowball", MaxStack: 16})
	reg(IGlowBerry, ItemInfo{Name: "Glow Berries", Food: 2})
	reg(IApple, ItemInfo{Name: "Apple", Food: 4})
	reg(IGoldenApple, ItemInfo{Name: "Golden Apple", Food: 4, Heal: 8})
	reg(IBread, ItemInfo{Name: "Bread", Food: 5})
	reg(IWheatSeeds, ItemInfo{Name: "Seeds", Food: 1})
	reg(IPorkchop, ItemInfo{Name: "Raw Porkchop", Food: 3})
	reg(ICookedPorkchop, ItemInfo{Name: "Cooked Porkchop", Food: 8})
	reg(IBeef, ItemInfo{Name: "Raw Beef", Food: 3})
	reg(ISteak, ItemInfo{Name: "Steak", Food: 8})
	reg(IChicken, ItemInfo{Name: "Raw Chicken", Food: 2})
	reg(ICookedChicken, ItemInfo{Name: "Cooked Chicken", Food: 6})
	reg(IBrownMushroomItem, ItemInfo{Name: "Brown Mushroom", Food: 1})
	reg(IRedMushroomItem, ItemInfo{Name: "Red Mushroom"})
	reg(IMushroomStew, ItemInfo{Name: "Mushroom Stew", MaxStack: 1, Food: 6})
	reg(IBone, ItemInfo{Name: "Bone"})
	reg(IString, ItemInfo{Name: "String"})
	reg(IGunpowder, ItemInfo{Name: "Gunpowder"})
	reg(IRottenFlesh, ItemInfo{Name: "Rotten Flesh", Food: 2})
	reg(ISpiderEye, ItemInfo{Name: "Spider Eye", Food: 1})
	reg(IEnderPearl, ItemInfo{Name: "Ender Pearl", MaxStack: 16})
	reg(ISlimeball, ItemInfo{Name: "Slimeball"})
	reg(IFeather, ItemInfo{Name: "Feather"})
	reg(ILeather, ItemInfo{Name: "Leather"})
	reg(IArrow, ItemInfo{Name: "Arrow"})
	reg(IBow, ItemInfo{Name: "Bow", MaxStack: 1, Damage: 5})
	reg(IBucket, ItemInfo{Name: "Bucket", MaxStack: 1})
	reg(IWaterBucket, ItemInfo{Name: "Water Bucket", MaxStack: 1})
	reg(ILavaBucket, ItemInfo{Name: "Lava Bucket", MaxStack: 1, FuelSmelt: 64})
	reg(IFlintSteel, ItemInfo{Name: "Flint and Steel", MaxStack: 1})
	reg(ICompass, ItemInfo{Name: "Compass", MaxStack: 1})
	reg(ICobbledDeepslate, ItemInfo{Name: "Cobbled Deepslate"})

	tool := func(i Item, name string, kind ToolKind, tier int, dmg float64) {
		reg(i, ItemInfo{Name: name, MaxStack: 1, Tool: kind, Tier: tier, Damage: dmg})
	}
	tool(IWoodPickaxe, "Wooden Pickaxe", ToolPickaxe, TierWood, 2)
	tool(IStonePickaxe, "Stone Pickaxe", ToolPickaxe, TierStone, 3)
	tool(IIronPickaxe, "Iron Pickaxe", ToolPickaxe, TierIron, 4)
	tool(IGoldPickaxe, "Golden Pickaxe", ToolPickaxe, TierGold, 2)
	tool(IDiamondPickaxe, "Diamond Pickaxe", ToolPickaxe, TierDiamond, 5)
	tool(IWoodAxe, "Wooden Axe", ToolAxe, TierWood, 3)
	tool(IStoneAxe, "Stone Axe", ToolAxe, TierStone, 4)
	tool(IIronAxe, "Iron Axe", ToolAxe, TierIron, 5)
	tool(IGoldAxe, "Golden Axe", ToolAxe, TierGold, 3)
	tool(IDiamondAxe, "Diamond Axe", ToolAxe, TierDiamond, 6)
	tool(IWoodShovel, "Wooden Shovel", ToolShovel, TierWood, 1)
	tool(IStoneShovel, "Stone Shovel", ToolShovel, TierStone, 2)
	tool(IIronShovel, "Iron Shovel", ToolShovel, TierIron, 3)
	tool(IGoldShovel, "Golden Shovel", ToolShovel, TierGold, 1)
	tool(IDiamondShovel, "Diamond Shovel", ToolShovel, TierDiamond, 4)
	tool(IWoodSword, "Wooden Sword", ToolSword, TierWood, 4)
	tool(IStoneSword, "Stone Sword", ToolSword, TierStone, 5)
	tool(IIronSword, "Iron Sword", ToolSword, TierIron, 6)
	tool(IGoldSword, "Golden Sword", ToolSword, TierGold, 4)
	tool(IDiamondSword, "Diamond Sword", ToolSword, TierDiamond, 7)
	reg(ILeatherArmor, ItemInfo{Name: "Leather Tunic", MaxStack: 1, Armor: 0.2})
	reg(IIronArmor, ItemInfo{Name: "Iron Chestplate", MaxStack: 1, Armor: 0.4})
	reg(IDiamondArmor, ItemInfo{Name: "Diamond Chestplate", MaxStack: 1, Armor: 0.6})
}

// Info returns item properties. Block items derive from their block.
func (i Item) Info() ItemInfo {
	if inf, ok := itemInfo[i]; ok {
		return inf
	}
	if i < 256 && Block(i) < BBlockCount {
		return ItemInfo{Name: Block(i).Info().Name, MaxStack: 64}
	}
	return ItemInfo{Name: "?", MaxStack: 64}
}

// IsBlock reports whether the item places a block.
func (i Item) IsBlock() bool { return i > 0 && i < 256 && Block(i) < BBlockCount }

// ItemStack is a quantity of one item.
type ItemStack struct {
	Item  Item
	Count int
}

func (s *ItemStack) Empty() bool { return s.Item == INone || s.Count <= 0 }
