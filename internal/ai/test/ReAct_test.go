package test

import (
	"PaperAgent/config"
	"PaperAgent/internal/ai/agent/chat_pipeline"
	"PaperAgent/utility"
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestReAct(t *testing.T) {
	ctx := context.Background()
	config.Load()
	id := "111"
	userMsg := &chat_pipeline.UserMessage{
		ID:      id,
		Query:   "你好",
		History: utility.GetSimpleMemory(id).GetMessages(),
	}
	runner, err := chat_pipeline.BuildChatReact(ctx)
	if err != nil {
		panic(err)
	}

	out, err := runner.Invoke(ctx, userMsg)
	if err != nil {
		panic(err)
	}
	answer := out.Content
	fmt.Println("Q: 你好")
	fmt.Println("A:", answer)
	utility.GetSimpleMemory(id).SetMessages(schema.UserMessage("你好"))
	utility.GetSimpleMemory(id).SetMessages(schema.SystemMessage(out.Content))

	// 第二次对话
	userMsg = &chat_pipeline.UserMessage{
		ID:      id,
		Query:   "给我总结论文Track4World: Feedforward World-centric Dense 3D Tracking of All Pixels的内容",
		History: utility.GetSimpleMemory(id).GetMessages(),
	}
	out, err = runner.Invoke(ctx, userMsg)
	if err != nil {
		panic(err)
	}
	answer = out.Content
	fmt.Println("----------------")
	fmt.Println("Q: 给我总结论文Track4World: Feedforward World-centric Dense 3D Tracking of All Pixels的内容")
	fmt.Println("A:", answer)

	//// 第三次对话
	//userMsg = &chat_pipeline.UserMessage{
	//	ID:      id,
	//	Query:   "你在数据库中找3份三维重建相关的内容，2025年的，并说出属于哪篇论文，用中文回答，不联网找，使用内接的向量数据库",
	//	History: utility.GetSimpleMemory(id).GetMessages(),
	//}
	//out, err = runner.Invoke(ctx, userMsg)
	//if err != nil {
	//	panic(err)
	//}
	//answer = out.Content
	//fmt.Println("----------------")
	//fmt.Println("Q: 你在数据库中找3份三维重建相关的内容，2025年的，并说出属于哪篇论文，用中文回答，不联网找，使用内接的向量数据库")
	//fmt.Println("A:", answer)
}
