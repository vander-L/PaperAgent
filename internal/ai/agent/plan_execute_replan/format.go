package plan_execute_replan

import (
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func formatMessages(messages []adk.Message) string {
	var builder strings.Builder
	for _, msg := range messages {
		if msg == nil || msg.Content == "" {
			continue
		}
		builder.WriteString(msg.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}

func formatExecutedSteps(steps []planexecute.ExecutedStep) string {
	if len(steps) == 0 {
		return "暂无已执行步骤。"
	}

	var builder strings.Builder
	for _, step := range steps {
		builder.WriteString("执行步骤：")
		builder.WriteString(step.Step)
		builder.WriteString("\n关键输出：")
		builder.WriteString(step.Result)
		builder.WriteString("\n")
	}
	return builder.String()
}
