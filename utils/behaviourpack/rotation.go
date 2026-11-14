package behaviourpack

import (
	"fmt"
	"maps"
	"strings"
)

const cardinalDirection = "minecraft:cardinal_direction"

// cardinalRotations is the Y rotation of each cardinal direction, as in
// https://wiki.bedrock.dev/blocks/custom-heads#applying-block-rotation.
var cardinalRotations = []struct {
	direction string
	y         float32
}{
	{"north", 0},
	{"west", 90},
	{"south", 180},
	{"east", -90},
}

// cardinalDefault is the value a saved block takes for a cardinal_direction
// state the server never sent: the trait's default.
const cardinalDefault = "north"

// cardinalYOffset turns the placement direction around, so a placed block's
// front faces the player while north (what saved blocks get) stays
// unrotated, as the server showed them.
const cardinalYOffset = 180

// addCardinalRotation lets a block be placed facing north, east, south or
// west: it enables minecraft:cardinal_direction with the placement_direction
// trait and adds a permutation rotating the block for each direction. It
// leaves alone a block without minecraft:geometry (a cube textured by
// blocks.json, which a transformation doesn't turn), one that already
// rotates (a placement trait, or permutations on cardinal_direction or
// facing_direction), one whose permutations set their own transformation,
// and one whose format is too old for traits. It says whether it added the
// rotation.
func addCardinalRotation(b *MinecraftBlock, version string) bool {
	if version < "1.20.20" || !hasGeometry(b) {
		return false
	}
	for name := range b.Description.Traits {
		if name == "minecraft:placement_direction" || name == "minecraft:placement_position" {
			return false
		}
	}
	if _, ok := b.Description.States[cardinalDirection]; ok {
		return false
	}
	for _, p := range b.Permutations {
		if strings.Contains(p.Condition, "cardinal_direction") || strings.Contains(p.Condition, "facing_direction") {
			return false
		}
		// Later permutations replace a component, so ours would undo the
		// block's own transformation.
		if _, ok := p.Components["minecraft:transformation"]; ok {
			return false
		}
	}

	if b.Description.Traits == nil {
		b.Description.Traits = make(map[string]Trait)
	}
	b.Description.Traits["minecraft:placement_direction"] = Trait{
		"enabled_states":    []string{cardinalDirection},
		"y_rotation_offset": cardinalYOffset,
	}

	base, _ := b.Components["minecraft:transformation"].(map[string][]float32)
	for _, r := range cardinalRotations {
		transform := map[string][]float32{"rotation": {0, r.y, 0}}
		if base != nil {
			// Keep the block's own translation, scale and rotation.
			transform = maps.Clone(base)
			rotation := []float32{0, 0, 0}
			copy(rotation, base["rotation"])
			rotation[1] += r.y
			transform["rotation"] = rotation
		}
		b.Permutations = append(b.Permutations, permutation{
			Condition:  fmt.Sprintf("q.block_state('%s') == '%s'", cardinalDirection, r.direction),
			Components: map[string]any{"minecraft:transformation": transform},
		})
	}
	return true
}

// hasGeometry says whether the block, or any of its permutations, is drawn
// with a model of its own.
func hasGeometry(b *MinecraftBlock) bool {
	if _, ok := b.Components["minecraft:geometry"]; ok {
		return true
	}
	for _, p := range b.Permutations {
		if _, ok := p.Components["minecraft:geometry"]; ok {
			return true
		}
	}
	return false
}
