package perms

import (
	"context"
	"fmt"
	"os"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newPublicCommand() *cobra.Command {
	var linkShare string
	var comment string
	var yes bool

	cmd := &cobra.Command{
		Use:   "public <URL>",
		Short: "查看/修改外部分享设置",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet
			globalYes, _ := cmd.Root().PersistentFlags().GetBool("yes")

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

			hasWrite := linkShare != "" || comment != ""

			if !hasWrite {
				// Read-only: Get public settings
				req := larkdrive.NewGetPermissionPublicReqBuilder().
					Token(token).
					Type(fileType).
					Build()

				resp, err := c.Client.Drive.PermissionPublic.Get(
					context.Background(),
					req,
					c.RequestOptions()...,
				)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				type publicSettings struct {
					ExternalAccess  bool   `json:"external_access"`
					SecurityEntity  string `json:"security_entity,omitempty"`
					CommentEntity   string `json:"comment_entity,omitempty"`
					ShareEntity     string `json:"share_entity,omitempty"`
					LinkShareEntity string `json:"link_share_entity,omitempty"`
				}

				s := publicSettings{}
				if resp.Data.PermissionPublic != nil {
					p := resp.Data.PermissionPublic
					if p.ExternalAccess != nil {
						s.ExternalAccess = *p.ExternalAccess
					}
					if p.SecurityEntity != nil {
						s.SecurityEntity = *p.SecurityEntity
					}
					if p.CommentEntity != nil {
						s.CommentEntity = *p.CommentEntity
					}
					if p.ShareEntity != nil {
						s.ShareEntity = *p.ShareEntity
					}
					if p.LinkShareEntity != nil {
						s.LinkShareEntity = *p.LinkShareEntity
					}
				}

				if format == "json" {
					return output.PrintJSON(os.Stdout, s)
				}

				fmt.Printf("外部访问:   %v\n", s.ExternalAccess)
				fmt.Printf("链接分享:   %s\n", s.LinkShareEntity)
				fmt.Printf("评论权限:   %s\n", s.CommentEntity)
				fmt.Printf("安全设置:   %s\n", s.SecurityEntity)
				return nil
			}

			// Write mode
			if !yes && !globalYes && !output.GlobalAgent {
				fmt.Printf("修改文档外部分享设置，确认？[y/N]: ")
				var input string
				fmt.Scan(&input)
				if input != "y" {
					fmt.Println("已取消")
					return nil
				}
			}

			builder := larkdrive.NewPermissionPublicRequestBuilder()

			if linkShare != "" {
				switch linkShare {
				case "off":
					builder = builder.ExternalAccess(false).LinkShareEntity("close")
				case "tenant":
					builder = builder.ExternalAccess(true).LinkShareEntity("tenant_readable")
				case "anyone":
					builder = builder.ExternalAccess(true).LinkShareEntity("anyone_readable")
				}
			}

			if comment != "" {
				switch comment {
				case "allow":
					builder = builder.CommentEntity("anyone_can_comment")
				case "deny":
					builder = builder.CommentEntity("no_one_can_comment")
				}
			}

			req := larkdrive.NewPatchPermissionPublicReqBuilder().
				Token(token).
				Type(fileType).
				PermissionPublicRequest(builder.Build()).
				Build()

			resp, err := c.Client.Drive.PermissionPublic.Patch(
				context.Background(),
				req,
				c.RequestOptions()...,
			)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			fmt.Println("✓ 外部分享设置已更新")
			return nil
		},
	}

	cmd.Flags().StringVar(&linkShare, "link-share", "", "链接分享范围：off|tenant|anyone")
	cmd.Flags().StringVar(&comment, "comment", "", "评论权限：allow|deny")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	return cmd
}
