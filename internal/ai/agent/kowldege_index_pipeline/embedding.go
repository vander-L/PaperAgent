package kowldege_index_pipeline

import (
	"PaperAgent/internal/ai/embedder"
	"context"

	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	return embedder.GetEmbedding(ctx)
}
