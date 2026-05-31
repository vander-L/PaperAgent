package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ChatReActReq struct {
	g.Meta   `path:"/chat" method:"post" summary:"ReAct对话"`
	Id       string
	Question string
}

type ChatReActRes struct {
	Answer string `json:"answer"`
}

type ChatReActStreamReq struct {
	g.Meta   `path:"/chat_stream" method:"post" summary:"ReAct流式对话"`
	Id       string
	Question string
}

type ChatReActStreamRes struct {
}

type FileUploadReq struct {
	g.Meta `path:"/upload" method:"post" mime:"multipart/form-data" summary:"文件上传"`
}

type FileUploadRes struct {
	FileName string `json:"fileName" dc:"保存的文件名"`
	FilePath string `json:"filePath" dc:"文件保存路径"`
	FileSize int64  `json:"fileSize" dc:"文件大小(字节)"`
}

type ChatPlanExecuteReq struct {
	g.Meta   `path:"/chat_p" method:"post" summary:"plan-execute对话"`
	Id       string
	Question string
}

type ChatPlanExecuteRes struct {
	Result string   `json:"result"`
	Detail []string `json:"detail"`
}

type ChatPlanExecuteStreamReq struct {
	g.Meta   `path:"/chat_p_stream" method:"post" summary:"plan-execute流式对话"`
	Id       string
	Question string
}

type ChatPlanExecuteStreamRes struct {
}
