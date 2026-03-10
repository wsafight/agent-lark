package doctor

import "testing"

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Use != "doctor" {
		t.Errorf("Use = %q, want %q", cmd.Use, "doctor")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}
