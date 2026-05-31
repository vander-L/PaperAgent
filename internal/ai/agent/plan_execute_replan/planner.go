package plan_execute_replan

import (
	"PaperAgent/internal/ai/models"
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func NewPlanner(ctx context.Context) (adk.Agent, error) {
	planModel, err := models.OpenAIForChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: planModel,
		GenInputFn:           genPlannerInput,
	})
}

func genPlannerInput(ctx context.Context, userInput []adk.Message) ([]adk.Message, error) {
	return planPromptTemplate().Format(ctx, map[string]any{
		"input": userInput,
	})
}

func planPromptTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.MessagesPlaceholder("input", false),
	)
}
