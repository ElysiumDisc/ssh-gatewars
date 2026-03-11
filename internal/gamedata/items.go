package gamedata

// ItemSlot is where an item can be equipped.
type ItemSlot int

const (
	SlotNone      ItemSlot = iota
	SlotWeapon
	SlotArmor
	SlotAccessory
)

// WeaponType classifies weapons as melee or ranged.
type WeaponType int

const (
	WeaponMelee  WeaponType = iota
	WeaponRanged
	WeaponThrown
	WeaponPlaced
)

// FireMode describes weapon fire modes.
type FireMode int

const (
	FireSingle FireMode = iota
	FireBurst         // 3-round burst, less accurate
	FireOvercharge    // 2x damage, high ammo cost
)

// ItemDef defines a type of item.
type ItemDef struct {
	ID          string
	Name        string
	Description string
	Slot        ItemSlot
	Attack      int // damage (weapons) or damage reduction (armor)
	Defense     int // damage reduction (armor)
	HealAmount  int // HP restored (consumables)
	Consumable  bool
	Rarity      int // 1=common, 2=uncommon, 3=rare, 4=epic
	Value       int // trade value

	// Ranged weapon stats
	WType       WeaponType
	Range       int      // max range in tiles (0 = melee only)
	Accuracy    int      // base accuracy percentage (0 = use default 85)
	AmmoType    string   // ammo item ID required
	ClipSize    int      // max ammo in clip
	ReloadTicks int      // ticks to reload
	FireModes   []FireMode // available fire modes
	Glyph       rune     // projectile glyph
	ProjColor   string   // projectile color
	ProjSpeed   int      // tiles per tick for projectile

	// Armor stats
	Weight    int // 0=light, 1=medium, 2=heavy (affects move speed)
	ArmorSlot string // "head", "torso", "full"
}

