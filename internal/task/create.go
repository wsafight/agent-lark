package task

import (
	"fmt"
	"os"

	larktask "github.com/larksuite/oapi-sdk-go/v3/service/task/v2"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
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
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			if title == "" {
				return fmt.Errorf("MISSING_FLAG：--title 为必填项")
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

			taskBuilder := larktask.NewInputTaskBuilder().
				Summary(title)

			if description != "" {
				taskBuilder = taskBuilder.Description(description)
			}

			if due != "" {
				dueObj := larktask.NewDueBuilder().
					Timestamp(due).
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

			reqBuilder := larktask.NewCreateTaskReqBuilder().
				UserIdType("open_id").
				InputTask(taskBuilder.Build())

			if tasklistID != "" {
				origin := larktask.NewOriginBuilder().
					PlatformI18nName(larktask.NewI18nTextBuilder().ZhCn(tasklistID).Build()).
					Build()
				_ = origin // tasklist association handled separately
			}

			req := reqBuilder.Build()

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

			if format == "json" {
				return output.PrintJSON(os.Stdout, map[string]string{"task_id": createdID})
			}

			fmt.Println(createdID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "任务标题（必填）")
	cmd.Flags().StringVar(&description, "description", "", "任务描述")
	cmd.Flags().StringVar(&due, "due", "", "截止日期（如 2024-12-31）")
	cmd.Flags().StringVar(&assignee, "assignee", "", "负责人 open_id")
	cmd.Flags().StringVar(&tasklistID, "tasklist-id", "", "任务清单 ID")
	return cmd
}
