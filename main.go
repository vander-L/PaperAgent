package main

import (
	"PaperAgent/config"
	"PaperAgent/internal/handler/chat"
	"PaperAgent/utility"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	fileDir, err := g.Cfg().Get(ctx, "file_dir")
	if err != nil {
		panic(err)
	}
	if fileDir.String() != "" {
		utility.FileDir = fileDir.String()
	}
	err = config.Load()
	if err != nil {
		panic("config Load failed: " + err.Error())
	}
	fmt.Println(config.AppConfig)
	s := g.Server()
	s.SetClientMaxBodySize(200 * 1024 * 1024)
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(utility.CORSMiddleware)
		group.Middleware(utility.ResponseMiddleware)
		group.Bind(chat.NewV1())
	})
	s.SetPort(6872)
	s.Run()
}
