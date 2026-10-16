package causes_test

import (
	"testing"

	"iacon/internal/causes"
)

func TestAllCausesAreValid(t *testing.T) {
	for _, cause := range causes.All() {
		if !causes.Valid(cause) {
			t.Errorf("cause %q is not valid", cause)
		}
	}
	if causes.Valid("redis down") {
		t.Fatal("human-readable cause must not be a valid slug")
	}
}
