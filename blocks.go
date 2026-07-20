package main

// Block is a world tile type.
type Block uint8

const (
	BAir Block = iota
	BGrass
	BDirt
	BStone
	BCobble
	BBedrock
	BSand
	BGravel
	BLog
	BPlanks
	BLeaves
	BCoalOre
	BIronOre
	BCopperOre
	BGoldOre
	BRedstoneOre
	BLapisOre
	BDiamondOre
	BEmeraldOre
	BWater
	BLava
	BTorch
	BCraftTable
	BFurnace
	BChest
	BLadder
	BGlass
	BMossyCobble
	BSpawner
	BMossBlock
	BGlowBerries
	BSculk
	BSculkSensor
	BAmethyst
	BAmethystCluster
	BObsidian
	BMushroomStem
	BMushroomCap
	BBrownMushroom
	BRedMushroom
	BTallGrass
	BFlowerYellow
	BFlowerRed
	BClay
	BSandstone
	BSnow
	BIce
	BCactus
	BDeepslate
	BDeepCoalOre
	BDeepIronOre
	BDeepGoldOre
	BDeepRedstoneOre
	BDeepLapisOre
	BDeepDiamondOre
	BDeepEmeraldOre
	BGlowstone
	BFence
	BRail
	BStoneBricks
	BMossyStoneBricks
	BDripstone
	BTNT
	BBookshelf
	BWool
	BDoor
	BBlockCount
)

// ToolKind categorizes tools.
type ToolKind uint8

const (
	ToolNone ToolKind = iota
	ToolPickaxe
	ToolAxe
	ToolShovel
	ToolSword
)

// Tool tiers.
const (
	TierNone = iota
	TierWood
	TierStone
	TierIron
	TierGold
	TierDiamond
)

// BlockInfo describes static block properties.
type BlockInfo struct {
	Name      string
	Solid     bool    // blocks movement
	Opaque    bool    // blocks light & sky
	Hardness  float64 // seconds to break by hand
	BestTool  ToolKind
	MinTier   int  // minimum tool tier required to get drops
	NeedsTool bool // if true, breaking with lesser tool yields nothing
	LightEmit int  // 0..15
	Climbable bool
	Liquid    bool
	Replace   bool // can be replaced by placing a block (grass, water...)
	Drops     Item // item dropped (ItemNone = drops itself as block item)
	DropMin   int
	DropMax   int
}

var blockInfo [BBlockCount]BlockInfo

