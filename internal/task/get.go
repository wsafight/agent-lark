package task

import (
	"fmt"
	"os"

	larktask "github.com/larksuite/oapi-sdk-go/v3/service/task/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newGetCommand() *cobra.Command {
	var taskID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "获取任务详情",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			if taskID == "" {
				return fmt.Errorf("MISSING_FLAG：--task-id 为必填项")
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

			req := larktask.NewGetTaskReqBuilder().
				TaskGuid(taskID).
				Build()

			resp, err := c.Client.Task.V2.Task.Get(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			type taskDetail struct {
				TaskID      string `json:"task_id"`
				Title       string `json:"title"`
				Description string `json:"description,omitempty"`
				Status      string `json:"status"`
				Due         string `json:"due,omitempty"`
				CreatedAt   string `json:"created_at,omitempty"`
				UpdatedAt   string `json:"updated_at,omitempty"`
			}

			detail := taskDetail{}
			if resp.Data.Task != nil {
				t := resp.Data.Task
				if t.Guid != nil {
					detail.TaskID = *t.Guid
				}
				if t.Summary != nil {
					detail.Title = *t.Summary
				}
				if t.Description != nil {
					detail.Description = *t.Description
				}
				completedAt := ""
				if t.CompletedAt != nil {
					completedAt = *t.CompletedAt
				}
				if t.Status != nil {
					detail.Status = deriveTaskStatus(*t.Status, completedAt)
				} else {
					detail.Status = deriveTaskStatus("", completedAt)
				}
				if t.Due != nil && t.Due.Timestamp != nil {
					detail.Due = *t.Due.Timestamp
				}
				if t.CreatedAt != nil {
					detail.CreatedAt = *t.CreatedAt
				}
				if t.UpdatedAt != nil {
					detail.UpdatedAt = *t.UpdatedAt
				}
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, detail)
			}

			fmt.Printf("task_id:     %s\n", detail.TaskID)
			fmt.Printf("标题:         %s\n", detail.Title)
			fmt.Printf("状态:         %s\n", detail.Status)
			if detail.Description != "" {
				fmt.Printf("描述:         %s\n", detail.Description)
			}
			if detail.Due != "" {
				fmt.Printf("截止:         %s\n", detail.Due)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&taskID, "task-id", "", "任务 ID（必填）")
	return cmd
}
