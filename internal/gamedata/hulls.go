package gamedata

// Hull size constants.
const (
	HullSmall  = iota // Al'kesh
	HullMedium        // Ha'tak
	HullLarge         // O'Neill-class
	HullHuge          // City-ship
	HullCount
)

// HullInfo describes a ship hull.
type HullInfo struct {
	Name     string
	SGName   string // Stargate-themed name
	Space    int    // component capacity
	HP       int    // hit points
	BaseCost float64
}

// Hulls defines the 4 hull sizes.
var Hulls = [HullCount]HullInfo{
	{Name: "Small", SGName: "Al'kesh", Space: 20, HP: 3, BaseCost: 10},
	{Name: "Medium", SGName: "Ha'tak", Space: 60, HP: 18, BaseCost: 50},
	{Name: "Large", SGName: "O'Neill-class", Space: 120, HP: 60, BaseCost: 200},
	{Name: "Huge", SGName: "City-ship", Space: 250, HP: 150, BaseCost: 600},
}
