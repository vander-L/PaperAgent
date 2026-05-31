package retriever

import (
	"PaperAgent/internal/ai/embedder"
	"PaperAgent/utility"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func NewMilvusRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	cli, err := utility.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	eb, err := embedder.GetEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	sp, err := entity.NewIndexAUTOINDEXSearchParam(1)
	if err != nil {
		return nil, err
	}

	return milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:            cli,
		Collection:        utility.MilvusCollectionName,
		VectorField:       "full_text_vector",
		OutputFields:      paperOutputFields(),
		Embedding:         eb,
		MetricType:        entity.COSINE,
		TopK:              5,
		Sp:                sp,
		VectorConverter:   floatVectorConverter,
		DocumentConverter: paperDocumentConverter,
	})
}

func paperOutputFields() []string {
	return []string{
		"id",
		"paper_id",
		"title",
		"authors",
		"abstract",
		"keywords",
		"publication_year",
		"citation_count",
		"venue",
		"full_text",
		"references",
		"methodology",
		"dataset_info",
		"innovation_points",
		"field",
		"metadata",
	}
}

func floatVectorConverter(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
	result := make([]entity.Vector, 0, len(vectors))
	for _, vector := range vectors {
		values := make([]float32, len(vector))
		for i, value := range vector {
			values[i] = float32(value)
		}
		result = append(result, entity.FloatVector(values))
	}
	return result, nil
}

func paperDocumentConverter(ctx context.Context, result client.SearchResult) ([]*schema.Document, error) {
	docs := make([]*schema.Document, result.IDs.Len())
	for i := range docs {
		id, err := result.IDs.GetAsString(i)
		if err != nil {
			return nil, fmt.Errorf("get id: %w", err)
		}

		docs[i] = &schema.Document{
			ID:       id,
			MetaData: map[string]any{},
		}
	}

	for _, field := range result.Fields {
		switch field.Name() {
		case "full_text":
			for i, doc := range docs {
				content, err := field.GetAsString(i)
				if err != nil {
					return nil, fmt.Errorf("get full_text: %w", err)
				}
				doc.Content = content
			}
		case "metadata":
			for i, doc := range docs {
				value, err := field.Get(i)
				if err != nil {
					return nil, fmt.Errorf("get metadata: %w", err)
				}

				data, ok := value.([]byte)
				if !ok {
					return nil, fmt.Errorf("metadata type is %T, want []byte", value)
				}

				if len(data) == 0 {
					continue
				}

				if err := json.Unmarshal(data, &doc.MetaData); err != nil {
					return nil, fmt.Errorf("unmarshal metadata: %w", err)
				}
			}
		default:
			for i, doc := range docs {
				value, err := field.Get(i)
				if err != nil {
					return nil, fmt.Errorf("get %s: %w", field.Name(), err)
				}
				doc.MetaData[field.Name()] = value
			}
		}
	}

	return docs, nil
}
