package plan_execute_replan

import (
	"PaperAgent/internal/ai/models"
	"PaperAgent/internal/ai/tools"
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func NewExecutor(ctx context.Context) (adk.Agent, error) {
	// log
	mcpTool, err := tools.GetLogMcpTool()
	if err != nil {
		return nil, err
	}
	toolList := mcpTool
	// alerts
	toolList = append(toolList, tools.NewPrometheusAlertsQueryTool())
	// file
	toolList = append(toolList, tools.NewQueryInternalDocsTool())
	// time
	toolList = append(toolList, tools.NewGetCurrentTimeTool())
	execModel, err := models.OpenAIForChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: execModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: toolList,
			},
		},
		MaxIterations: 999999,
		GenInputFn:    genExecutorInput,
	})
}

func genExecutorInput(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
	planContent, err := in.Plan.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return executorPromptTemplate().Format(ctx, map[string]any{
		"input":          formatMessages(in.UserInput),
		"plan":           string(planContent),
		"executed_steps": formatExecutedSteps(in.ExecutedSteps),
		"step":           in.Plan.FirstStep(),
	})
}

func executorPromptTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(`用户目标：{input}
当前计划：{plan}
已执行步骤及结果：{executed_steps}
请执行当前步骤：{step}`))
}

var systemPrompt = `1、首先，你需要根据用户的目标或问题，制定一个清晰的执行计划。将计划分解为一系列可独立执行的步骤，每个步骤应包含行动类型、所需信息和预期产出。使用以下格式输出计划：步骤1：[行动描述]；步骤2：[行动描述]；… 直到完成目标所需的所有步骤。
2、在制定计划后，依次执行每个步骤。对于当前步骤，输出“执行步骤X：[具体行动内容]”，然后根据该行动调用相应的工具或API获取结果。等待工具返回结果后，再继续下一步。
3、每执行完一个步骤，需总结该步骤的关键输出，并判断是否满足预期。如果发现计划需要调整（例如遇到错误、信息不足或环境变化），应暂停执行，重新评估并更新剩余步骤的计划。
4、当所有步骤执行完毕后，汇总各步骤的输出结果，最终回答用户的问题或完成用户的目标。
5、在整个过程中，保持对用户目标的跟踪，并确保每个行动都明确服务于最终目标。如果用户中途提出新指令，需立即停止当前计划，重新从步骤1开始制定新计划。`
