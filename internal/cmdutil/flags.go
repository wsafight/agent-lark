package cmdutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

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

// envStr returns the env var value when the named flag was not explicitly set.
func envStr(f interface{ Changed(string) bool }, flagName, envKey string, val string) string {
	if !f.Changed(flagName) {
		if v := os.Getenv(envKey); v != "" {
			return v
		}
	}
	return val
}

// envBool returns true when the env var is "1" or "true" and the flag was not explicitly set.
func envBool(f interface{ Changed(string) bool }, flagName, envKey string, val bool) bool {
	if !f.Changed(flagName) {
		v := os.Getenv(envKey)
		if v == "1" || v == "true" {
			return true
		}
	}
	return val
}

// GetGlobalFlags reads the root command's persistent flags from any subcommand.
// Env vars (AGENT_LARK_*) are used as fallbacks when a flag is not explicitly set.
func GetGlobalFlags(cmd *cobra.Command) GlobalFlags {
	root := cmd.Root()
	f := root.PersistentFlags()
	var g GlobalFlags
	raw, _ := f.GetString("format")
	g.Format = envStr(f, "format", "AGENT_LARK_FORMAT", raw)
	raw, _ = f.GetString("token-mode")
	g.TokenMode = envStr(f, "token-mode", "AGENT_LARK_TOKEN_MODE", raw)
	raw, _ = f.GetString("profile")
	g.Profile = envStr(f, "profile", "AGENT_LARK_PROFILE", raw)
	raw, _ = f.GetString("config")
	g.Config = envStr(f, "config", "AGENT_LARK_CONFIG", raw)
	raw, _ = f.GetString("domain")
	g.Domain = envStr(f, "domain", "AGENT_LARK_DOMAIN", raw)
	braw, _ := f.GetBool("debug")
	g.Debug = envBool(f, "debug", "AGENT_LARK_DEBUG", braw)
	braw, _ = f.GetBool("quiet")
	g.Quiet = envBool(f, "quiet", "AGENT_LARK_QUIET", braw)
	braw, _ = f.GetBool("agent")
	g.Agent = envBool(f, "agent", "AGENT_LARK_AGENT", braw)
	return g
}

// ResolveGlobalFlags applies cross-cutting output behavior from global flags.
func ResolveGlobalFlags(cmd *cobra.Command) GlobalFlags {
	g := GetGlobalFlags(cmd)
	if g.Agent {
		output.GlobalAgent = true
		g.Format = "json"
	}
	g.Format = output.FormatFromCmd(g.Format)
	return g
}

// ClientOptions returns a client.Options populated from the global flags.
func (g GlobalFlags) ClientOptions() client.Options {
	return client.Options{
		TokenMode: g.TokenMode,
		Debug:     g.Debug,
		Profile:   g.Profile,
		Config:    g.Config,
		Domain:    g.Domain,
	}
}

// NewClient creates a Feishu client from global flags, wrapping the error with CLIENT_ERROR code.
func (g GlobalFlags) NewClient() (*client.Result, error) {
	c, err := client.New(g.ClientOptions())
	if err != nil {
		return nil, fmt.Errorf("CLIENT_ERROR：%s", err.Error())
	}
	return c, nil
}
