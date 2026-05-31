package embedder

import (
	"PaperAgent/config"
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
)

func GetEmbedding(ctx context.Context) (embedding.Embedder, error) {
	embCfg := config.AppConfig.LLM.Embedder
	dim := embCfg.Dimension
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     embCfg.APIKey,
		Model:      embCfg.Model,
		Dimensions: &dim,
	})
	if err != nil {
		log.Printf("new embedder error: %v\n", err)
		return nil, err
	}
	return embedder, nil
}
