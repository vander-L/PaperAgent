package kowldege_index_pipeline

import (
	"context"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

func Buildstart(ctx context.Context) (r compose.Runnable[document.Source, []string], err error) {
	const (
		FileLoader           = "FileLoader"
		RecursiveTransformer = "RecursiveTransformer"
		Indexer              = "Indexer"
	)
	g := compose.NewGraph[document.Source, []string]()
	fileLoaderKeyOfLoader, err := newLoader(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)
	recursiveTransformerKeyOfDocumentTransformer, err := newDocumentTransformer(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddDocumentTransformerNode(RecursiveTransformer, recursiveTransformerKeyOfDocumentTransformer)
	indexerKeyOfIndexer, err := newIndexer(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddIndexerNode(Indexer, indexerKeyOfIndexer)
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, RecursiveTransformer)
	_ = g.AddEdge(RecursiveTransformer, Indexer)
	_ = g.AddEdge(Indexer, compose.END)
	r, err = g.Compile(ctx, compose.WithGraphName("start"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, nil
}
