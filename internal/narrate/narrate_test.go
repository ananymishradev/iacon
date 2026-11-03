package narrate

import (
	"testing"

	"iacon/internal/evidence"
)

func TestValidateAndRedact(t *testing.T) {
	bundle := evidence.Evidence{Items: []evidence.Item{{ID: "E1", Data: map[string]any{
		"ready_body": "postgres://workshop:workshop@example/db",
	}}}}
	out := Output{Summary: "summary", Cause: "unknown", Confidence: "low", CitedEvidence: []string{"E1"}}
	if err := Validate(out, bundle); err != nil {
		t.Fatal(err)
	}
	prompt := Redact(bundle)
	if len(prompt.Data) == 0 || contains(prompt.Data, "workshop:workshop") {
		t.Fatalf("redaction failed: %s", prompt.Data)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
