package behaviourpack

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// blockJSON adds the block to a pack and returns its generated JSON, decoded.
func blockJSON(t *testing.T, bp *Pack, entry protocol.BlockEntry) map[string]any {
	t.Helper()
	bp.AddBlock(entry)
	data, err := json.Marshal(bp.blocks[entry.Name])
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out["minecraft:block"].(map[string]any)
}

func rotationPerm(dir string, y float64) map[string]any {
	return map[string]any{
		"condition": "q.block_state('minecraft:cardinal_direction') == '" + dir + "'",
		"components": map[string]any{
			"minecraft:transformation": map[string]any{"rotation": []any{0.0, y, 0.0}},
		},
	}
}

func TestCardinalRotationAdded(t *testing.T) {
	bp := New("test")
	block := blockJSON(t, bp, protocol.BlockEntry{
		Name: "test:plain",
		Properties: map[string]any{
			"components": map[string]any{
				"minecraft:geometry": map[string]any{"identifier": "geometry.plain"},
			},
		},
	})

	traits := block["description"].(map[string]any)["traits"]
	wantTraits := map[string]any{
		"minecraft:placement_direction": map[string]any{
			"enabled_states":    []any{"minecraft:cardinal_direction"},
			"y_rotation_offset": 180.0,
		},
	}
	if !reflect.DeepEqual(traits, wantTraits) {
		t.Errorf("traits = %v, want %v", traits, wantTraits)
	}

	wantPerms := []any{
		rotationPerm("north", 0),
		rotationPerm("west", 90),
		rotationPerm("south", 180),
		rotationPerm("east", -90),
	}
	if !reflect.DeepEqual(block["permutations"], wantPerms) {
		t.Errorf("permutations = %v, want %v", block["permutations"], wantPerms)
	}

	want := map[string]map[string]any{"test:plain": {"minecraft:cardinal_direction": "north"}}
	if got := bp.AddedBlockStates(); !reflect.DeepEqual(got, want) {
		t.Errorf("AddedBlockStates = %v, want %v", got, want)
	}
}

func TestCardinalRotationMergesPermutations(t *testing.T) {
	bp := New("test")
	block := blockJSON(t, bp, protocol.BlockEntry{
		Name: "test:lamp",
		Properties: map[string]any{
			"properties": []any{map[string]any{"name": "test:lit", "enum": []uint8{0, 1}}},
			"permutations": []any{map[string]any{
				"condition": "q.block_state('test:lit') == true",
				"components": map[string]any{
					"minecraft:light_emission": map[string]any{"emission": uint8(15)},
				},
			}},
			"components": map[string]any{
				"minecraft:geometry": map[string]any{"identifier": "geometry.lamp"},
				"minecraft:transformation": map[string]any{
					"RY": int32(1), "SX": float32(1), "SY": float32(1), "SZ": float32(1), "TY": float32(0.5),
				},
			},
		},
	})

	perms := block["permutations"].([]any)
	if len(perms) != 5 {
		t.Fatalf("%d permutations, want the block's own one and 4 rotations: %v", len(perms), perms)
	}
	if c := perms[0].(map[string]any)["condition"]; c != "q.block_state('test:lit') == true" {
		t.Errorf("block's own permutation not kept first: %v", c)
	}
	// The block's own transformation is kept, its rotation turned further.
	east := perms[4].(map[string]any)["components"].(map[string]any)["minecraft:transformation"]
	wantEast := map[string]any{
		"rotation":    []any{0.0, 0.0, 0.0},
		"scale":       []any{1.0, 1.0, 1.0},
		"translation": []any{0.0, 0.5, 0.0},
	}
	if !reflect.DeepEqual(east, wantEast) {
		t.Errorf("east transformation = %v, want %v", east, wantEast)
	}
	if _, ok := block["description"].(map[string]any)["states"].(map[string]any)["test:lit"]; !ok {
		t.Error("block's own state lost")
	}
}

func TestCardinalRotationSkipped(t *testing.T) {
	for _, c := range []struct {
		name  string
		props map[string]any
	}{
		{"no geometry", map[string]any{
			"components": map[string]any{
				"minecraft:destructible_by_mining": map[string]any{"value": float32(1)},
			},
		}},
		{"placement_direction", map[string]any{
			"traits": []any{map[string]any{
				"name":           "placement_direction",
				"enabled_states": map[string]any{"facing_direction": uint8(1), "cardinal_direction": uint8(0)},
			}},
		}},
		{"placement_position", map[string]any{
			"traits": []any{map[string]any{
				"name":           "placement_position",
				"enabled_states": map[string]any{"block_face": uint8(1)},
			}},
		}},
		{"cardinal permutations", map[string]any{
			"permutations": []any{map[string]any{
				"condition":  "q.block_state('minecraft:cardinal_direction') == 'east'",
				"components": map[string]any{},
			}},
		}},
		{"facing permutations", map[string]any{
			"permutations": []any{map[string]any{
				"condition":  "query.block_property('minecraft:facing_direction') == 2",
				"components": map[string]any{},
			}},
		}},
		{"transforming permutations", map[string]any{
			"permutations": []any{map[string]any{
				"condition": "q.block_state('test:open') == true",
				"components": map[string]any{
					"minecraft:transformation": map[string]any{"RY": int32(2)},
				},
			}},
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if c.name != "no geometry" {
				components, _ := c.props["components"].(map[string]any)
				if components == nil {
					components = map[string]any{}
					c.props["components"] = components
				}
				components["minecraft:geometry"] = map[string]any{"identifier": "geometry.own"}
			}
			bp := New("test")
			block := blockJSON(t, bp, protocol.BlockEntry{Name: "test:own", Properties: c.props})

			perms, _ := block["permutations"].([]any)
			if own, ok := c.props["permutations"].([]any); ok && len(perms) != len(own) {
				want := len(own)
				t.Errorf("%d permutations, want the block's %d", len(perms), want)
			}
			if traits, ok := block["description"].(map[string]any)["traits"].(map[string]any); ok {
				if p, ok := traits["minecraft:placement_direction"].(map[string]any); ok && c.name != "placement_direction" {
					t.Errorf("placement_direction added: %v", p)
				}
			}
			if len(bp.AddedBlockStates()) != 0 {
				t.Errorf("states added: %v", bp.AddedBlockStates())
			}
		})
	}
}
