package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/output"
)

func newTestRootCommand() (*cobra.Command, *cobra.Command) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("format", "text", "")
	root.PersistentFlags().String("token-mode", "auto", "")
	root.PersistentFlags().String("profile", "", "")
	root.PersistentFlags().String("config", "", "")
	root.PersistentFlags().String("domain", "", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.PersistentFlags().Bool("quiet", false, "")
	root.PersistentFlags().Bool("agent", false, "")

	sub := &cobra.Command{Use: "sub"}
	root.AddCommand(sub)
	return root, sub
}

func TestResolveGlobalFlagsAgentMode(t *testing.T) {
	prev := output.GlobalAgent
	t.Cleanup(func() { output.GlobalAgent = prev })
	output.GlobalAgent = false

	root, sub := newTestRootCommand()
	_ = root.PersistentFlags().Set("format", "md")
	_ = root.PersistentFlags().Set("token-mode", "user")
	_ = root.PersistentFlags().Set("agent", "true")

	g := ResolveGlobalFlags(sub)
	if !g.Agent {
		t.Fatalf("Agent: got false, want true")
	}
	if g.Format != "json" {
		t.Fatalf("Format: got %q, want %q", g.Format, "json")
	}
	if g.TokenMode != "user" {
		t.Fatalf("TokenMode: got %q, want %q", g.TokenMode, "user")
	}
	if !output.GlobalAgent {
		t.Fatalf("output.GlobalAgent: got false, want true")
	}
}

func TestResolveGlobalFlagsNonAgentMode(t *testing.T) {
	prev := output.GlobalAgent
	t.Cleanup(func() { output.GlobalAgent = prev })
	output.GlobalAgent = false

	root, sub := newTestRootCommand()
	_ = root.PersistentFlags().Set("format", "md")
	_ = root.PersistentFlags().Set("agent", "false")

	g := ResolveGlobalFlags(sub)
	if g.Agent {
		t.Fatalf("Agent: got true, want false")
	}
	if g.Format != "md" {
		t.Fatalf("Format: got %q, want %q", g.Format, "md")
	}
	if output.GlobalAgent {
		t.Fatalf("output.GlobalAgent: got true, want false")
	}
}
