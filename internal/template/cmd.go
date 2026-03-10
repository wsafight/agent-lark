package template

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/docs"
	"github.com/wsafight/agent-lark/internal/docxutil"
	"github.com/wsafight/agent-lark/internal/output"
)

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
			format, _, _, _, _, _, quiet, _ := cmdutil.ResolveTuple(cmd)
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
			_, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			name := args[0]

			tmplContent, source, err := resolveTemplateSource(
				cmd.Context(),
				fromURL,
				filePath,
				content,
				client.Options{
					TokenMode: tokenMode,
					Debug:     debug,
					Profile:   profile,
					Config:    cfg,
					Domain:    domain,
				},
			)
			if err != nil {
				return err
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

func resolveTemplateSource(ctx context.Context, fromURL, filePath, inlineContent string, opts client.Options) (content, source string, err error) {
	switch {
	case fromURL != "":
		markdown, err := fetchTemplateFromDoc(ctx, fromURL, opts)
		if err != nil {
			return "", "", err
		}
		return markdown, fromURL, nil
	case filePath != "":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", "", fmt.Errorf("读取文件失败：%w", err)
		}
		return string(data), "local", nil
	case inlineContent != "":
		return inlineContent, "local", nil
	default:
		return "", "", fmt.Errorf("MISSING_FLAG：请提供 --file、--from 或 --content")
	}
}

func fetchTemplateFromDoc(ctx context.Context, fromURL string, opts client.Options) (string, error) {
	docToken := docs.ExtractDocID(fromURL)
	if docToken == "" {
		return "", fmt.Errorf("INVALID_URL：无法解析文档 token")
	}

	c, err := client.New(opts)
	if err != nil {
		return "", fmt.Errorf("CLIENT_ERROR：%s", err.Error())
	}

	blocks, err := docxutil.FetchAllBlocks(ctx, c, docToken)
	if err != nil {
		return "", err
	}
	outBlocks := docxutil.ConvertBlocks(blocks)
	return strings.TrimSpace(output.BlocksToMarkdown(outBlocks)), nil
}

func newTemplateGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <名称>",
		Short: "显示模板原始内容",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _, _, _, _, _, quiet, _ := cmdutil.ResolveTuple(cmd)
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
			format, _, _, _, _, _, quiet, _ := cmdutil.ResolveTuple(cmd)
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
			_, _, _, _, _, _, quiet, _ := cmdutil.ResolveTuple(cmd)
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
			_, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
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

			url, err := Apply(cmd.Context(), opts)
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
