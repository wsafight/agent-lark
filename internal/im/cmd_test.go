package im

import "testing"

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Name() != "im" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "im")
	}

	// verify top-level subcommands
	want := map[string]bool{
		"chats": false, "messages": false, "react": false, "send": false,
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

func TestSendCommandFlags(t *testing.T) {
	cmd := newSendCommand()

	flags := []string{"chat-id", "user-id", "text", "card-file", "card"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag %q not defined on send command", name)
		}
	}
}

func TestChatsSubcommands(t *testing.T) {
	cmd := NewCommand()
	chats, _, err := cmd.Find([]string{"chats"})
	if err != nil || chats.Name() != "chats" {
		t.Fatal("chats subcommand not found")
	}

	want := map[string]bool{"list": false, "search": false}
	for _, sub := range chats.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("chats subcommand %q not registered", name)
		}
	}
}

func TestMessagesSubcommands(t *testing.T) {
	cmd := NewCommand()
	messages, _, err := cmd.Find([]string{"messages"})
	if err != nil || messages.Name() != "messages" {
		t.Fatal("messages subcommand not found")
	}

	want := map[string]bool{"list": false, "get": false}
	for _, sub := range messages.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("messages subcommand %q not registered", name)
		}
	}
}

func TestReactSubcommands(t *testing.T) {
	cmd := NewCommand()
	react, _, err := cmd.Find([]string{"react"})
	if err != nil || react.Name() != "react" {
		t.Fatal("react subcommand not found")
	}

	want := map[string]bool{"add": false, "remove": false}
	for _, sub := range react.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("react subcommand %q not registered", name)
		}
	}
}
