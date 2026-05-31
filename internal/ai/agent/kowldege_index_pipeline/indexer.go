package kowldege_index_pipeline

import (
	indexer2 "PaperAgent/internal/ai/indexer"
	"context"

	"github.com/cloudwego/eino/components/indexer"
)

// newIndexer component initialization function of node 'Indexer' in graph 'start'
func newIndexer(ctx context.Context) (idr indexer.Indexer, err error) {
	return indexer2.GetMilvusIndexer(ctx)
}
