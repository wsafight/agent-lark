package docs

import (
	"fmt"
	"os"

	larksearch "github.com/larksuite/oapi-sdk-go/v3/service/search/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newSearchCommand() *cobra.Command {
	var limit int
	var existsOnly bool

	cmd := &cobra.Command{
		Use:   "search <keyword>",
		Short: "按关键词搜索文档",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			keyword := args[0]

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

			req := larksearch.NewSearchDocWikiReqBuilder().
				Body(larksearch.NewSearchDocWikiReqBodyBuilder().
					Query(keyword).
					PageSize(limit).
					Build()).
				Build()

			resp, err := c.Client.Search.DocWiki.Search(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			units := resp.Data.ResUnits
			count := len(units)

			if existsOnly {
				if count > 0 {
					fmt.Printf("found %d\n", count)
				} else {
					fmt.Println("not_found")
				}
				return nil
			}

			type fileItem struct {
				Token string `json:"token"`
				Title string `json:"title"`
				Type  string `json:"type"`
				URL   string `json:"url"`
			}

			var items []fileItem
			for _, u := range units {
				item := fileItem{}
				if u.TitleHighlighted != nil {
					item.Title = *u.TitleHighlighted
				}
				if u.EntityType != nil {
					item.Type = *u.EntityType
				}
				if u.ResultMeta != nil {
					if u.ResultMeta.Token != nil {
						item.Token = *u.ResultMeta.Token
					}
					if u.ResultMeta.Url != nil {
						item.URL = *u.ResultMeta.Url
					}
				}
				items = append(items, item)
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("%s\t%s\n", item.Title, item.URL)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")
	cmd.Flags().BoolVar(&existsOnly, "exists", false, "只输出是否存在")

	return cmd
}
