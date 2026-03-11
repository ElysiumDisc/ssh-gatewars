package gamedata

// LootEntry is a possible item drop with a weight.
type LootEntry struct {
	ItemID string
	Weight int // relative probability
}

// LootTables maps loot table IDs to their possible drops.
var LootTables = map[string][]LootEntry{
	// ─── Goa'uld ────────────────────────────────────────────
	"jaffa_common": {
		{"bandage", 30},
		{"ammo_naquadah", 25},
		{"staff_weapon", 8},
		{"jaffa_armor", 5},
		{"pain_stick", 5},
		{"shock_grenade", 5},
		{"naquadah_sample", 10},
	},
	"serpent_guard": {
		{"bandage", 20},
		{"ammo_naquadah", 20},
		{"staff_weapon", 15},
		{"jaffa_armor", 12},
		{"zat", 8},
		{"ammo_zat", 10},
		{"shock_grenade", 8},
		{"naquadah_sample", 15},
		{"serpent_helm", 5},
	},
	"kull_warrior": {
		{"tretonin", 15},
		{"jaffa_armor", 10},
		{"staff_weapon", 10},
		{"kull_fragment", 20},
		{"naquadah_sample", 15},
		{"ammo_naquadah", 15},
		{"naquadah_grenade", 8},
	},
	"boss_goauld": {
		{"hand_device", 15},
		{"personal_shield", 10},
		{"naquadah_sample", 20},
		{"ancient_data_pad", 10},
		{"naquadah_grenade", 15},
		{"tretonin", 10},
		{"kull_fragment", 10},
	},

	// ─── Replicators ────────────────────────────────────────
	"replicator": {
		{"replicator_shard", 40},
		{"naquadah_sample", 20},
		{"ancient_data_pad", 10},
	},

	// ─── Ori ────────────────────────────────────────────────
	"ori_common": {
		{"bandage", 20},
		{"ammo_ori", 25},
		{"ori_staff", 10},
		{"naquadah_sample", 15},
	},
	"ori_rare": {
		{"ori_staff", 15},
		{"ori_armor", 10},
		{"ammo_ori", 20},
		{"naquadah_sample", 15},
		{"ancient_data_pad", 10},
		{"tretonin", 10},
	},

	// ─── Wildlife ───────────────────────────────────────────
	"wildlife": {
		{"bandage", 40},
		{"mre", 30},
		{"naquadah_sample", 5},
	},

	// ─── Crates ─────────────────────────────────────────────
	"crate_common": {
		{"bandage", 25},
		{"ammo_57mm", 20},
		{"ammo_9mm", 15},
		{"mre", 15},
		{"malp_beacon", 5},
		{"tac_vest", 5},
	},
	"crate_rare": {
		{"tretonin", 15},
		{"ammo_naquadah", 15},
		{"staff_weapon", 10},
		{"jaffa_armor", 10},
		{"naquadah_grenade", 8},
		{"naquadah_sample", 12},
		{"healing_device", 5},
		{"ancient_data_pad", 5},
		{"tunnel_crystal", 5},
	},
	"crate_earth": {
		{"ammo_57mm", 25},
		{"ammo_9mm", 20},
		{"ammo_12ga", 10},
		{"ammo_556", 10},
		{"bandage", 15},
		{"morphine", 8},
		{"c4", 5},
		{"malp_beacon", 5},
	},
}
