package docs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newListCommand() *cobra.Command {
	var folder string
	var since string
	var limit int
	var all bool
	var cursor string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列举文档",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			folderToken := ""
			if folder != "" {
				folderToken = ExtractFolderToken(folder)
			}

			var sinceTime *time.Time
			if since != "" {
				t, err := parseSince(since)
				if err != nil {
					return fmt.Errorf("INVALID_SINCE：%s", err.Error())
				}
				sinceTime = &t
			}

			type fileItem struct {
				Token      string `json:"token"`
				Name       string `json:"name"`
				Type       string `json:"type"`
				URL        string `json:"url"`
				ModifiedAt string `json:"modified_at,omitempty"`
			}

			var items []fileItem
			var nextToken string

			pageToken := cursor
			pageSize := int32(limit)

			for {
				req := larkdrive.NewListFileReqBuilder().
					PageSize(int(pageSize))

				if folderToken != "" {
					req = req.FolderToken(folderToken)
				}
				if pageToken != "" {
					req = req.PageToken(pageToken)
				}

				resp, err := c.Client.Drive.File.List(cmd.Context(), req.Build(), c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				for _, f := range resp.Data.Files {
					item := fileItem{}
					if f.Token != nil {
						item.Token = *f.Token
					}
					if f.Name != nil {
						item.Name = *f.Name
					}
					if f.Type != nil {
						item.Type = *f.Type
					}
					if f.Url != nil {
						item.URL = *f.Url
					}
					if f.ModifiedTime != nil {
						if ts, err := strconv.ParseInt(*f.ModifiedTime, 10, 64); err == nil {
							t := time.Unix(ts, 0)
							item.ModifiedAt = t.Format("2006-01-02 15:04:05")
						}
					}

					if sinceTime != nil {
						if f.ModifiedTime == nil {
							continue
						}
						if ts, err := strconv.ParseInt(*f.ModifiedTime, 10, 64); err == nil {
							t := time.Unix(ts, 0)
							if t.Before(*sinceTime) {
								continue
							}
						}
					}

					items = append(items, item)
				}

				hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
				nextPageToken := ""
				if resp.Data.NextPageToken != nil {
					nextPageToken = *resp.Data.NextPageToken
				}

				if !all || !hasMore || nextPageToken == "" {
					if hasMore && nextPageToken != "" {
						nextToken = nextPageToken
					}
					break
				}
				pageToken = nextPageToken
			}

			if g.Format == "json" {
				if g.Agent {
					return output.PrintJSON(os.Stdout, PagedResponse{Items: items, NextCursor: nextToken})
				}
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				if item.ModifiedAt != "" {
					fmt.Printf("%s\t%s\t（修改于 %s）\n", item.Name, item.URL, item.ModifiedAt)
				} else {
					fmt.Printf("%s\t%s\n", item.Name, item.URL)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&folder, "folder", "", "文件夹 URL 或 token")
	cmd.Flags().StringVar(&since, "since", "", "按修改时间过滤，如 24h/7d/today/2024-11-01")
	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().StringVar(&cursor, "cursor", "", "分页游标")

	return cmd
}

func parseSince(since string) (time.Time, error) {
	now := time.Now()
	switch since {
	case "24h":
		return now.Add(-24 * time.Hour), nil
	case "7d":
		return now.Add(-7 * 24 * time.Hour), nil
	case "today":
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location()), nil
	default:
		t, err := time.Parse("2006-01-02", since)
		if err != nil {
			return time.Time{}, fmt.Errorf("无法解析日期格式 %q，支持: 24h, 7d, today, YYYY-MM-DD", since)
		}
		return t, nil
	}
}

// PagedResponse is used for agent mode paged responses.
type PagedResponse = cmdutil.PagedResponse
