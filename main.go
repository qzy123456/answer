package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chaoyang/answer/app"
	"github.com/chaoyang/answer/control"
	"github.com/chaoyang/answer/server"
)

func main() {

	a := app.NewApp()

	a.Route("POST,GET", "/", control.Index)
	a.Get("/ws/", control.Ws)

	a.Static(func(context *app.Context) { //静态文件处理
		static := "public"

		url := strings.TrimPrefix(context.Url, "/")
		if url == "favicon.ico" {
			url = path.Join(static, url)
		}
		if !strings.HasPrefix(url, static) {
			return
		}

		f, e := os.Stat(url)
		if e == nil {
			if f.IsDir() {
				context.Status = 403
				context.End()
				return
			}
		}

		http.ServeFile(context.Response, context.Request, url)
		context.IsEnd = true
	})


	server.InitServer()
	a.Run()
}
