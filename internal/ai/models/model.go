package models

import (
	"PaperAgent/config"
	"context"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

func OpenAIForChatModel(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	chatModelCfg := config.AppConfig.LLM.Chat
	cm, err = ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  chatModelCfg.APIKey,
		BaseURL: chatModelCfg.BaseURL,
		Model:   chatModelCfg.Model,
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}
