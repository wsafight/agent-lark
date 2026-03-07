package task

import (
	"fmt"
	"os"

	larktask "github.com/larksuite/oapi-sdk-go/v3/service/task/v2"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

type taskItem struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Due         string `json:"due,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
}

func newListCommand() *cobra.Command {
	var assignee string
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列举任务",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

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

			reqBuilder := larktask.NewListTaskReqBuilder().
				PageSize(limit)

			if assignee != "" {
				reqBuilder = reqBuilder.UserIdType("open_id")
			}

			req := reqBuilder.Build()

			resp, err := c.Client.Task.V2.Task.List(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			var items []taskItem
			for _, t := range resp.Data.Items {
				item := taskItem{}
				if t.Guid != nil {
					item.TaskID = *t.Guid
				}
				if t.Summary != nil {
					item.Title = *t.Summary
				}
				if t.CompletedAt != nil && *t.CompletedAt != "" {
					item.Status = "done"
				} else {
					item.Status = "todo"
				}
				if t.Due != nil && t.Due.Timestamp != nil {
					item.Due = *t.Due.Timestamp
				}
				if len(t.Members) > 0 {
					for _, m := range t.Members {
						if m.Id != nil {
							item.AssigneeID = *m.Id
							break
						}
					}
				}

				// Filter by assignee if specified
				if assignee != "" && item.AssigneeID != assignee {
					continue
				}

				// Filter by status if specified
				if status != "" && item.Status != status {
					continue
				}

				items = append(items, item)
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				due := ""
				if item.Due != "" {
					due = fmt.Sprintf("\t截止: %s", item.Due)
				}
				fmt.Printf("%s\t%s\t[%s]%s\n", item.TaskID, item.Title, item.Status, due)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&assignee, "assignee", "", "负责人 open_id")
	cmd.Flags().StringVar(&status, "status", "", "任务状态：todo|in_progress|done")
	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")
	return cmd
}
