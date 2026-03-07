package template

import (
	"context"
	"fmt"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/docs"
)

// ApplyOptions apply 命令的参数。
type ApplyOptions struct {
	TemplateName string
	// 创建新文档
	New       bool
	Title     string
	FolderURL string
	// 追加到已有文档
	TargetURL  string
	After      string
	Before     string
	MatchIndex int
	// 变量
	CustomVars map[string]string
	DryRun     bool
	// client options
	ClientOpts client.Options
	AuthorName string
}

// Apply 执行模板应用（创建新文档或追加到已有文档）。
func Apply(opts ApplyOptions) (string, error) {
	t, err := Load(opts.TemplateName)
	if err != nil {
		return "", err
	}

	// 合并变量
	allVars := BuiltinVars(opts.Title, opts.AuthorName)
	for k, v := range opts.CustomVars {
		allVars[k] = v
	}
	content := Render(t.Content, allVars)

	if opts.DryRun {
		fmt.Printf("[dry-run] 模板：%s\n变量替换后预览：\n\n%s\n", opts.TemplateName, content)
		if opts.After != "" {
			fmt.Printf("---\n插入位置：「%s」段落之后\n", opts.After)
		} else if opts.Before != "" {
			fmt.Printf("---\n插入位置：「%s」段落之前\n", opts.Before)
		}
		fmt.Println("确认无误后去掉 --dry-run 执行写入。")
		return "", nil
	}

	res, err := client.New(opts.ClientOpts)
	if err != nil {
		return "", err
	}

	if opts.New {
		// 创建新文档
		folderToken := docs.ExtractFolderToken(opts.FolderURL)
		bodyBuilder := larkdocx.NewCreateDocumentReqBodyBuilder().Title(opts.Title)
		if folderToken != "" {
			bodyBuilder = bodyBuilder.FolderToken(folderToken)
		}
		createResp, err := res.Client.Docx.Document.Create(
			context.Background(),
			larkdocx.NewCreateDocumentReqBuilder().Body(bodyBuilder.Build()).Build(),
			res.RequestOptions()...,
		)
		if err != nil {
			return "", fmt.Errorf("创建文档失败：%w", err)
		}
		if !createResp.Success() {
			return "", fmt.Errorf("API 错误：%s（code %d）", createResp.Msg, createResp.Code)
		}
		docToken := *createResp.Data.Document.DocumentId
		// 追加内容
		if err := appendContent(res, docToken, content, "", "", 0); err != nil {
			return "", err
		}
		domain := "feishu.cn"
		if res.Cfg != nil && strings.Contains(res.Cfg.Domain, "larksuite") {
			domain = "larksuite.com"
		}
		return fmt.Sprintf("https://company.%s/docx/%s", domain, docToken), nil
	}

	// 追加到已有文档
	docToken := docs.ExtractDocID(opts.TargetURL)
	if err := appendContent(res, docToken, content, opts.After, opts.Before, opts.MatchIndex); err != nil {
		return "", err
	}
	return opts.TargetURL, nil
}

func appendContent(res *client.Result, docToken, content, after, before string, matchIndex int) error {
	paragraphs := strings.Split(strings.TrimSpace(content), "\n\n")
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		textElem := larkdocx.NewTextElementBuilder().
			TextRun(larkdocx.NewTextRunBuilder().
				Content(para).
				Build()).
			Build()
		block := larkdocx.NewBlockBuilder().
			BlockType(2). // paragraph
			Text(larkdocx.NewTextBuilder().
				Elements([]*larkdocx.TextElement{textElem}).
				Build()).
			Build()
		req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
			DocumentId(docToken).
			BlockId(docToken).
			Body(larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
				Children([]*larkdocx.Block{block}).
				Index(-1).
				Build()).
			Build()
		resp, err := res.Client.Docx.DocumentBlockChildren.Create(context.Background(), req, res.RequestOptions()...)
		if err != nil {
			return fmt.Errorf("追加内容失败：%w", err)
		}
		if !resp.Success() {
			return fmt.Errorf("API 错误：%s（code %d）", resp.Msg, resp.Code)
		}
	}
	return nil
}
