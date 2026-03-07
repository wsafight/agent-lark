package contact

import (
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

// PagedResponse is used for agent mode paged responses.
type PagedResponse struct {
	Items      any    `json:"items"`
	NextCursor string `json:"next_cursor"`
}

func getGlobalFlags(cmd *cobra.Command) (format, tokenMode, profile, config, domain string, debug, quiet, agent bool) {
	g := cmdutil.GetGlobalFlags(cmd)
	return g.Format, g.TokenMode, g.Profile, g.Config, g.Domain, g.Debug, g.Quiet, g.Agent
}
