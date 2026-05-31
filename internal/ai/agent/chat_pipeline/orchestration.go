package chat_pipeline

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func BuildChatReact(ctx context.Context) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	const (
		InputToChat     = "InputToChat"
		InputToRag      = "InputToRag"
		MilvusRetriever = "MilvusRetriever"
		ChatTemplate    = "ChatTemplate"
		ReActAgent      = "ReActAgent"
	)
	g := compose.NewGraph[*UserMessage, *schema.Message]()
	_ = g.AddLambdaNode(InputToRag, compose.InvokableLambdaWithOption(newInputToRagLambda), compose.WithNodeName("UserMessageToQuery"))
	chatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)
	reActAgentKeyOfLambda, err := newReactAgentLambda(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(ReActAgent, reActAgentKeyOfLambda, compose.WithNodeName("ReActAgent"))
	milvusRetrieverKeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}

	_ = g.AddRetrieverNode(MilvusRetriever, milvusRetrieverKeyOfRetriever, compose.WithOutputKey("documents"))
	_ = g.AddLambdaNode(InputToChat, compose.InvokableLambdaWithOption(newInputToChatLambda), compose.WithNodeName("UserMsgToChat"))
	_ = g.AddEdge(compose.START, InputToChat)
	_ = g.AddEdge(compose.START, InputToRag)
	_ = g.AddEdge(ReActAgent, compose.END)
	_ = g.AddEdge(InputToChat, ChatTemplate)
	_ = g.AddEdge(InputToRag, MilvusRetriever)
	_ = g.AddEdge(MilvusRetriever, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ReActAgent)
	r, err = g.Compile(ctx, compose.WithGraphName("ChatAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
