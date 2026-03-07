package wiki

import (
	"fmt"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newCreateCommand() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "create <wiki-space-url-or-token>",
		Short: "在知识空间创建新页面",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}

			if title == "" {
				return fmt.Errorf("MISSING_FLAG：--title 是必填项")
			}

			spaceToken := ExtractWikiToken(args[0])
			if spaceToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析 wiki token")
			}

			c, err := client.New(client.Options{
				TokenMode: tokenMode,
				Debug:     debug,
				Profile:   profile,
				Config:    cfg,
				Domain:    domain,
			})
			if err != nil {
				return fmt.Errorf("CLIENT_ERROR：%s", err.Error())
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

			pageURL := fmt.Sprintf("https://feishu.cn/wiki/%s", nodeToken)
			output.PrintSuccess(quiet, fmt.Sprintf("页面已创建：%s", pageURL))

			if output.GlobalAgent {
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
