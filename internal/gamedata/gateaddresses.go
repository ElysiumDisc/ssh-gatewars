package gamedata

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
)

// GlyphCount is the number of distinct glyphs in the gate alphabet.
const GlyphCount = 39

// GateAddress is a 7-symbol stargate address.
type GateAddress [7]int

// Glyphs are the 39 constellation symbols displayed on the DHD.
// Using Unicode geometric shapes as stand-ins for the actual constellations.
var Glyphs = [GlyphCount]rune{
	'◆', '◇', '△', '▽', '○', '□', '☆', '◈', '▷', '◁',
	'⬡', '⬢', '◉', '◎', '▣', '▤', '▥', '▦', '▧', '▨',
	'◐', '◑', '◒', '◓', '⊕', '⊗', '⊙', '⊛', '⊞', '⊟',
	'⊠', '⊡', '⌂', '⌘', '⍟', '⏣', '⏥', '⏦', '⎈',
}

// ParseAddress converts a dash-separated numeric string to a GateAddress.
// E.g. "26-6-14-31-11-29-0" → GateAddress{26,6,14,31,11,29,0}
func ParseAddress(s string) (GateAddress, bool) {
	var addr GateAddress
	parts := strings.Split(s, "-")
	if len(parts) != 7 {
		return addr, false
	}
	for i, p := range parts {
		n := 0
		for _, c := range p {
			if c < '0' || c > '9' {
				return addr, false
			}
			n = n*10 + int(c-'0')
		}
		if n < 0 || n >= GlyphCount {
			return addr, false
		}
		addr[i] = n
	}
	return addr, true
}

// String returns the display form of a gate address using glyph symbols.
func (a GateAddress) String() string {
	parts := make([]string, 7)
	for i, g := range a {
		if g >= 0 && g < GlyphCount {
			parts[i] = string(Glyphs[g])
		} else {
			parts[i] = "?"
		}
	}
	return strings.Join(parts, "-")
}

// Code returns the numeric form (e.g. "26-6-14-31-11-29-0").
func (a GateAddress) Code() string {
	return fmt.Sprintf("%d-%d-%d-%d-%d-%d-%d", a[0], a[1], a[2], a[3], a[4], a[5], a[6])
}

// Seed computes a deterministic int64 seed from a gate address.
func (a GateAddress) Seed() int64 {
	h := sha256.New()
	for _, g := range a {
		binary.Write(h, binary.LittleEndian, int32(g))
	}
	sum := h.Sum(nil)
	return int64(binary.LittleEndian.Uint64(sum[:8]))
}

// IsValid checks that all glyphs are in range and no duplicates.
func (a GateAddress) IsValid() bool {
	seen := make(map[int]bool)
	for _, g := range a {
		if g < 0 || g >= GlyphCount {
			return false
		}
		if seen[g] {
			return false
		}
		seen[g] = true
	}
	return true
}

// NamedAddress is a well-known gate address with a planet name.
type NamedAddress struct {
	Name    string
	Address GateAddress
}

// NamedAddresses are the canonical SG-1 planet addresses.
var NamedAddresses = []NamedAddress{
	{"Earth", GateAddress{26, 6, 14, 31, 11, 29, 0}},
	{"Abydos", GateAddress{1, 3, 9, 15, 22, 36, 2}},
	{"Chulak", GateAddress{4, 12, 18, 25, 33, 7, 5}},
	{"Tollana", GateAddress{8, 20, 27, 35, 16, 10, 13}},
	{"Cimmeria", GateAddress{17, 23, 30, 38, 19, 34, 21}},
	{"Dakara", GateAddress{28, 32, 37, 24, 2, 8, 14}},
	{"Langara", GateAddress{11, 35, 22, 6, 17, 30, 9}},
	{"Atlantis", GateAddress{33, 19, 5, 27, 13, 38, 24}},
}

// EarthAddress is the address for SGC/Earth.
var EarthAddress = NamedAddresses[0].Address

// NameForAddress returns the planet name if the address is named, or "".
func NameForAddress(a GateAddress) string {
	for _, na := range NamedAddresses {
		if na.Address == a {
			return na.Name
		}
	}
	return ""
}

// PDesignation generates a P-designation name from an address seed.
// Format: P<digit><letter><digit>-<3 digits>
func PDesignation(a GateAddress) string {
	seed := a.Seed()
	if seed < 0 {
		seed = -seed
	}
	d1 := seed % 10
	letter := rune('A' + (seed/10)%26)
	d2 := (seed / 260) % 10
	num := (seed / 2600) % 1000
	return fmt.Sprintf("P%d%c%d-%03d", d1, letter, d2, num)
}

// PlanetName returns the canonical name or a P-designation.
func PlanetName(a GateAddress) string {
	if name := NameForAddress(a); name != "" {
		return name
	}
	return PDesignation(a)
}
