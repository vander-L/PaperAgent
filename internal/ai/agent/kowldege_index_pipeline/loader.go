package kowldege_index_pipeline

import (
	"PaperAgent/internal/ai/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/ledongthuc/pdf"
)

// newLoader component initialization function of node 'FileLoader' in graph 'start'
func newLoader(ctx context.Context) (ldr document.Loader, err error) {
	config := &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      pdfParser{},
	}
	ldr, err = file.NewFileLoader(ctx, config)
	if err != nil {
		return nil, err
	}
	return ldr, nil
}

type pdfParser struct{}

type paperInfo struct {
	Title            string   `json:"title"`
	Authors          []string `json:"authors"`
	Abstract         string   `json:"abstract"`
	Keywords         []string `json:"keywords"`
	References       []string `json:"references"`
	InnovationPoints string   `json:"innovation_points"`
}

func (pdfParser) Parse(ctx context.Context, reader io.Reader, opts ...parser.Option) ([]*schema.Document, error) {
	options := parser.GetCommonOptions(&parser.Options{}, opts...)

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	pdfReader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read pdf failed: %w", err)
	}

	var text bytes.Buffer
	for pageIndex := 1; pageIndex <= pdfReader.NumPage(); pageIndex++ {
		page := pdfReader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return nil, fmt.Errorf("extract pdf page %d failed: %w", pageIndex, err)
		}

		text.WriteString(pageText)
		text.WriteString("\n")
	}

	fullText := text.String()
	metadata := cloneMetaData(options.ExtraMeta)
	paperInfo, err := extractPaperInfo(ctx, fullText)
	if err != nil {
		return nil, err
	}
	metadata["title"] = paperInfo.Title
	metadata["authors"] = paperInfo.Authors
	metadata["abstract"] = paperInfo.Abstract
	metadata["keywords"] = paperInfo.Keywords
	metadata["references"] = paperInfo.References
	metadata["innovation_points"] = paperInfo.InnovationPoints

	return []*schema.Document{
		{
			ID:       paperInfo.Title,
			Content:  fullText,
			MetaData: metadata,
		},
	}, nil
}

func cloneMetaData(metadata map[string]any) map[string]any {
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func extractPaperInfo(ctx context.Context, fullText string) (paperInfo, error) {
	cm, err := models.OpenAIForChatModel(ctx)
	if err != nil {
		return paperInfo{}, err
	}

	chunks := splitPaperText(fullText, 30000)
	partials := make([]paperInfo, 0, len(chunks))
	for _, chunk := range chunks {
		info, err := extractPaperInfoChunk(ctx, cm, chunk)
		if err != nil {
			return paperInfo{}, err
		}
		partials = append(partials, info)
	}

	info, err := summarizePaperInfo(ctx, cm, partials)
	if err != nil {
		return paperInfo{}, err
	}
	if info.Title == "" {
		return paperInfo{}, fmt.Errorf("extract paper title failed")
	}
	return info, nil
}

func extractPaperInfoChunk(ctx context.Context, cm interface {
	Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error)
}, chunk string) (paperInfo, error) {
	out, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(`你是论文信息抽取助手。请从用户提供的论文片段中抽取标题、作者、摘要、关键词、参考文献和创新点。只返回紧凑JSON，不要返回Markdown，不要解释。JSON格式：{"title":"论文标题","authors":["作者1","作者2"],"abstract":"摘要或片段概要","keywords":["关键词1","关键词2"],"references":["参考文献1","参考文献2"],"innovation_points":"创新点"}。每个文本字段控制在300字以内，关键词控制在10个以内，参考文献控制在20条以内。如果某项无法确定，返回空字符串或空数组。`),
		schema.UserMessage(chunk),
	}, model.WithMaxTokens(2048))
	if err != nil {
		return paperInfo{}, err
	}
	if out == nil {
		return paperInfo{}, fmt.Errorf("extract paper info chunk failed: empty model response")
	}
	return parsePaperInfo(out.Content)
}

func summarizePaperInfo(ctx context.Context, cm interface {
	Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error)
}, partials []paperInfo) (paperInfo, error) {
	data, err := json.Marshal(partials)
	if err != nil {
		return paperInfo{}, err
	}

	out, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(`你是论文信息汇总助手。请根据多个论文片段抽取结果，合并并总结出整篇论文的标题、作者、摘要、关键词、参考文献和创新点。只返回紧凑JSON，不要返回Markdown，不要解释。JSON格式：{"title":"论文标题","authors":["作者1","作者2"],"abstract":"摘要","keywords":["关键词1","关键词2"],"references":["参考文献1","参考文献2"],"innovation_points":"创新点总结"}。摘要、参考文献、创新点每项内容保持在300字以内，关键词控制在10个以内。`),
		schema.UserMessage(string(data)),
	}, model.WithMaxTokens(2048))
	if err != nil {
		return paperInfo{}, err
	}
	if out == nil {
		return paperInfo{}, fmt.Errorf("summarize paper info failed: empty model response")
	}
	return parsePaperInfo(out.Content)
}

func parsePaperInfo(content string) (paperInfo, error) {
	content = normalizeJSONContent(content)

	var info paperInfo
	if err := json.Unmarshal([]byte(content), &info); err != nil {
		return paperInfo{}, fmt.Errorf("parse paper info failed: %w, content: %.300s", err, content)
	}
	info.Title = strings.TrimSpace(info.Title)
	info.Abstract = strings.TrimSpace(info.Abstract)
	info.InnovationPoints = strings.TrimSpace(info.InnovationPoints)
	info.Authors = trimStringSlice(info.Authors)
	info.Keywords = trimStringSlice(info.Keywords)
	info.References = trimStringSlice(info.References)
	return info, nil
}

func normalizeJSONContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end >= start {
		return content[start : end+1]
	}
	return content
}

func splitPaperText(text string, chunkSize int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{""}
	}

	runes := []rune(text)
	chunks := make([]string, 0, len(runes)/chunkSize+1)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}

func trimStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
