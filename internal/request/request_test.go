package request

import "testing"

func TestBuildPromptProducesCurrentCoordinatorRequest(t *testing.T) {
	body, err := BuildPrompt("hello coordinator", 60)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(body); err != nil {
		t.Fatalf("generated request did not validate: %v", err)
	}
}

func TestValidateRejectsUnsupportedHarness(t *testing.T) {
	body := []byte(`{"schemaVersion":"0.1.0","executionId":"e","correlationId":"c","harness":"other","task":"x","timeoutSeconds":1}`)
	if err := Validate(body); err == nil {
		t.Fatal("expected unsupported harness to fail validation")
	}
}

func TestValidateRejectsTrailingJSON(t *testing.T) {
	body := []byte(`{"schemaVersion":"0.1.0","executionId":"e","correlationId":"c","harness":"pi.dev","task":"x","timeoutSeconds":1}{}`)
	if err := Validate(body); err == nil {
		t.Fatal("expected trailing JSON to fail validation")
	}
}
