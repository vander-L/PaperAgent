package chat_pipeline

import (
	"PaperAgent/internal/ai/models"
	"context"

	"github.com/cloudwego/eino/components/model"
)

func newChatModel(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	cm, err = models.OpenAIForChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
