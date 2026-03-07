package wiki

import (
	"fmt"
	"os"
	"strings"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newListCommand() *cobra.Command {
	var limit int
	var all bool
	var depth int

	cmd := &cobra.Command{
		Use:   "list <wiki-url-or-token>",
		Short: "列举知识空间节点",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

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

			type NodeItem struct {
				NodeToken  string `json:"node_token"`
				ObjToken   string `json:"obj_token"`
				ObjType    string `json:"obj_type"`
				Title      string `json:"title"`
				HasChild   bool   `json:"has_child"`
				Depth      int    `json:"depth"`
			}

			var items []NodeItem
			var nextToken string

			pageToken := ""
			pageSize := int32(limit)

			var fetchNodes func(parentToken string, currentDepth int) error
			fetchNodes = func(parentToken string, currentDepth int) error {
				if depth > 0 && currentDepth > depth {
					return nil
				}

				localPageToken := pageToken
				if currentDepth > 0 {
					localPageToken = ""
				}

				for {
					builder := larkwiki.NewListSpaceNodeReqBuilder().
						SpaceId(spaceToken).
						PageSize(int(pageSize))

					if parentToken != "" {
						builder = builder.ParentNodeToken(parentToken)
					}
					if localPageToken != "" {
						builder = builder.PageToken(localPageToken)
					}

					resp, err := c.Client.Wiki.SpaceNode.List(cmd.Context(), builder.Build(), c.RequestOptions()...)
					if err != nil {
						return fmt.Errorf("API_ERROR：%s", err.Error())
					}
					if !resp.Success() {
						return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
					}

					for _, node := range resp.Data.Items {
						item := NodeItem{Depth: currentDepth}
						if node.NodeToken != nil {
							item.NodeToken = *node.NodeToken
						}
						if node.ObjToken != nil {
							item.ObjToken = *node.ObjToken
						}
						if node.ObjType != nil {
							item.ObjType = *node.ObjType
						}
						if node.Title != nil {
							item.Title = *node.Title
						}
						if node.HasChild != nil {
							item.HasChild = *node.HasChild
						}
						items = append(items, item)

						if item.HasChild && (depth == 0 || currentDepth < depth) {
							if err := fetchNodes(item.NodeToken, currentDepth+1); err != nil {
								return err
							}
						}
					}

					hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
					nextPageToken := ""
					if resp.Data.PageToken != nil {
						nextPageToken = *resp.Data.PageToken
					}

					if currentDepth == 0 {
						if !all || !hasMore || nextPageToken == "" {
							if hasMore && nextPageToken != "" {
								nextToken = nextPageToken
							}
							break
						}
						localPageToken = nextPageToken
					} else {
						break
					}
				}
				return nil
			}

			if err := fetchNodes("", 0); err != nil {
				return err
			}

			if format == "json" {
				if agent {
					return output.PrintJSON(os.Stdout, PagedResponse{Items: items, NextCursor: nextToken})
				}
				return output.PrintJSON(os.Stdout, items)
			}

			// text format with indentation for hierarchy
			for _, item := range items {
				indent := strings.Repeat("  ", item.Depth)
				fmt.Printf("%s%s\t(%s)\n", indent, item.Title, item.NodeToken)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().IntVar(&depth, "depth", 0, "最大层级深度（0表示不限制）")

	return cmd
}

// PagedResponse is used for agent mode paged responses.
type PagedResponse struct {
	Items      any    `json:"items"`
	NextCursor string `json:"next_cursor"`
}

func getGlobalFlags(cmd *cobra.Command) (format, tokenMode, profile, config, domain string, debug, quiet, agent bool) {
	g := cmdutil.GetGlobalFlags(cmd)
	return g.Format, g.TokenMode, g.Profile, g.Config, g.Domain, g.Debug, g.Quiet, g.Agent
}
