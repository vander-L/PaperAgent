package plan_execute_replan

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-examples/adk/common/prints"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func NewPlanExecuteAgent(ctx context.Context) (adk.ResumableAgent, error) {
	planAgent, err := NewPlanner(ctx)
	if err != nil {
		return nil, err
	}
	executeAgent, err := NewExecutor(ctx)
	if err != nil {
		return nil, err
	}
	replanAgent, err := NewRePlanAgent(ctx)
	if err != nil {
		return nil, err
	}
	planExecuteAgent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planAgent,
		Executor:      executeAgent,
		Replanner:     replanAgent,
		MaxIterations: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("build PlanExecuteAgent Error: %v", err)
	}
	return planExecuteAgent, nil
}

func BuildPlanAgent(ctx context.Context, query string) (string, []string, error) {
	planExecuteAgent, err := NewPlanExecuteAgent(ctx)
	if err != nil {
		return "", []string{}, err
	}
	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: planExecuteAgent,
	})
	iter := r.Query(ctx, query)
	var lastMessage adk.Message
	var detail []string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Println("------------- Event -------------")
		prints.Event(event)
		if event.Output != nil {
			lastMessage, _, err = adk.GetMessage(event)
			detail = append(detail, lastMessage.String())
		}
	}
	if lastMessage == nil {
		return "", []string{}, fmt.Errorf("get lastMessage Error")
	}
	return lastMessage.Content, detail, nil
}
