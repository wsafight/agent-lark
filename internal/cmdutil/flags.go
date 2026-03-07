package cmdutil

import "github.com/spf13/cobra"

// GlobalFlags holds the parsed root-level persistent flags.
type GlobalFlags struct {
	Format    string
	TokenMode string
	Profile   string
	Config    string
	Domain    string
	Debug     bool
	Quiet     bool
	Agent     bool
}

// GetGlobalFlags reads the root command's persistent flags from any subcommand.
func GetGlobalFlags(cmd *cobra.Command) GlobalFlags {
	root := cmd.Root()
	f := root.PersistentFlags()
	var g GlobalFlags
	g.Format, _ = f.GetString("format")
	g.TokenMode, _ = f.GetString("token-mode")
	g.Profile, _ = f.GetString("profile")
	g.Config, _ = f.GetString("config")
	g.Domain, _ = f.GetString("domain")
	g.Debug, _ = f.GetBool("debug")
	g.Quiet, _ = f.GetBool("quiet")
	g.Agent, _ = f.GetBool("agent")
	return g
}
