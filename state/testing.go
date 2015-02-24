package state

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

// TestState is a helper for testing state implementations. It is expected
// that the given implementation is pre-loaded with the TestStateInitial
// state.
func TestState(t *testing.T, s interface{}) {
	reader, ok := s.(StateReader)
	if !ok {
		t.Fatalf("must at least be a StateReader")
	}

	// If it implements refresh, refresh
	if rs, ok := s.(StateRefresher); ok {
		if err := rs.RefreshState(); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// current will track our current state
	current := TestStateInitial()

	// Check that the initial state is correct
	state := reader.State()
	current.Serial = state.Serial
	if !reflect.DeepEqual(state, current) {
		t.Fatalf("not initial: %#v\n\n%#v", state, current)
	}

	// Write a new state and verify that we have it
	if ws, ok := s.(StateWriter); ok {
		current.Modules = append(current.Modules, &terraform.ModuleState{
			Path: []string{"root"},
			Outputs: map[string]string{
				"bar": "baz",
			},
		})

		if err := ws.WriteState(current); err != nil {
			t.Fatalf("err: %s", err)
		}

		if actual := reader.State(); !reflect.DeepEqual(actual, current) {
			t.Fatalf("bad: %#v\n\n%#v", actual, current)
		}
	}

	// Test persistence
	if ps, ok := s.(StatePersister); ok {
		if err := ps.PersistState(); err != nil {
			t.Fatalf("err: %s", err)
		}

		// Refresh if we got it
		if rs, ok := s.(StateRefresher); ok {
			if err := rs.RefreshState(); err != nil {
				t.Fatalf("err: %s", err)
			}
		}

		// Just set the serials the same... Then compare.
		actual := reader.State()
		actual.Serial = current.Serial
		if !reflect.DeepEqual(actual, current) {
			t.Fatalf("bad: %#v\n\n%#v", actual, current)
		}
	}
}

// TestStateInitial is the initial state that a State should have
// for TestState.
func TestStateInitial() *terraform.State {
	initial := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root", "child"},
				Outputs: map[string]string{
					"foo": "bar",
				},
			},
		},
	}

	var scratch bytes.Buffer
	terraform.WriteState(initial, &scratch)
	return initial
}