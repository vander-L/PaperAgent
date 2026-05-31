// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package chat

import (
	"context"

	"PaperAgent/api/chat/v1"
)

type IChatV1 interface {
	ChatReAct(ctx context.Context, req *v1.ChatReActReq) (res *v1.ChatReActRes, err error)
	ChatReActStream(ctx context.Context, req *v1.ChatReActStreamReq) (res *v1.ChatReActStreamRes, err error)
	FileUpload(ctx context.Context, req *v1.FileUploadReq) (res *v1.FileUploadRes, err error)
	ChatPlanExecute(ctx context.Context, req *v1.ChatPlanExecuteReq) (res *v1.ChatPlanExecuteRes, err error)
	ChatPlanExecuteStream(ctx context.Context, req *v1.ChatPlanExecuteStreamReq) (res *v1.ChatPlanExecuteStreamRes, err error)
}
