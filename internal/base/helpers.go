package base

import "github.com/spf13/cobra"

// PagedResponse is used for agent mode paged responses.
type PagedResponse struct {
	Items      any    `json:"items"`
	NextCursor string `json:"next_cursor"`
}

func getGlobalFlags(cmd *cobra.Command) (format, tokenMode, profile, config, domain string, debug, quiet, agent bool) {
	root := cmd.Root()
	format, _ = root.PersistentFlags().GetString("format")
	tokenMode, _ = root.PersistentFlags().GetString("token-mode")
	profile, _ = root.PersistentFlags().GetString("profile")
	config, _ = root.PersistentFlags().GetString("config")
	domain, _ = root.PersistentFlags().GetString("domain")
	debug, _ = root.PersistentFlags().GetBool("debug")
	quiet, _ = root.PersistentFlags().GetBool("quiet")
	agent, _ = root.PersistentFlags().GetBool("agent")
	return
}
