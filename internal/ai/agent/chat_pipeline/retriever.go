package chat_pipeline

import (
	retriever2 "PaperAgent/internal/ai/retriever"
	"context"

	"github.com/cloudwego/eino/components/retriever"
)

// newRetriever component initialization function of node 'MilvusRetriever' in graph 'react'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	return retriever2.NewMilvusRetriever(ctx)
}
