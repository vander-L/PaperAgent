package indexer

import (
	"PaperAgent/internal/ai/embedder"
	"PaperAgent/utility"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

type batchEmbedder struct {
	embedding.Embedder
	BatchSize int
}

func (b batchEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if b.BatchSize <= 0 {
		b.BatchSize = 10
	}

	vectors := make([][]float64, 0, len(texts))
	for start := 0; start < len(texts); start += b.BatchSize {
		end := start + b.BatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batchVectors, err := b.Embedder.EmbedStrings(ctx, texts[start:end], opts...)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, batchVectors...)
	}

	return vectors, nil
}

func GetMilvusIndexer(ctx context.Context) (*milvus.Indexer, error) {
	cli, err := utility.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	eb, err := embedder.GetEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            cli,
		Collection:        utility.MilvusCollectionName,
		Fields:            utility.Fields,
		Embedding:         batchEmbedder{Embedder: eb, BatchSize: 10},
		DocumentConverter: convertPaperDocuments,
	})
	if err != nil {
		return nil, err
	}
	return indexer, nil
}

type paperRow struct {
	ID               string    `json:"id" milvus:"name:id"`
	PaperID          string    `json:"paper_id" milvus:"name:paper_id"`
	Title            string    `json:"title" milvus:"name:title"`
	Authors          string    `json:"authors" milvus:"name:authors"`
	Abstract         string    `json:"abstract" milvus:"name:abstract"`
	TitleVector      []float32 `json:"title_vector" milvus:"name:title_vector"`
	AbstractVector   []float32 `json:"abstract_vector" milvus:"name:abstract_vector"`
	TopicVector      []float32 `json:"topic_vector" milvus:"name:topic_vector"`
	Keywords         [][]byte  `json:"keywords" milvus:"name:keywords"`
	PublicationYear  int32     `json:"publication_year" milvus:"name:publication_year"`
	CitationCount    int32     `json:"citation_count" milvus:"name:citation_count"`
	Venue            string    `json:"venue" milvus:"name:venue"`
	FullText         string    `json:"full_text" milvus:"name:full_text"`
	FullTextVector   []float32 `json:"full_text_vector" milvus:"name:full_text_vector"`
	References       string    `json:"references" milvus:"name:references"`
	Methodology      string    `json:"methodology" milvus:"name:methodology"`
	DatasetInfo      string    `json:"dataset_info" milvus:"name:dataset_info"`
	InnovationPoints string    `json:"innovation_points" milvus:"name:innovation_points"`
	Field            string    `json:"field" milvus:"name:field"`
	Metadata         []byte    `json:"metadata" milvus:"name:metadata"`
}

func convertPaperDocuments(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	rows := make([]interface{}, 0, len(docs))

	for i, doc := range docs {
		metadata, err := json.Marshal(doc.MetaData)
		if err != nil {
			return nil, err
		}

		paperTitle := getPaperTitle(doc)
		abstract := getMetaString(doc, "abstract")
		keywordsText := strings.Join(getMetaStringSlice(doc, "keywords"), " ")
		titleVector, abstractVector, topicVector, err := embedMetadataVectors(ctx, paperTitle, abstract, keywordsText, len(vectors[i]))
		if err != nil {
			return nil, err
		}

		row := &paperRow{
			ID:               fmt.Sprintf("%s_%d", paperTitle, i),
			PaperID:          paperTitle,
			Title:            paperTitle,
			Authors:          getMetaJSONString(doc, "authors", "[]"),
			Abstract:         abstract,
			TitleVector:      titleVector,
			AbstractVector:   abstractVector,
			TopicVector:      topicVector,
			Keywords:         getMetaVarCharArray(doc, "keywords"),
			PublicationYear:  getMetaInt32(doc, "publication_year"),
			CitationCount:    getMetaInt32(doc, "citation_count"),
			Venue:            getMetaString(doc, "venue"),
			FullText:         doc.Content,
			FullTextVector:   float64ToFloat32(vectors[i]),
			References:       getMetaJSONString(doc, "references", "[]"),
			Methodology:      getMetaString(doc, "methodology"),
			DatasetInfo:      getMetaString(doc, "dataset_info"),
			InnovationPoints: getMetaString(doc, "innovation_points"),
			Field:            getMetaString(doc, "field"),
			Metadata:         metadata,
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func embedMetadataVectors(ctx context.Context, title, abstract, topic string, dim int) ([]float32, []float32, []float32, error) {
	eb, err := embedder.GetEmbedding(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	texts := []string{title, abstract, topic}
	vecs, err := batchEmbedder{Embedder: eb, BatchSize: 10}.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(vecs) != len(texts) {
		return nil, nil, nil, fmt.Errorf("metadata embedding count mismatch: expected %d, got %d", len(texts), len(vecs))
	}

	return vectorOrZero(vecs[0], dim), vectorOrZero(vecs[1], dim), vectorOrZero(vecs[2], dim), nil
}

func vectorOrZero(values []float64, dim int) []float32 {
	if len(values) == 0 {
		return zeroVector(dim)
	}
	return float64ToFloat32(values)
}

func getPaperTitle(doc *schema.Document) string {
	title := strings.TrimSpace(getMetaString(doc, "title"))
	if title != "" {
		return title
	}

	name := strings.TrimSpace(getMetaString(doc, "_file_name"))
	if name == "" {
		name = filepath.Base(doc.ID)
	}
	title = strings.TrimSuffix(name, filepath.Ext(name))
	if title == "" {
		return doc.ID
	}
	return title
}

func getMetaString(doc *schema.Document, key string) string {
	if doc.MetaData == nil {
		return ""
	}

	value, ok := doc.MetaData[key]
	if !ok || value == nil {
		return ""
	}

	if s, ok := value.(string); ok {
		return s
	}

	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(data)
}

func getMetaJSONString(doc *schema.Document, key string, defaultValue string) string {
	if doc.MetaData == nil {
		return defaultValue
	}

	value, ok := doc.MetaData[key]
	if !ok || value == nil {
		return defaultValue
	}

	if s, ok := value.(string); ok {
		return s
	}

	data, err := json.Marshal(value)
	if err != nil {
		return defaultValue
	}

	return string(data)
}

func getMetaVarCharArray(doc *schema.Document, key string) [][]byte {
	values := getMetaStringSlice(doc, key)
	result := make([][]byte, 0, len(values))
	for _, value := range values {
		result = append(result, []byte(value))
	}
	return result
}

func getMetaStringSlice(doc *schema.Document, key string) []string {
	if doc.MetaData == nil {
		return []string{}
	}

	value, ok := doc.MetaData[key]
	if !ok || value == nil {
		return []string{}
	}

	switch v := value.(type) {
	case []string:
		return v
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case string:
		if v == "" {
			return []string{}
		}
		return []string{v}
	default:
		return []string{}
	}
}

func getMetaInt32(doc *schema.Document, key string) int32 {
	if doc.MetaData == nil {
		return 0
	}

	value, ok := doc.MetaData[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	case float32:
		return int32(v)
	default:
		return 0
	}
}

func float64ToFloat32(values []float64) []float32 {
	result := make([]float32, len(values))
	for i, value := range values {
		result[i] = float32(value)
	}
	return result
}

func zeroVector(dim int) []float32 {
	return make([]float32, dim)
}
