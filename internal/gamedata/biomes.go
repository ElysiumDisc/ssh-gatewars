package gamedata

// Biome defines a planet environment template.
type Biome struct {
	ID          string
	Name        string
	Description string
	FloorTile   TileType // primary floor tile
	WallTile    TileType // primary wall tile
	AccentTiles []TileType // extra tiles scattered in rooms
	Enemies     []string   // enemy type IDs that spawn here
	MinThreat   int        // minimum threat level for this biome
	MaxThreat   int
	Color       string // theme color for UI
}

// Biomes is the list of all available biomes.
var Biomes = []Biome{
	{
		ID:          "desert",
		Name:        "Desert World",
		Description: "Endless dunes and ancient ruins baking under twin suns.",
		FloorTile:   TileSand,
		WallTile:    TileWall,
		AccentTiles: []TileType{TileRubble, TileCrate},
		Enemies:     []string{"jaffa_warrior", "serpent_guard"},
		MinThreat:   1,
		MaxThreat:   5,
		Color:       "#CCAA44",
	},
	{
		ID:          "forest",
		Name:        "Forest World",
		Description: "Dense alien forests with Jaffa encampments hidden among the trees.",
		FloorTile:   TileFloor,
		WallTile:    TileTree,
		AccentTiles: []TileType{TileWater, TileRubble},
		Enemies:     []string{"jaffa_warrior", "jaffa_patrol"},
		MinThreat:   1,
		MaxThreat:   4,
		Color:       "#22AA22",
	},
	{
		ID:          "ice",
		Name:        "Ice World",
		Description: "Frozen wastelands concealing Goa'uld research facilities.",
		FloorTile:   TileIce,
		WallTile:    TileWall,
		AccentTiles: []TileType{TileWater, TileConsole},
		Enemies:     []string{"serpent_guard", "kull_warrior"},
		MinThreat:   3,
		MaxThreat:   7,
		Color:       "#AADDFF",
	},
	{
		ID:          "ruins",
		Name:        "Ancient Ruins",
		Description: "Crumbling Ancient outpost. Technology beyond comprehension hums in the walls.",
		FloorTile:   TileFloor,
		WallTile:    TileWall,
		AccentTiles: []TileType{TileConsole, TileRubble, TileCrate},
		Enemies:     []string{"jaffa_warrior", "serpent_guard", "kull_warrior"},
		MinThreat:   4,
		MaxThreat:   9,
		Color:       "#8866CC",
	},
	{
		ID:          "volcanic",
		Name:        "Volcanic World",
		Description: "Rivers of lava flow through Sokar's domain. The air burns.",
		FloorTile:   TileFloor,
		WallTile:    TileWall,
		AccentTiles: []TileType{TileLava, TileRubble},
		Enemies:     []string{"serpent_guard", "kull_warrior"},
		MinThreat:   5,
		MaxThreat:   10,
		Color:       "#FF4400",
	},
	{
		ID:          "station",
		Name:        "Space Station",
		Description: "A derelict Ha'tak vessel or orbital station. Corridors echo with distant footsteps.",
		FloorTile:   TileFloor,
		WallTile:    TileWall,
		AccentTiles: []TileType{TileConsole, TileCrate, TileDoor},
		Enemies:     []string{"jaffa_warrior", "serpent_guard", "kull_warrior"},
		MinThreat:   6,
		MaxThreat:   10,
		Color:       "#666688",
	},
}

// BiomeByID returns a biome by its ID, or the desert biome if not found.
func BiomeByID(id string) Biome {
	for _, b := range Biomes {
		if b.ID == id {
			return b
		}
	}
	return Biomes[0]
}
