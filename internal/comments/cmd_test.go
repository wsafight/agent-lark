package comments

import "testing"

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Name() != "comments" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "comments")
	}

	want := map[string]bool{
		"list": false, "add": false, "reply": false, "resolve": false,
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
