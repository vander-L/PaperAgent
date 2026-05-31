package test

import (
	"PaperAgent/config"
	"PaperAgent/internal/ai/retriever"
	"context"
	"fmt"
	"testing"
)

func TestRecall(t *testing.T) {
	ctx := context.Background()
	if err := config.Load(); err != nil {
		panic("config Load failed: " + err.Error())
	}
	r, err := retriever.NewMilvusRetriever(ctx)
	if err != nil {
		panic(err)
	}
	query := "如何将高斯点转换为mesh网格的？"
	docs, err := r.Retrieve(ctx, query)
	if err != nil {
		panic(err)
	}
	fmt.Println("Q：", query)
	for _, doc := range docs {
		fmt.Println("A：", doc.Content)
	}
	fmt.Println("Done", len(docs))
}
