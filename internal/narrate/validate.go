package narrate

import (
	"fmt"

	"iacon/internal/causes"
	"iacon/internal/evidence"
)

func Validate(out Output, bundle evidence.Evidence) error {
	if out.Summary == "" || len(out.Summary) > 4000 {
		return fmt.Errorf("summary must be non-empty and <= 4000 characters")
	}
	if !causes.Valid(out.Cause) {
		return fmt.Errorf("invalid cause %q", out.Cause)
	}
	if out.Confidence != "high" && out.Confidence != "medium" && out.Confidence != "low" {
		return fmt.Errorf("invalid confidence %q", out.Confidence)
	}
	known := map[string]bool{}
	for _, item := range bundle.Items {
		known[item.ID] = true
	}
	if len(out.CitedEvidence) > 32 {
		return fmt.Errorf("too many cited evidence IDs")
	}
	for _, id := range out.CitedEvidence {
		if !known[id] {
			return fmt.Errorf("cited evidence %q does not exist", id)
		}
	}
	return nil
}
