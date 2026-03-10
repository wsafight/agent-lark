package perms

import (
	"context"
	"fmt"
	"os"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

type memberItem struct {
	MemberType string `json:"member_type"`
	MemberID   string `json:"member_id"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
	Perm       string `json:"perm"`
	Role       string `json:"role,omitempty"` // owner / collaborator
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <URL>",
		Short: "列举文档协作者",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			token, fileType := ExtractResourceToken(args[0])

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

			items, err := listMembers(c, token, fileType)
			if err != nil {
				return err
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, items)
			}

			for _, m := range items {
				fmt.Printf("%s\t%s\t%s\t%s\n",
					m.Name, m.MemberID, m.Perm, m.Role)
			}
			return nil
		},
	}
	return cmd
}

func listMembers(c *client.Result, token, fileType string) ([]memberItem, error) {
	req := larkdrive.NewListPermissionMemberReqBuilder().
		Type(fileType).
		Token(token).
		Build()

	resp, err := c.Client.Drive.PermissionMember.List(
		context.Background(),
		req,
		c.RequestOptions()...,
	)
	if err != nil {
		return nil, fmt.Errorf("API_ERROR：%s", err.Error())
	}
	if !resp.Success() {
		return nil, fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
	}

	var items []memberItem
	for _, m := range resp.Data.Items {
		item := memberItem{}
		if m.MemberType != nil {
			item.MemberType = *m.MemberType
		}
		if m.MemberId != nil {
			item.MemberID = *m.MemberId
		}
		if m.Name != nil {
			item.Name = *m.Name
		}
		if m.Perm != nil {
			item.Perm = *m.Perm
		}
		if m.Type != nil {
			item.Role = *m.Type
		}
		items = append(items, item)
	}

	return items, nil
}
