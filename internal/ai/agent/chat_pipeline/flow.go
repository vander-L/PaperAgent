package chat_pipeline

import (
	"PaperAgent/internal/ai/tools"
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// newReactAgentLambda component initialization function of node 'ReActAgent' in graph 'react'
func newReactAgentLambda(ctx context.Context) (lba *compose.Lambda, err error) {
	// TODO Modify component configuration here.
	config := &react.AgentConfig{
		MaxStep:            25,
		ToolReturnDirectly: make(map[string]struct{}),
	}
	chatModelIns11, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	config.ToolCallingModel = chatModelIns11

	mcpTool, err := tools.GetLogMcpTool()
	if err != nil {
		return nil, err
	}
	config.ToolsConfig.Tools = mcpTool
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewGetCurrentTimeTool())
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewMysqlCrudTool())
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewQueryInternalDocsTool())

	ins, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}
