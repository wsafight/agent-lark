package wiki

import (
	"fmt"
	"strings"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newCreateCommand() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "create <wiki-space-url-or-token>",
		Short: "在知识空间创建新页面",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if title == "" {
				return fmt.Errorf("MISSING_FLAG：--title 是必填项")
			}

			spaceToken := ExtractWikiToken(args[0])
			if spaceToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析 wiki token")
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			objType := "docx"
			req := larkwiki.NewCreateSpaceNodeReqBuilder().
				SpaceId(spaceToken).
				Node(
					larkwiki.NewNodeBuilder().
						ObjType(objType).
						Title(title).
						Build(),
				).
				Build()

			resp, err := c.Client.Wiki.SpaceNode.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			nodeToken := ""
			if resp.Data.Node != nil && resp.Data.Node.NodeToken != nil {
				nodeToken = *resp.Data.Node.NodeToken
			}

			effectiveOpenDomain := g.Domain
			if effectiveOpenDomain == "" && c.Cfg != nil {
				effectiveOpenDomain = c.Cfg.Domain
			}
			webDomain := "feishu.cn"
			if strings.Contains(strings.ToLower(effectiveOpenDomain), "larksuite") {
				webDomain = "larksuite.com"
			}

			pageURL := fmt.Sprintf("https://%s/wiki/%s", webDomain, nodeToken)
			output.PrintSuccess(g.Quiet, fmt.Sprintf("页面已创建：%s", pageURL))

			if g.Agent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]string{
					"node_token": nodeToken,
					"url":        pageURL,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "页面标题（必填）")

	return cmd
}
