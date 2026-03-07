package docxutil

import (
	"context"
	"fmt"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

// FetchAllBlocks fetches all document blocks with pagination.
func FetchAllBlocks(ctx context.Context, c *client.Result, docToken string) ([]*larkdocx.Block, error) {
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
			ctx,
			builder.Build(),
			c.RequestOptions()...,
		)
		if err != nil {
			return nil, fmt.Errorf("API_ERROR：%s", err.Error())
		}
		if !resp.Success() {
			return nil, fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
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

	return allBlocks, nil
}

// ConvertBlocks converts larkdocx blocks to output.BlockItem.
func ConvertBlocks(blocks []*larkdocx.Block) []*output.BlockItem {
	result := make([]*output.BlockItem, 0, len(blocks))
	for _, b := range blocks {
		if b == nil {
			continue
		}
		item := &output.BlockItem{}
		if b.BlockId != nil {
			item.BlockID = *b.BlockId
		}
		if b.BlockType != nil {
			item.BlockType = *b.BlockType
		}
		item.Texts = ExtractTextsFromBlock(b)
		result = append(result, item)
	}
	return result
}

// ExtractTextsFromBlock extracts plain text fragments from a docx block.
func ExtractTextsFromBlock(b *larkdocx.Block) []string {
	if b == nil || b.BlockType == nil {
		return nil
	}

	switch *b.BlockType {
	case 2: // text
		if b.Text != nil && b.Text.Elements != nil {
			return extractFromElements(b.Text.Elements)
		}
	case 3: // heading1
		if b.Heading1 != nil && b.Heading1.Elements != nil {
			return extractFromElements(b.Heading1.Elements)
		}
	case 4: // heading2
		if b.Heading2 != nil && b.Heading2.Elements != nil {
			return extractFromElements(b.Heading2.Elements)
		}
	case 5: // heading3
		if b.Heading3 != nil && b.Heading3.Elements != nil {
			return extractFromElements(b.Heading3.Elements)
		}
	case 6: // heading4
		if b.Heading4 != nil && b.Heading4.Elements != nil {
			return extractFromElements(b.Heading4.Elements)
		}
	case 7: // heading5
		if b.Heading5 != nil && b.Heading5.Elements != nil {
			return extractFromElements(b.Heading5.Elements)
		}
	case 8: // heading6
		if b.Heading6 != nil && b.Heading6.Elements != nil {
			return extractFromElements(b.Heading6.Elements)
		}
	case 9: // ordered_list
		if b.Ordered != nil && b.Ordered.Elements != nil {
			return extractFromElements(b.Ordered.Elements)
		}
	case 10: // bullet
		if b.Bullet != nil && b.Bullet.Elements != nil {
			return extractFromElements(b.Bullet.Elements)
		}
	case 11: // code
		if b.Code != nil && b.Code.Elements != nil {
			return extractFromElements(b.Code.Elements)
		}
	case 12: // quote
		if b.Quote != nil && b.Quote.Elements != nil {
			return extractFromElements(b.Quote.Elements)
		}
	}
	return nil
}

func extractFromElements(elements []*larkdocx.TextElement) []string {
	texts := make([]string, 0, len(elements))
	for _, el := range elements {
		if el == nil || el.TextRun == nil || el.TextRun.Content == nil {
			continue
		}
		texts = append(texts, *el.TextRun.Content)
	}
	return texts
}
