package gamedata

// Faction represents an NPC faction for the reputation system.
type Faction struct {
	ID          string
	Name        string
	Description string
	Color       string
}

// Factions are the NPC factions players can build reputation with.
var Factions = []Faction{
	{
		ID:          "tokra",
		Name:        "Tok'ra",
		Description: "Symbiote resistance movement. Experts in espionage and infiltration.",
		Color:       "#CCAA44",
	},
	{
		ID:          "free_jaffa",
		Name:        "Free Jaffa Nation",
		Description: "Former servants of the Goa'uld fighting for freedom.",
		Color:       "#CC4444",
	},
	{
		ID:          "asgard",
		Name:        "Asgard",
		Description: "Advanced alien race. Protectors of developing worlds.",
		Color:       "#AAAACC",
	},
	{
		ID:          "lucian_alliance",
		Name:        "Lucian Alliance",
		Description: "Criminal syndicate filling the power vacuum left by the Goa'uld.",
		Color:       "#44AA44",
	},
}