// Items is the catalog of all item definitions.
var Items = map[string]ItemDef{
	// ─── Earth Weapons ───────────────────────────────────────
	"combat_knife": {
		ID: "combat_knife", Name: "Combat Knife", Slot: SlotWeapon,
		Description: "Standard issue combat knife. Silent, never breaks.",
		Attack: 8, WType: WeaponMelee, Range: 1, Glyph: '/',
		Rarity: 1, Value: 5,
	},
	"m9": {
		ID: "m9", Name: "M9 Beretta", Slot: SlotWeapon,
		Description: "Standard sidearm. Fast reload.",
		Attack: 10, WType: WeaponRanged, Range: 8, Accuracy: 85,
		AmmoType: "ammo_9mm", ClipSize: 15, ReloadTicks: 2,
		FireModes: []FireMode{FireSingle},
		Glyph: '-', ProjColor: "#FFCC44", ProjSpeed: 3,
		Rarity: 1, Value: 10,
	},
	"p90": {
		ID: "p90", Name: "P-90", Slot: SlotWeapon,
		Description: "Standard SG team primary. Reliable and fast. Burst fire mode available.",
		Attack: 12, WType: WeaponRanged, Range: 12, Accuracy: 85,
		AmmoType: "ammo_57mm", ClipSize: 50, ReloadTicks: 3,
		FireModes: []FireMode{FireSingle, FireBurst},
		Glyph: '-', ProjColor: "#FFCC44", ProjSpeed: 3,
		Rarity: 1, Value: 15,
	},
	"usas12": {
		ID: "usas12", Name: "USAS-12 Shotgun", Slot: SlotWeapon,
		Description: "Automatic shotgun. Devastating at close range.",
		Attack: 20, WType: WeaponRanged, Range: 4, Accuracy: 75,
		AmmoType: "ammo_12ga", ClipSize: 10, ReloadTicks: 4,
		FireModes: []FireMode{FireSingle},
		Glyph: '*', ProjColor: "#FF8844", ProjSpeed: 2,
		Rarity: 2, Value: 40,
	},
	"m249": {
		ID: "m249", Name: "M249 SAW", Slot: SlotWeapon,
		Description: "Light machine gun. Suppressive fire.",
		Attack: 14, WType: WeaponRanged, Range: 10, Accuracy: 70,
		AmmoType: "ammo_556", ClipSize: 100, ReloadTicks: 5,
		FireModes: []FireMode{FireSingle, FireBurst},
		Glyph: '-', ProjColor: "#FFCC44", ProjSpeed: 3,
		Rarity: 2, Value: 50,
	},
	"c4": {
		ID: "c4", Name: "C-4 Charge", Slot: SlotNone,
		Description: "Explosive charge. 5-tick fuse, destroys breakable walls.",
		Attack: 50, WType: WeaponPlaced, Consumable: true,
		Rarity: 2, Value: 30,
	},

	// ─── Goa'uld Weapons ────────────────────────────────────
	"staff_weapon": {
		ID: "staff_weapon", Name: "Staff Weapon", Slot: SlotWeapon,
		Description: "Jaffa energy staff. Powerful but slow. Overcharge mode available.",
		Attack: 18, WType: WeaponRanged, Range: 10, Accuracy: 80,
		AmmoType: "ammo_naquadah", ClipSize: 20, ReloadTicks: 4,
		FireModes: []FireMode{FireSingle, FireOvercharge},
		Glyph: '*', ProjColor: "#FFAA00", ProjSpeed: 2,
		Rarity: 2, Value: 30,
	},
	"zat": {
		ID: "zat", Name: "Zat'nik'tel", Slot: SlotWeapon,
		Description: "Goa'uld stun weapon. One shot stuns, two kills.",
		Attack: 6, WType: WeaponRanged, Range: 6, Accuracy: 90,
		AmmoType: "ammo_zat", ClipSize: 10, ReloadTicks: 2,
		FireModes: []FireMode{FireSingle},
		Glyph: '~', ProjColor: "#44CCFF", ProjSpeed: 2,
		Rarity: 2, Value: 25,
	},
	"hand_device": {
		ID: "hand_device", Name: "Kara Kesh", Slot: SlotWeapon,
		Description: "Goa'uld hand device. Pushback 3 tiles. Requires Naquadah affinity.",
		Attack: 25, WType: WeaponRanged, Range: 3, Accuracy: 95,
		AmmoType: "ammo_naquadah", ClipSize: 5, ReloadTicks: 3,
		Glyph: '☼', ProjColor: "#FFDD00", ProjSpeed: 1,
		Rarity: 4, Value: 100,
	},
	"pain_stick": {
		ID: "pain_stick", Name: "Pain Stick", Slot: SlotWeapon,
		Description: "Goa'uld melee weapon. 1-tick stun on hit.",
		Attack: 15, WType: WeaponMelee, Range: 1,
		Rarity: 2, Value: 20,
	},
	"shock_grenade": {
		ID: "shock_grenade", Name: "Shock Grenade", Slot: SlotNone,
		Description: "AoE 2x2, 3-tick stun. Thrown.",
		Attack: 5, WType: WeaponThrown, Range: 6, Consumable: true,
		Rarity: 2, Value: 15,
	},

	// ─── Ancient Weapons ─────────────────────────────────────
	"ancient_drone": {
		ID: "ancient_drone", Name: "Ancient Drone", Slot: SlotWeapon,
		Description: "Homing drone. Ignores cover. Limited ammo.",
		Attack: 30, WType: WeaponRanged, Range: 15, Accuracy: 95,
		AmmoType: "ammo_drone", ClipSize: 3, ReloadTicks: 5,
		Glyph: '>', ProjColor: "#88CCFF", ProjSpeed: 4,
		Rarity: 4, Value: 200,
	},
	"ancient_pistol": {
		ID: "ancient_pistol", Name: "Ancient Hand Weapon", Slot: SlotWeapon,
		Description: "Energy weapon of the Ancients. Pierces shields.",
		Attack: 22, WType: WeaponRanged, Range: 8, Accuracy: 88,
		AmmoType: "ammo_ancient", ClipSize: 12, ReloadTicks: 3,
		Glyph: '†', ProjColor: "#AADDFF", ProjSpeed: 3,
		Rarity: 4, Value: 150,
	},
	"arg": {
		ID: "arg", Name: "Anti-Replicator Gun", Slot: SlotWeapon,
		Description: "Disintegrates Replicators. Useless against organic targets.",
		Attack: 0, WType: WeaponRanged, Range: 10, Accuracy: 90,
		AmmoType: "ammo_arg", ClipSize: 10, ReloadTicks: 3,
		Glyph: '=', ProjColor: "#FFFFFF", ProjSpeed: 3,
		Rarity: 4, Value: 250,
	},

	// ─── Asgard Weapons ──────────────────────────────────────
	"asgard_beam": {
		ID: "asgard_beam", Name: "Asgard Plasma Beam", Slot: SlotWeapon,
		Description: "The most powerful weapon in the galaxy. Pierces all shields.",
		Attack: 40, WType: WeaponRanged, Range: 20, Accuracy: 95,
		AmmoType: "ammo_asgard", ClipSize: 5, ReloadTicks: 6,
		Glyph: '=', ProjColor: "#CCDDFF", ProjSpeed: 5,
		Rarity: 4, Value: 500,
	},

	// ─── Ori Weapons ─────────────────────────────────────────
	"ori_staff": {
		ID: "ori_staff", Name: "Ori Staff", Slot: SlotWeapon,
		Description: "Ori energy weapon. Burns target for extra damage over time.",
		Attack: 20, WType: WeaponRanged, Range: 12, Accuracy: 85,
		AmmoType: "ammo_ori", ClipSize: 15, ReloadTicks: 4,
		Glyph: '*', ProjColor: "#FF8800", ProjSpeed: 2,
		Rarity: 3, Value: 80,
	},

	// ─── Ammo Types ──────────────────────────────────────────
	"ammo_9mm":     {ID: "ammo_9mm", Name: "9mm Rounds", Consumable: true, Rarity: 1, Value: 2},
	"ammo_57mm":    {ID: "ammo_57mm", Name: "5.7mm Rounds", Consumable: true, Rarity: 1, Value: 2},
	"ammo_12ga":    {ID: "ammo_12ga", Name: "12ga Shells", Consumable: true, Rarity: 1, Value: 3},
	"ammo_556":     {ID: "ammo_556", Name: "5.56mm Rounds", Consumable: true, Rarity: 1, Value: 2},
	"ammo_naquadah": {ID: "ammo_naquadah", Name: "Naquadah Charge", Consumable: true, Rarity: 2, Value: 5},
	"ammo_zat":     {ID: "ammo_zat", Name: "Zat Charge", Consumable: true, Rarity: 2, Value: 5},
	"ammo_drone":   {ID: "ammo_drone", Name: "Drone Charge", Consumable: true, Rarity: 4, Value: 30},
	"ammo_ancient": {ID: "ammo_ancient", Name: "Ancient Cell", Consumable: true, Rarity: 3, Value: 15},
	"ammo_arg":     {ID: "ammo_arg", Name: "ARG Charge", Consumable: true, Rarity: 4, Value: 25},
	"ammo_asgard":  {ID: "ammo_asgard", Name: "Asgard Plasma Cell", Consumable: true, Rarity: 4, Value: 50},
	"ammo_ori":     {ID: "ammo_ori", Name: "Ori Power Cell", Consumable: true, Rarity: 3, Value: 10},

	// ─── Armor ───────────────────────────────────────────────
	"bdu": {
		ID: "bdu", Name: "Battle Dress Uniform", Slot: SlotArmor,
		Description: "Standard SGC fatigues.",
		Defense: 2, Weight: 0, ArmorSlot: "torso",
		Rarity: 1, Value: 5,
	},
	"tac_vest": {
		ID: "tac_vest", Name: "SGC Tac Vest", Slot: SlotArmor,
		Description: "Standard SGC tactical vest with ceramic plates. +6 inventory slots.",
		Defense: 5, Weight: 0, ArmorSlot: "torso",
		Rarity: 1, Value: 10,
	},
	"kevlar_helmet": {
		ID: "kevlar_helmet", Name: "Kevlar Helmet", Slot: SlotArmor,
		Description: "Standard protective helmet.",
		Defense: 3, Weight: 0, ArmorSlot: "head",
		Rarity: 1, Value: 8,
	},
	"jaffa_armor": {
		ID: "jaffa_armor", Name: "Jaffa Chainmail", Slot: SlotArmor,
		Description: "Heavy Jaffa combat armor. Durable.",
		Defense: 8, Weight: 1, ArmorSlot: "torso",
		Rarity: 2, Value: 30,
	},
	"serpent_helm": {
		ID: "serpent_helm", Name: "Serpent Guard Helm", Slot: SlotArmor,
		Description: "Serpent Guard helmet. Reduces FOV by 1.",
		Defense: 6, Weight: 1, ArmorSlot: "head",
		Rarity: 2, Value: 25,
	},
	"kull_armor": {
		ID: "kull_armor", Name: "Kull Warrior Armor", Slot: SlotArmor,
		Description: "Full Kull armor. Immune to zat stun. Heavy.",
		Defense: 15, Weight: 2, ArmorSlot: "full",
		Rarity: 3, Value: 80,
	},
	"personal_shield": {
		ID: "personal_shield", Name: "Personal Shield", Slot: SlotArmor,
		Description: "Ancient personal shield emitter. Blocks ranged, vulnerable to melee.",
		Defense: 20, Weight: 0, ArmorSlot: "torso",
		Rarity: 4, Value: 120,
	},
	"ori_armor": {
		ID: "ori_armor", Name: "Ori Armor", Slot: SlotArmor,
		Description: "Ori battle armor. Fire resistance.",
		Defense: 12, Weight: 1, ArmorSlot: "full",
		Rarity: 3, Value: 60,
	},

	// ─── Accessories ─────────────────────────────────────────
	"gdo": {
		ID: "gdo", Name: "GDO", Slot: SlotAccessory,
		Description: "Garage Door Opener — sends your IDC through the gate.",
		Rarity: 1, Value: 5,
	},
	"healing_device": {
		ID: "healing_device", Name: "Healing Device", Slot: SlotAccessory,
		Description: "Goa'uld healing device. Slowly regenerates HP.",
		HealAmount: 1, Rarity: 3, Value: 80,
	},

	// ─── Consumables ─────────────────────────────────────────
	"bandage": {
		ID: "bandage", Name: "Field Bandage", Slot: SlotNone,
		Description: "Basic first aid. Restores 15 HP over 5 ticks.",
		HealAmount: 15, Consumable: true, Rarity: 1, Value: 5,
	},
	"morphine": {
		ID: "morphine", Name: "Morphine Syrette", Slot: SlotNone,
		Description: "Instant heal 30 HP, -20% accuracy for 10 ticks.",
		HealAmount: 30, Consumable: true, Rarity: 2, Value: 15,
	},
	"tretonin": {
		ID: "tretonin", Name: "Tretonin", Slot: SlotNone,
		Description: "Tok'ra-developed compound. Restores 50 HP over 10 ticks.",
		HealAmount: 50, Consumable: true, Rarity: 3, Value: 30,
	},
	"naquadah_grenade": {
		ID: "naquadah_grenade", Name: "Naquadah Grenade", Slot: SlotNone,
		Description: "Explosive device. Deals massive area damage.",
		Attack: 50, WType: WeaponThrown, Range: 6, Consumable: true,
		Rarity: 3, Value: 50,
	},
	"mre": {
		ID: "mre", Name: "MRE", Slot: SlotNone,
		Description: "Meal Ready to Eat. Restores 10 HP.",
		HealAmount: 10, Consumable: true, Rarity: 1, Value: 2,
	},
	"malp_beacon": {
		ID: "malp_beacon", Name: "MALP Beacon", Slot: SlotNone,
		Description: "Reveals 10-tile radius through fog of war.",
		Consumable: true, Rarity: 2, Value: 15,
	},
	"tunnel_crystal": {
		ID: "tunnel_crystal", Name: "Tok'ra Tunnel Crystal", Slot: SlotNone,
		Description: "Creates a 3-tile tunnel through rock walls.",
		Consumable: true, Rarity: 3, Value: 25,
	},
	"naquadah_reactor_cell": {
		ID: "naquadah_reactor_cell", Name: "Naquadah Reactor Cell", Slot: SlotNone,
		Description: "Recharges all energy weapons to full.",
		Consumable: true, Rarity: 2, Value: 20,
	},

	// ─── Key Items / Materials ───────────────────────────────
	"naquadah_sample": {
		ID: "naquadah_sample", Name: "Naquadah Sample", Slot: SlotNone,
		Description: "Raw naquadah. Used for crafting and research.",
		Rarity: 2, Value: 10,
	},
	"ancient_data_pad": {
		ID: "ancient_data_pad", Name: "Ancient Data Pad", Slot: SlotNone,
		Description: "Encrypted research data. Can be decoded at the SGC lab.",
		Rarity: 3, Value: 25,
	},
	"kull_fragment": {
		ID: "kull_fragment", Name: "Kull Armor Fragment", Slot: SlotNone,
		Description: "Salvaged Kull Warrior armor piece.",
		Rarity: 3, Value: 20,
	},
	"replicator_shard": {
		ID: "replicator_shard", Name: "Replicator Shard", Slot: SlotNone,
		Description: "Fragment of a destroyed Replicator.",
		Rarity: 3, Value: 15,
	},
	"asgard_component": {
		ID: "asgard_component", Name: "Asgard Component", Slot: SlotNone,
		Description: "Advanced Asgard technology component.",
		Rarity: 4, Value: 50,
	},
}
