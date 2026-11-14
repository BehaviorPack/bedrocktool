package worldstate

import (
	"reflect"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestAddedStatesRegistry(t *testing.T) {
	reg := world.NewBlockRegistry()
	for _, lit := range []bool{false, true} {
		reg.RegisterBlockState(world.BlockState{Name: "test:lamp", Properties: map[string]any{"test:lit": lit}})
	}
	reg.Finalize()
	lit, ok := reg.StateToRuntimeID("test:lamp", map[string]any{"test:lit": true})
	if !ok {
		t.Fatal("test:lamp not registered")
	}

	added := map[string]map[string]any{"test:lamp": {"minecraft:cardinal_direction": "north"}}
	saved := withAddedStates(reg, added)

	name, props, ok := saved.RuntimeIDToState(lit)
	want := map[string]any{"test:lit": true, "minecraft:cardinal_direction": "north"}
	if !ok || name != "test:lamp" || !reflect.DeepEqual(props, want) {
		t.Errorf("saved as %s %v, want test:lamp %v", name, props, want)
	}
	// The server's registry is left as it was.
	if _, props, _ := reg.RuntimeIDToState(lit); len(props) != 1 {
		t.Errorf("server registry changed: %v", props)
	}

	// A saved block reads back as the server's block, whichever way it faces.
	for _, dir := range []string{"north", "east"} {
		rid, ok := saved.StateToRuntimeID("test:lamp", map[string]any{"test:lit": true, "minecraft:cardinal_direction": dir})
		if !ok || rid != lit {
			t.Errorf("%s: read back as %d %v, want %d", dir, rid, ok, lit)
		}
	}

	// Vanilla blocks are untouched.
	stone, _ := saved.StateToRuntimeID("minecraft:stone", map[string]any{})
	if _, props, _ := saved.RuntimeIDToState(stone); len(props) != 0 {
		t.Errorf("stone saved with %v", props)
	}

	if withAddedStates(reg, nil) != reg {
		t.Error("registry wrapped with nothing to add")
	}
}
