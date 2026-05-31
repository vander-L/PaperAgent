package kowldege_index_pipeline

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
)

// newDocumentTransformer component initialization function of node 'RecursiveTransformer' in graph 'start'
func newDocumentTransformer(ctx context.Context) (tfr document.Transformer, err error) {
	config := &recursive.Config{
		ChunkSize:   1000,
		OverlapSize: 200,
		Separators:  []string{"\n\n", "\n", "。", ".", "?", "!", "；", ";", "，", ",", " "},
	}
	tfr, err = recursive.NewSplitter(ctx, config)
	if err != nil {
		return nil, err
	}
	return tfr, nil
}
