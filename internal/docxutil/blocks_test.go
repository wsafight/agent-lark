package docxutil

import (
	"reflect"
	"testing"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

func TestExtractTextsFromBlock(t *testing.T) {
	t.Run("nil block", func(t *testing.T) {
		if got := ExtractTextsFromBlock(nil); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})

	t.Run("text block", func(t *testing.T) {
		blockType := 2
		block := &larkdocx.Block{
			BlockType: &blockType,
			Text: &larkdocx.Text{
				Elements: []*larkdocx.TextElement{
					makeTextElement("hello"),
					nil,
					{},
					makeTextElement(" world"),
				},
			},
		}

		got := ExtractTextsFromBlock(block)
		want := []string{"hello", " world"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected texts: got=%v want=%v", got, want)
		}
	})

	t.Run("unsupported block type", func(t *testing.T) {
		blockType := 31
		block := &larkdocx.Block{BlockType: &blockType}
		if got := ExtractTextsFromBlock(block); got != nil {
			t.Fatalf("expected nil for unsupported block type, got %#v", got)
		}
	})
}

func TestConvertBlocks(t *testing.T) {
	textType := 2
	bulletType := 10
	blockID1 := "blk_1"
	blockID2 := "blk_2"

	blocks := []*larkdocx.Block{
		nil,
		{
			BlockId:   &blockID1,
			BlockType: &textType,
			Text: &larkdocx.Text{
				Elements: []*larkdocx.TextElement{
					makeTextElement("A"),
					makeTextElement("B"),
				},
			},
		},
		{
			BlockId:   &blockID2,
			BlockType: &bulletType,
			Bullet: &larkdocx.Text{
				Elements: []*larkdocx.TextElement{
					makeTextElement("Item"),
				},
			},
		},
	}

	got := ConvertBlocks(blocks)
	if len(got) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(got))
	}

	if got[0].BlockID != blockID1 || got[0].BlockType != textType {
		t.Fatalf("unexpected first block metadata: %+v", got[0])
	}
	if !reflect.DeepEqual(got[0].Texts, []string{"A", "B"}) {
		t.Fatalf("unexpected first block texts: %v", got[0].Texts)
	}

	if got[1].BlockID != blockID2 || got[1].BlockType != bulletType {
		t.Fatalf("unexpected second block metadata: %+v", got[1])
	}
	if !reflect.DeepEqual(got[1].Texts, []string{"Item"}) {
		t.Fatalf("unexpected second block texts: %v", got[1].Texts)
	}
}

func makeTextElement(content string) *larkdocx.TextElement {
	return &larkdocx.TextElement{
		TextRun: &larkdocx.TextRun{
			Content: &content,
		},
	}
}
