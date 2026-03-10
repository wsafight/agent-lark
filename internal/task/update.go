package task

import (
	"fmt"
	"os"
	"strconv"
	"time"

	larktask "github.com/larksuite/oapi-sdk-go/v3/service/task/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
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
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			if taskID == "" {
				return fmt.Errorf("MISSING_FLAG：--task-id 为必填项")
			}
			normalizedStatus, err := normalizeTaskStatus(status)
			if err != nil {
				return err
			}
			status = normalizedStatus
			if title == "" && due == "" && status == "" {
				return fmt.Errorf("MISSING_FLAG：至少提供 --title / --due / --status 之一")
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
				dueMillis, err := parseDueToMillis(due)
				if err != nil {
					return fmt.Errorf("INVALID_DUE：%s", err.Error())
				}
				dueObj := larktask.NewDueBuilder().
					Timestamp(dueMillis).
					IsAllDay(true).
					Build()
				taskBuilder = taskBuilder.Due(dueObj)
				updateFields = append(updateFields, "due")
			}

			if status == "done" {
				taskBuilder = taskBuilder.CompletedAt(strconv.FormatInt(time.Now().UnixMilli(), 10))
				updateFields = append(updateFields, "completed_at")
			} else if status == "todo" {
				taskBuilder = taskBuilder.CompletedAt("0")
				updateFields = append(updateFields, "completed_at")
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
	cmd.Flags().StringVar(&status, "status", "", "任务状态：todo|done")
	cmd.Flags().StringVar(&title, "title", "", "新标题")
	cmd.Flags().StringVar(&due, "due", "", "截止时间（支持 YYYY-MM-DD / RFC3339 / Unix 秒或毫秒）")
	return cmd
}
