package template

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/docs"
	"github.com/wsafight/agent-lark/internal/output"
)

func getGlobalFlags(cmd *cobra.Command) (format, tokenMode, profile, config, domain string, debug, quiet, agent bool) {
	g := cmdutil.GetGlobalFlags(cmd)
	return g.Format, g.TokenMode, g.Profile, g.Config, g.Domain, g.Debug, g.Quiet, g.Agent
}

// NewCommand returns the template subcommand group.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "template", Short: "本地模板管理"}
	cmd.AddCommand(
		newTemplateListCommand(),
		newTemplateSaveCommand(),
		newTemplateGetCommand(),
		newTemplateVarsCommand(),
		newTemplateDeleteCommand(),
		newTemplateApplyCommand(),
	)
	return cmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有本地模板",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _, _, _, _, _, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			templates, err := ListAll()
			if err != nil {
				return fmt.Errorf("读取模板失败：%w", err)
			}

			if len(templates) == 0 {
				if format == "json" {
					return output.PrintJSON(os.Stdout, []any{})
				}
				fmt.Println("（无模板，运行 'agent-lark template save <名称>' 创建）")
				return nil
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, templates)
			}

			for _, t := range templates {
				desc := t.Description
				if desc == "" {
					desc = "-"
				}
				fmt.Printf("%s\t%s\t（保存于 %s）\n", t.Name, desc, t.UpdatedAt.Format("2006-01-02 15:04"))
			}
			return nil
		},
	}
}

func newTemplateSaveCommand() *cobra.Command {
	var filePath string
	var fromURL string
	var content string
	var description string
	var force bool

	cmd := &cobra.Command{
		Use:   "save <名称>",
		Short: "保存模板",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}
			_ = quiet

			name := args[0]

			var tmplContent string
			var source string

			switch {
			case fromURL != "":
				// Fetch from Feishu doc
				source = fromURL
				docToken := docs.ExtractDocID(fromURL)
				if docToken == "" {
					return fmt.Errorf("INVALID_URL：无法解析文档 token")
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

				// Get document metadata (title)
				docResp, err := c.Client.Docx.Document.Get(
					context.Background(),
					larkdocx.NewGetDocumentReqBuilder().DocumentId(docToken).Build(),
					c.RequestOptions()...,
				)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !docResp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", docResp.Code, docResp.Msg)
				}

				// Get all blocks
				var allBlocks []*larkdocx.Block
				var pageToken string
				for {
					builder := larkdocx.NewListDocumentBlockReqBuilder().
						DocumentId(docToken).
						PageSize(200)
					if pageToken != "" {
						builder = builder.PageToken(pageToken)
					}
					resp, err := c.Client.Docx.DocumentBlock.List(
						context.Background(),
						builder.Build(),
						c.RequestOptions()...,
					)
					if err != nil {
						return fmt.Errorf("API_ERROR：%s", err.Error())
					}
					if !resp.Success() {
						return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
					}
					allBlocks = append(allBlocks, resp.Data.Items...)
					hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
					if !hasMore {
						break
					}
					if resp.Data.PageToken == nil || *resp.Data.PageToken == "" {
						break
					}
					pageToken = *resp.Data.PageToken
				}

				// Convert blocks to markdown
				var sb strings.Builder
				for _, b := range allBlocks {
					if b == nil || b.BlockType == nil {
						continue
					}
					text := extractBlockText(b)
					switch *b.BlockType {
					case 1: // page
						// skip
					case 2:
						if text != "" {
							sb.WriteString(text)
							sb.WriteString("\n\n")
						}
					case 3:
						sb.WriteString("# " + text + "\n\n")
					case 4:
						sb.WriteString("## " + text + "\n\n")
					case 5:
						sb.WriteString("### " + text + "\n\n")
					case 6:
						sb.WriteString("#### " + text + "\n\n")
					case 7:
						sb.WriteString("##### " + text + "\n\n")
					case 8:
						sb.WriteString("###### " + text + "\n\n")
					case 9:
						sb.WriteString("1. " + text + "\n")
					case 10:
						sb.WriteString("- " + text + "\n")
					case 11:
						sb.WriteString("```\n" + text + "\n```\n\n")
					case 12:
						sb.WriteString("> " + text + "\n\n")
					case 19:
						sb.WriteString("---\n\n")
					default:
						if text != "" {
							sb.WriteString(text + "\n\n")
						}
					}
				}
				tmplContent = strings.TrimSpace(sb.String())

			case filePath != "":
				source = "local"
				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("读取文件失败：%w", err)
				}
				tmplContent = string(data)

			case content != "":
				source = "local"
				tmplContent = content

			default:
				return fmt.Errorf("MISSING_FLAG：请提供 --file、--from 或 --content")
			}

			now := time.Now()
			t := &Template{
				Name:        name,
				Description: description,
				Content:     tmplContent,
				Source:      source,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// If force and existing, preserve CreatedAt
			if force {
				if existing, err := Load(name); err == nil {
					t.CreatedAt = existing.CreatedAt
				}
			}

			if err := Save(t, force); err != nil {
				return err
			}

			fmt.Printf("✓ 模板 %q 已保存（%d 字符）\n", name, len(tmplContent))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "本地 Markdown 文件路径")
	cmd.Flags().StringVar(&fromURL, "from", "", "从飞书文档 URL 拉取内容")
	cmd.Flags().StringVar(&content, "content", "", "内联 Markdown 内容")
	cmd.Flags().StringVar(&description, "description", "", "模板描述")
	cmd.Flags().BoolVar(&force, "force", false, "覆盖已有模板")
	return cmd
}

