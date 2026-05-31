package plan_execute_replan

import (
	"PaperAgent/internal/ai/models"
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func NewRePlanAgent(ctx context.Context) (adk.Agent, error) {
	model, err := models.OpenAIForChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel:  model,
		GenInputFn: genReplannerInput,
	})
}

func genReplannerInput(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
	planContent, err := in.Plan.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return replannerPromptTemplate().Format(ctx, map[string]any{
		"input":          formatMessages(in.UserInput),
		"plan":           string(planContent),
		"executed_steps": formatExecutedSteps(in.ExecutedSteps),
	})
}

func replannerPromptTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(`用户目标：{input}
原始计划：{plan}
已执行步骤及结果：{executed_steps}
请判断目标是否已经完成。若已完成，调用 respond 返回最终答案；若未完成，调用 plan 生成仅包含剩余步骤的新计划。`),
	)
}
