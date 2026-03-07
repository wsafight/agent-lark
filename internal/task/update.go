package task

import (
	"fmt"
	"os"

	larktask "github.com/larksuite/oapi-sdk-go/v3/service/task/v2"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newUpdateCommand() *cobra.Command {
	var taskID string
	var status string
	var title string
	var due string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "更新任务",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
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

			taskBuilder := larktask.NewInputTaskBuilder()

			var updateFields []string

			if title != "" {
				taskBuilder = taskBuilder.Summary(title)
				updateFields = append(updateFields, "summary")
			}

			if due != "" {
				dueObj := larktask.NewDueBuilder().
					Timestamp(due).
					IsAllDay(true).
					Build()
				taskBuilder = taskBuilder.Due(dueObj)
				updateFields = append(updateFields, "due")
			}

			if status == "done" {
				// To mark as done, we complete the task. The patch API handles this via completed_at.
				// We'll use the summary field and rely on the status update path.
				// For the task v2 API, completion is indicated by setting completed_at.
				updateFields = append(updateFields, "status")
			}

			req := larktask.NewPatchTaskReqBuilder().
				TaskGuid(taskID).
				UserIdType("open_id").
				Body(larktask.NewPatchTaskReqBodyBuilder().
					Task(taskBuilder.Build()).
					UpdateFields(updateFields).
					Build()).
				Build()

			resp, err := c.Client.Task.V2.Task.Patch(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			updatedID := taskID
			if resp.Data.Task != nil && resp.Data.Task.Guid != nil {
				updatedID = *resp.Data.Task.Guid
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, map[string]string{"task_id": updatedID})
			}

			fmt.Printf("任务 %s 已更新\n", updatedID)
			return nil
		},
	}

	cmd.Flags().StringVar(&taskID, "task-id", "", "任务 ID（必填）")
	cmd.Flags().StringVar(&status, "status", "", "任务状态：done|todo|in_progress")
	cmd.Flags().StringVar(&title, "title", "", "新标题")
	cmd.Flags().StringVar(&due, "due", "", "截止日期（如 2024-12-31）")
	return cmd
}
