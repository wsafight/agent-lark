package contact

import "testing"

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Name() != "contact" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "contact")
	}

	want := map[string]bool{
		"search": false, "list": false, "resolve": false,
	}
	for _, sub := range cmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}
