package test

import (
	"PaperAgent/config"
	"PaperAgent/internal/ai/agent/kowldege_index_pipeline"
	"PaperAgent/utility"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

func TestPdf2Milvus(t *testing.T) {
	ctx := context.Background()
	if err := config.Load(); err != nil {
		panic("config Load failed: " + err.Error())
	}

	r, err := kowldege_index_pipeline.Buildstart(ctx)
	if err != nil {
		panic("11111" + err.Error())
	}
	err = filepath.WalkDir("../../../papers", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".pdf" {
			return nil
		}

		fmt.Println("[start] indexing file:", path)
		ids, err := r.Invoke(ctx, document.Source{URI: path},
			compose.WithCallbacks(utility.LogCallback(nil)))
		if err != nil {
			return fmt.Errorf("index %s: %w", path, err)
		}
		fmt.Println("[done] indexing file:", path, "len of parts", len(ids))
		return nil
	})
	if err != nil {
		panic(err)
	}
}
