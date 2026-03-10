package base

import (
	"encoding/json"
	"fmt"
	"os"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newRecordsCreateCommand() *cobra.Command {
	var fieldsJSON string
	var fieldPairs []string

	cmd := &cobra.Command{
		Use:   "create <URL>",
		Short: "创建记录",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if fieldsJSON == "" && len(fieldPairs) == 0 {
				return fmt.Errorf("MISSING_FLAG：--fields 或 --field 至少提供一个")
			}

			fields := make(map[string]interface{})
			if fieldsJSON != "" {
				if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
					return fmt.Errorf("INVALID_JSON：--fields 解析失败：%s", err.Error())
				}
			}
			for _, pair := range fieldPairs {
				k, v, err := parseFieldPair(pair)
				if err != nil {
					return fmt.Errorf("INVALID_FIELD：%s", err.Error())
				}
				fields[k] = v
			}

			appToken, tableID, err := parseBitableURLStrict(args[0])
			if err != nil {
				return err
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			record := larkbitable.NewAppTableRecordBuilder().
				Fields(fields).
				Build()

			req := larkbitable.NewCreateAppTableRecordReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				AppTableRecord(record).
				Build()

			resp, err := c.Client.Bitable.AppTableRecord.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			recordID := ""
			if resp.Data.Record != nil && resp.Data.Record.RecordId != nil {
				recordID = *resp.Data.Record.RecordId
			}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, map[string]string{"record_id": recordID})
			}

			fmt.Println(recordID)
			return nil
		},
	}

	cmd.Flags().StringVar(&fieldsJSON, "fields", "", `字段 JSON（如 '{"姓名":"张三"}'）`)
	cmd.Flags().StringArrayVar(&fieldPairs, "field", nil, `单个字段键值对，可重复使用（如 --field "姓名=张三" --field "年龄=25"）`)
	return cmd
}
