package task

import (
	"fmt"
	"os"

	larktask "github.com/larksuite/oapi-sdk-go/v3/service/task/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newCreateCommand() *cobra.Command {
	var title string
	var description string
	var due string
	var assignee string
	var tasklistID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建任务",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if title == "" {
				return fmt.Errorf("MISSING_FLAG：--title 为必填项")
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			taskBuilder := larktask.NewInputTaskBuilder().
				Summary(title)

			if description != "" {
				taskBuilder = taskBuilder.Description(description)
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
			}

			if assignee != "" {
				member := larktask.NewMemberBuilder().
					Id(assignee).
					Type("user").
					Role("assignee").
					Build()
				taskBuilder = taskBuilder.Members([]*larktask.Member{member})
			}

			if tasklistID != "" {
				tasklist := larktask.NewTaskInTasklistInfoBuilder().
					TasklistGuid(tasklistID).
					Build()
				taskBuilder = taskBuilder.Tasklists([]*larktask.TaskInTasklistInfo{tasklist})
			}

			req := larktask.NewCreateTaskReqBuilder().
				UserIdType("open_id").
				InputTask(taskBuilder.Build()).
				Build()

			resp, err := c.Client.Task.V2.Task.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			createdID := ""
			if resp.Data.Task != nil && resp.Data.Task.Guid != nil {
				createdID = *resp.Data.Task.Guid
			}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, map[string]string{"task_id": createdID})
			}

			fmt.Println(createdID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "任务标题（必填）")
	cmd.Flags().StringVar(&description, "description", "", "任务描述")
	cmd.Flags().StringVar(&due, "due", "", "截止时间（支持 YYYY-MM-DD / RFC3339 / Unix 秒或毫秒）")
	cmd.Flags().StringVar(&assignee, "assignee", "", "负责人 open_id")
	cmd.Flags().StringVar(&tasklistID, "tasklist-id", "", "任务清单 ID")
	return cmd
}