func init() {
	bi := func(b Block, inf BlockInfo) { blockInfo[b] = inf }
	bi(BAir, BlockInfo{Name: "Air", Replace: true})
	bi(BGrass, BlockInfo{Name: "Grass Block", Solid: true, Opaque: true, Hardness: 0.9, BestTool: ToolShovel, Drops: IDirt, DropMin: 1, DropMax: 1})
	bi(BDirt, BlockInfo{Name: "Dirt", Solid: true, Opaque: true, Hardness: 0.75, BestTool: ToolShovel})
	bi(BStone, BlockInfo{Name: "Stone", Solid: true, Opaque: true, Hardness: 2.2, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true, Drops: ICobble, DropMin: 1, DropMax: 1})
	bi(BCobble, BlockInfo{Name: "Cobblestone", Solid: true, Opaque: true, Hardness: 2.5, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BBedrock, BlockInfo{Name: "Bedrock", Solid: true, Opaque: true, Hardness: -1})
	bi(BSand, BlockInfo{Name: "Sand", Solid: true, Opaque: true, Hardness: 0.7, BestTool: ToolShovel})
	bi(BGravel, BlockInfo{Name: "Gravel", Solid: true, Opaque: true, Hardness: 0.8, BestTool: ToolShovel})
	// Log is non-solid so the player can walk through tree trunks on the
	// surface instead of being wall-blocked by them; still opaque so it
	// renders as wood and casts shade.
	bi(BLog, BlockInfo{Name: "Oak Log", Opaque: true, Hardness: 2.4, BestTool: ToolAxe})
	bi(BPlanks, BlockInfo{Name: "Oak Planks", Solid: true, Opaque: true, Hardness: 2.2, BestTool: ToolAxe})
	bi(BLeaves, BlockInfo{Name: "Leaves", Solid: true, Hardness: 0.35, Drops: IApple, DropMin: 0, DropMax: 1})
	bi(BCoalOre, BlockInfo{Name: "Coal Ore", Solid: true, Opaque: true, Hardness: 3.2, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true, Drops: ICoal, DropMin: 1, DropMax: 2})
	bi(BIronOre, BlockInfo{Name: "Iron Ore", Solid: true, Opaque: true, Hardness: 3.4, BestTool: ToolPickaxe, MinTier: TierStone, NeedsTool: true, Drops: IRawIron, DropMin: 1, DropMax: 1})
	bi(BCopperOre, BlockInfo{Name: "Copper Ore", Solid: true, Opaque: true, Hardness: 3.2, BestTool: ToolPickaxe, MinTier: TierStone, NeedsTool: true, Drops: IRawCopper, DropMin: 1, DropMax: 3})
	bi(BGoldOre, BlockInfo{Name: "Gold Ore", Solid: true, Opaque: true, Hardness: 3.6, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, Drops: IRawGold, DropMin: 1, DropMax: 1})
	bi(BRedstoneOre, BlockInfo{Name: "Redstone Ore", Solid: true, Opaque: true, Hardness: 3.6, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, LightEmit: 3, Drops: IRedstone, DropMin: 2, DropMax: 5})
	bi(BLapisOre, BlockInfo{Name: "Lapis Ore", Solid: true, Opaque: true, Hardness: 3.4, BestTool: ToolPickaxe, MinTier: TierStone, NeedsTool: true, Drops: ILapis, DropMin: 2, DropMax: 6})
	bi(BDiamondOre, BlockInfo{Name: "Diamond Ore", Solid: true, Opaque: true, Hardness: 4.0, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, Drops: IDiamond, DropMin: 1, DropMax: 1})
	bi(BEmeraldOre, BlockInfo{Name: "Emerald Ore", Solid: true, Opaque: true, Hardness: 4.0, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, Drops: IEmerald, DropMin: 1, DropMax: 1})
	bi(BWater, BlockInfo{Name: "Water", Liquid: true, Replace: true, Hardness: -1})
	bi(BLava, BlockInfo{Name: "Lava", Liquid: true, Replace: true, Hardness: -1, LightEmit: 15})
	bi(BTorch, BlockInfo{Name: "Torch", Hardness: 0.05, LightEmit: 14, Replace: false})
	bi(BCraftTable, BlockInfo{Name: "Crafting Table", Solid: true, Opaque: true, Hardness: 2.2, BestTool: ToolAxe})
	bi(BFurnace, BlockInfo{Name: "Furnace", Solid: true, Opaque: true, Hardness: 3.0, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BChest, BlockInfo{Name: "Chest", Solid: true, Hardness: 2.2, BestTool: ToolAxe})
	bi(BLadder, BlockInfo{Name: "Ladder", Hardness: 0.4, BestTool: ToolAxe, Climbable: true})
	bi(BGlass, BlockInfo{Name: "Glass", Solid: true, Hardness: 0.4, DropMin: 0, DropMax: 0, Drops: INone2})
	bi(BMossyCobble, BlockInfo{Name: "Mossy Cobblestone", Solid: true, Opaque: true, Hardness: 2.5, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BSpawner, BlockInfo{Name: "Monster Spawner", Solid: true, Hardness: 6.0, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true, LightEmit: 4, DropMin: 0, DropMax: 0, Drops: INone2})
	bi(BMossBlock, BlockInfo{Name: "Moss Block", Solid: true, Opaque: true, Hardness: 0.5})
	bi(BGlowBerries, BlockInfo{Name: "Glow Berries", Hardness: 0.05, LightEmit: 12, Drops: IGlowBerry, DropMin: 1, DropMax: 2})
	bi(BSculk, BlockInfo{Name: "Sculk", Solid: true, Opaque: true, Hardness: 1.2, Drops: IXPGem, DropMin: 1, DropMax: 1})
	bi(BSculkSensor, BlockInfo{Name: "Sculk Sensor", Solid: true, Hardness: 1.5, LightEmit: 6, Drops: IXPGem, DropMin: 2, DropMax: 3})
	bi(BAmethyst, BlockInfo{Name: "Amethyst Block", Solid: true, Opaque: true, Hardness: 2.5, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BAmethystCluster, BlockInfo{Name: "Amethyst Cluster", Hardness: 1.5, LightEmit: 8, BestTool: ToolPickaxe, Drops: IAmethystShard, DropMin: 2, DropMax: 4})
	bi(BObsidian, BlockInfo{Name: "Obsidian", Solid: true, Opaque: true, Hardness: 12.0, BestTool: ToolPickaxe, MinTier: TierDiamond, NeedsTool: true})
	bi(BMushroomStem, BlockInfo{Name: "Mushroom Stem", Solid: true, Opaque: true, Hardness: 0.4})
	bi(BMushroomCap, BlockInfo{Name: "Mushroom Cap", Solid: true, Opaque: true, Hardness: 0.4, Drops: IBrownMushroomItem, DropMin: 0, DropMax: 2})
	bi(BBrownMushroom, BlockInfo{Name: "Brown Mushroom", Hardness: 0.05, LightEmit: 2, Replace: true, Drops: IBrownMushroomItem, DropMin: 1, DropMax: 1})
	bi(BRedMushroom, BlockInfo{Name: "Red Mushroom", Hardness: 0.05, Replace: true, Drops: IRedMushroomItem, DropMin: 1, DropMax: 1})
	bi(BTallGrass, BlockInfo{Name: "Tall Grass", Hardness: 0.05, Replace: true, Drops: IWheatSeeds, DropMin: 0, DropMax: 1})
	bi(BFlowerYellow, BlockInfo{Name: "Dandelion", Hardness: 0.05, Replace: true})
	bi(BFlowerRed, BlockInfo{Name: "Poppy", Hardness: 0.05, Replace: true})
	bi(BClay, BlockInfo{Name: "Clay", Solid: true, Opaque: true, Hardness: 0.8, BestTool: ToolShovel, Drops: IClayBall, DropMin: 4, DropMax: 4})
	bi(BSandstone, BlockInfo{Name: "Sandstone", Solid: true, Opaque: true, Hardness: 1.8, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BSnow, BlockInfo{Name: "Snow Block", Solid: true, Opaque: true, Hardness: 0.4, BestTool: ToolShovel, Drops: ISnowball, DropMin: 4, DropMax: 4})
	bi(BIce, BlockInfo{Name: "Ice", Solid: true, Hardness: 0.7, BestTool: ToolPickaxe, DropMin: 0, DropMax: 0, Drops: INone2})
	bi(BCactus, BlockInfo{Name: "Cactus", Solid: true, Hardness: 0.6})
	bi(BDeepslate, BlockInfo{Name: "Deepslate", Solid: true, Opaque: true, Hardness: 3.8, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true, Drops: ICobbledDeepslate, DropMin: 1, DropMax: 1})
	bi(BDeepCoalOre, BlockInfo{Name: "Deepslate Coal Ore", Solid: true, Opaque: true, Hardness: 4.6, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true, Drops: ICoal, DropMin: 1, DropMax: 2})
	bi(BDeepIronOre, BlockInfo{Name: "Deepslate Iron Ore", Solid: true, Opaque: true, Hardness: 4.8, BestTool: ToolPickaxe, MinTier: TierStone, NeedsTool: true, Drops: IRawIron, DropMin: 1, DropMax: 1})
	bi(BDeepGoldOre, BlockInfo{Name: "Deepslate Gold Ore", Solid: true, Opaque: true, Hardness: 5.0, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, Drops: IRawGold, DropMin: 1, DropMax: 1})
	bi(BDeepRedstoneOre, BlockInfo{Name: "Deepslate Redstone Ore", Solid: true, Opaque: true, Hardness: 5.0, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, LightEmit: 3, Drops: IRedstone, DropMin: 2, DropMax: 5})
	bi(BDeepLapisOre, BlockInfo{Name: "Deepslate Lapis Ore", Solid: true, Opaque: true, Hardness: 4.8, BestTool: ToolPickaxe, MinTier: TierStone, NeedsTool: true, Drops: ILapis, DropMin: 2, DropMax: 6})
	bi(BDeepDiamondOre, BlockInfo{Name: "Deepslate Diamond Ore", Solid: true, Opaque: true, Hardness: 5.5, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, Drops: IDiamond, DropMin: 1, DropMax: 1})
	bi(BDeepEmeraldOre, BlockInfo{Name: "Deepslate Emerald Ore", Solid: true, Opaque: true, Hardness: 5.5, BestTool: ToolPickaxe, MinTier: TierIron, NeedsTool: true, Drops: IEmerald, DropMin: 1, DropMax: 1})
	bi(BGlowstone, BlockInfo{Name: "Glowstone", Solid: true, Hardness: 0.6, LightEmit: 15, Drops: IGlowstoneDust, DropMin: 2, DropMax: 4})
	bi(BFence, BlockInfo{Name: "Oak Fence", Solid: true, Hardness: 2.2, BestTool: ToolAxe})
	bi(BRail, BlockInfo{Name: "Rail", Hardness: 0.7, BestTool: ToolPickaxe})
	bi(BStoneBricks, BlockInfo{Name: "Stone Bricks", Solid: true, Opaque: true, Hardness: 2.5, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BMossyStoneBricks, BlockInfo{Name: "Mossy Stone Bricks", Solid: true, Opaque: true, Hardness: 2.5, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BDripstone, BlockInfo{Name: "Dripstone", Solid: true, Hardness: 1.5, BestTool: ToolPickaxe, MinTier: TierWood, NeedsTool: true})
	bi(BTNT, BlockInfo{Name: "TNT", Solid: true, Opaque: true, Hardness: 0.1})
	bi(BBookshelf, BlockInfo{Name: "Bookshelf", Solid: true, Opaque: true, Hardness: 2.0, BestTool: ToolAxe})
	bi(BWool, BlockInfo{Name: "Wool", Solid: true, Opaque: true, Hardness: 1.0})
	bi(BDoor, BlockInfo{Name: "Oak Door", Hardness: 2.2, BestTool: ToolAxe, Climbable: false})
}

func (b Block) Info() *BlockInfo { return &blockInfo[b] }
func (b Block) Solid() bool      { return blockInfo[b].Solid }
func (b Block) Opaque() bool     { return blockInfo[b].Opaque }
func (b Block) Liquid() bool     { return blockInfo[b].Liquid }