func extractBlockText(b *larkdocx.Block) string {
	if b == nil || b.BlockType == nil {
		return ""
	}
	extractFromElements := func(elements []*larkdocx.TextElement) string {
		var sb strings.Builder
		for _, el := range elements {
			if el != nil && el.TextRun != nil && el.TextRun.Content != nil {
				sb.WriteString(*el.TextRun.Content)
			}
		}
		return sb.String()
	}
	switch *b.BlockType {
	case 2:
		if b.Text != nil {
			return extractFromElements(b.Text.Elements)
		}
	case 3:
		if b.Heading1 != nil {
			return extractFromElements(b.Heading1.Elements)
		}
	case 4:
		if b.Heading2 != nil {
			return extractFromElements(b.Heading2.Elements)
		}
	case 5:
		if b.Heading3 != nil {
			return extractFromElements(b.Heading3.Elements)
		}
	case 6:
		if b.Heading4 != nil {
			return extractFromElements(b.Heading4.Elements)
		}
	case 7:
		if b.Heading5 != nil {
			return extractFromElements(b.Heading5.Elements)
		}
	case 8:
		if b.Heading6 != nil {
			return extractFromElements(b.Heading6.Elements)
		}
	case 9:
		if b.Ordered != nil {
			return extractFromElements(b.Ordered.Elements)
		}
	case 10:
		if b.Bullet != nil {
			return extractFromElements(b.Bullet.Elements)
		}
	case 11:
		if b.Code != nil {
			return extractFromElements(b.Code.Elements)
		}
	case 12:
		if b.Quote != nil {
			return extractFromElements(b.Quote.Elements)
		}
	}
	return ""
}

func newTemplateGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <名称>",
		Short: "显示模板原始内容",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _, _, _, _, _, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			t, err := Load(args[0])
			if err != nil {
				return err
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, t)
			}

			fmt.Print(t.Content)
			return nil
		},
	}
}

