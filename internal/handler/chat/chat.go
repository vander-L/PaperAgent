package chat

import (
	"PaperAgent/api/chat"
	"PaperAgent/api/chat/v1"
	"PaperAgent/internal/ai/agent/chat_pipeline"
	"PaperAgent/internal/ai/agent/kowldege_index_pipeline"
	"PaperAgent/internal/ai/agent/plan_execute_replan"
	"PaperAgent/internal/logic/sse"
	"PaperAgent/utility"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

type ControllerV1 struct {
	service *sse.SSEService
}

func NewV1() chat.IChatV1 {
	return &ControllerV1{
		service: sse.NewSSEClient(),
	}
}

func (c ControllerV1) ChatReAct(ctx context.Context, req *v1.ChatReActReq) (res *v1.ChatReActRes, err error) {
	id := req.Id
	msg := req.Question
	userMessage := &chat_pipeline.UserMessage{
		ID:      id,
		Query:   msg,
		History: utility.GetSimpleMemory(id).GetMessages(),
	}

	runner, err := chat_pipeline.BuildChatReact(ctx)
	if err != nil {
		return nil, err
	}

	out, err := runner.Invoke(ctx, userMessage, compose.WithCallbacks(utility.LogCallback(nil)))
	if err != nil {
		return nil, err
	}
	res = &v1.ChatReActRes{
		Answer: out.Content,
	}
	utility.GetSimpleMemory(id).SetMessages(schema.UserMessage(msg))
	utility.GetSimpleMemory(id).SetMessages(schema.SystemMessage(out.Content))

	return res, nil
}

func (c ControllerV1) ChatReActStream(ctx context.Context, req *v1.ChatReActStreamReq) (res *v1.ChatReActStreamRes, err error) {
	id := req.Id
	msg := req.Question

	ctx = context.WithValue(ctx, "client_id", req.Id)
	client, err := c.service.Create(ctx, g.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	userMessage := &chat_pipeline.UserMessage{
		ID:      id,
		Query:   msg,
		History: utility.GetSimpleMemory(id).GetMessages(),
	}

	runner, err := chat_pipeline.BuildChatReact(ctx)
	sr, err := runner.Stream(ctx, userMessage, compose.WithCallbacks(utility.LogCallback(nil)))
	if err != nil {
		client.SendToClient("error", err.Error())
		return nil, err
	}
	defer sr.Close()

	var fullResponse strings.Builder

	defer func() {
		completeResponse := fullResponse.String()
		if completeResponse != "" {
			utility.GetSimpleMemory(id).SetMessages(schema.UserMessage(msg))
			utility.GetSimpleMemory(id).SetMessages(schema.SystemMessage(completeResponse))
		}
	}()

	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			client.SendToClient("done", "Stream completed")
			return &v1.ChatReActStreamRes{}, nil
		}
		if err != nil {
			client.SendToClient("error", err.Error())
			return &v1.ChatReActStreamRes{}, nil
		}
		fullResponse.WriteString(chunk.Content)
		client.SendToClient("message", chunk.Content)
	}
}

func (c ControllerV1) FileUpload(ctx context.Context, req *v1.FileUploadReq) (res *v1.FileUploadRes, err error) {
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	if uploadFile == nil {
		return nil, gerror.New("请上传文件")
	}

	// 确保保存目录存在
	if !gfile.Exists(utility.FileDir) {
		if err := gfile.Mkdir(utility.FileDir); err != nil {
			return nil, gerror.Wrapf(err, "创建目录失败: %s", utility.FileDir)
		}
	}

	newFileName := filepath.Base(uploadFile.Filename)
	if !strings.EqualFold(filepath.Ext(newFileName), ".pdf") {
		return nil, gerror.New("仅支持上传PDF文件")
	}

	saveDir := filepath.Clean(utility.FileDir)
	_, err = uploadFile.Save(saveDir, false)
	if err != nil {
		return nil, gerror.Wrapf(err, "保存文件失败")
	}

	savePath := filepath.Join(saveDir, newFileName)
	fileInfo, err := os.Stat(savePath)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取文件信息失败")
	}

	err = buildIntoIndex(ctx, savePath)
	if err != nil {
		return nil, gerror.Wrapf(err, "构建知识库失败")
	}

	res = &v1.FileUploadRes{
		FileName: newFileName,
		FilePath: savePath,
		FileSize: fileInfo.Size(),
	}
	return res, nil
}

func (c ControllerV1) ChatPlanExecute(ctx context.Context, req *v1.ChatPlanExecuteReq) (res *v1.ChatPlanExecuteRes, err error) {
	id := req.Id
	msg := buildPlanExecuteQuestion(id, req.Question)

	result, detail, err := plan_execute_replan.BuildPlanAgent(ctx, msg)
	if err != nil {
		return nil, err
	}

	utility.GetSimpleMemory(id).SetMessages(schema.UserMessage(req.Question))
	utility.GetSimpleMemory(id).SetMessages(schema.SystemMessage(result))

	return &v1.ChatPlanExecuteRes{
		Result: result,
		Detail: detail,
	}, nil
}

func (c ControllerV1) ChatPlanExecuteStream(ctx context.Context, req *v1.ChatPlanExecuteStreamReq) (res *v1.ChatPlanExecuteStreamRes, err error) {
	id := req.Id
	msg := buildPlanExecuteQuestion(id, req.Question)

	ctx = context.WithValue(ctx, "client_id", req.Id)
	client, err := c.service.Create(ctx, g.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	result, detail, err := plan_execute_replan.BuildPlanAgent(ctx, msg)
	if err != nil {
		client.SendToClient("error", err.Error())
		return nil, err
	}

	for _, item := range detail {
		client.SendToClient("message", item)
	}
	client.SendToClient("done", "Stream completed")

	utility.GetSimpleMemory(id).SetMessages(schema.UserMessage(req.Question))
	utility.GetSimpleMemory(id).SetMessages(schema.SystemMessage(result))

	return &v1.ChatPlanExecuteStreamRes{}, nil
}

func buildPlanExecuteQuestion(id, question string) string {
	history := utility.GetSimpleMemory(id).GetMessages()
	if len(history) == 0 {
		return question
	}

	var builder strings.Builder
	builder.WriteString("历史对话:\n")
	for _, msg := range history {
		if msg == nil || msg.Content == "" {
			continue
		}
		builder.WriteString(string(msg.Role))
		builder.WriteString(": ")
		builder.WriteString(msg.Content)
		builder.WriteString("\n")
	}
	builder.WriteString("当前问题:\n")
	builder.WriteString(question)
	return builder.String()
}

func buildIntoIndex(ctx context.Context, path string) error {
	if path == "" {
		return gerror.New("文件路径不能为空")
	}

	r, err := kowldege_index_pipeline.Buildstart(ctx)
	if err != nil {
		return gerror.Wrap(err, "构建索引流水线失败")
	}

	_, err = r.Invoke(ctx, document.Source{URI: path}, compose.WithCallbacks(utility.LogCallback(nil)))
	if err != nil {
		return gerror.Wrapf(err, "索引文件失败: %s", path)
	}

	return nil
}