func newTemplateVarsCommand() *cobra.Command {
	var customValues []string

	cmd := &cobra.Command{
		Use:   "vars <名称>",
		Short: "分析模板变量",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _, _, _, _, _, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			t, err := Load(args[0])
			if err != nil {
				return err
			}

			allVarNames := ExtractVars(t.Content)
			builtin := BuiltinVars("", "")

			// Parse custom values from --var flags
			customMap := map[string]string{}
			for _, kv := range customValues {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					customMap[parts[0]] = parts[1]
				}
			}

			type varInfo struct {
				Name     string  `json:"name"`
				Resolved *string `json:"resolved"`
			}

			type varsResult struct {
				Template string    `json:"template"`
				Builtin  []varInfo `json:"builtin"`
				Custom   []varInfo `json:"custom"`
			}

			result := varsResult{Template: args[0]}

			for _, name := range allVarNames {
				if v, ok := builtin[name]; ok {
					resolved := v
					result.Builtin = append(result.Builtin, varInfo{Name: name, Resolved: &resolved})
				} else {
					if v, ok := customMap[name]; ok {
						resolved := v
						result.Custom = append(result.Custom, varInfo{Name: name, Resolved: &resolved})
					} else {
						result.Custom = append(result.Custom, varInfo{Name: name, Resolved: nil})
					}
				}
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, result)
			}

			fmt.Printf("模板: %s\n", result.Template)
			if len(result.Builtin) > 0 {
				fmt.Println("内置变量:")
				for _, v := range result.Builtin {
					if v.Resolved != nil {
						fmt.Printf("  {{%s}} = %s\n", v.Name, *v.Resolved)
					} else {
						fmt.Printf("  {{%s}} = (null)\n", v.Name)
					}
				}
			}
			if len(result.Custom) > 0 {
				fmt.Println("自定义变量:")
				for _, v := range result.Custom {
					if v.Resolved != nil {
						fmt.Printf("  {{%s}} = %s\n", v.Name, *v.Resolved)
					} else {
						fmt.Printf("  {{%s}} = (null) ← 未提供值，将保留原样\n", v.Name)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&customValues, "var", nil, "自定义变量值 key=value（可重复）")
	return cmd
}

func newTemplateDeleteCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <名称>",
		Short: "删除模板",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, _, _, _, _, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}
			_ = quiet
			globalYes, _ := cmd.Root().PersistentFlags().GetBool("yes")

			name := args[0]

			if !yes && !globalYes && !output.GlobalAgent {
				fmt.Printf("确认删除模板 %q？[y/N]: ", name)
				var input string
				fmt.Scan(&input)
				if input != "y" {
					fmt.Println("已取消")
					return nil
				}
			}

			if err := Delete(name); err != nil {
				return err
			}
			fmt.Printf("✓ 模板 %q 已删除\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	return cmd
}

func newTemplateApplyCommand() *cobra.Command {
	var newDoc bool
	var title string
	var folderURL string
	var targetURL string
	var after string
	var before string
	var matchIndex int
	var customVars []string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply <名称>",
		Short: "应用模板（创建新文档或追加到已有文档）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}
			_ = quiet

			if !newDoc && targetURL == "" {
				return fmt.Errorf("MISSING_FLAG：请提供 --new 或 --to <URL>")
			}
			if newDoc && title == "" {
				return fmt.Errorf("MISSING_FLAG：--new 时需要提供 --title")
			}
			if after != "" && before != "" {
				return fmt.Errorf("INVALID_FLAGS：--after 和 --before 不能同时使用")
			}

			// Parse custom vars
			customMap := map[string]string{}
			for _, kv := range customVars {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					customMap[parts[0]] = parts[1]
				}
			}

			// Get author name from config if available
			authorName := ""
			c, err := client.New(client.Options{
				TokenMode: tokenMode,
				Debug:     debug,
				Profile:   profile,
				Config:    cfg,
				Domain:    domain,
			})
			if err == nil && c.Cfg != nil && c.Cfg.UserSession != nil {
				authorName = c.Cfg.UserSession.Name
			}

			opts := ApplyOptions{
				TemplateName: args[0],
				New:          newDoc,
				Title:        title,
				FolderURL:    folderURL,
				TargetURL:    targetURL,
				After:        after,
				Before:       before,
				MatchIndex:   matchIndex,
				CustomVars:   customMap,
				DryRun:       dryRun,
				ClientOpts: client.Options{
					TokenMode: tokenMode,
					Debug:     debug,
					Profile:   profile,
					Config:    cfg,
					Domain:    domain,
				},
				AuthorName: authorName,
			}

			url, err := Apply(opts)
			if err != nil {
				return err
			}

			if url != "" {
				fmt.Printf("✓ 已应用模板 %q\n", args[0])
				fmt.Println(url)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&newDoc, "new", false, "创建新文档")
	cmd.Flags().StringVar(&title, "title", "", "新文档标题（--new 时使用）")
	cmd.Flags().StringVar(&folderURL, "folder", "", "新文档所在文件夹 URL（--new 时使用）")
	cmd.Flags().StringVar(&targetURL, "to", "", "追加到已有文档的 URL")
	cmd.Flags().StringVar(&after, "after", "", "插入到含该关键词的段落之后")
	cmd.Flags().StringVar(&before, "before", "", "插入到含该关键词的段落之前")
	cmd.Flags().IntVar(&matchIndex, "match-index", 0, "当关键词匹配多个段落时，选择第 N 个（从 0 开始）")
	cmd.Flags().StringArrayVar(&customVars, "var", nil, "自定义变量 key=value（可重复）")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "预览而不实际写入")
	return cmd
}
