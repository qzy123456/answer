package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chaoyang/answer/app"
	"github.com/chaoyang/answer/control"
	"github.com/chaoyang/answer/model"
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
	model.NewModel(a.Config().MustValue("mysql_url", "127.0.0.1"),
		a.Config().MustValue("mysql_user", "root"),
		a.Config().MustValue("mysql_pass", "root"),
		a.Config().MustValue("mysql_port", "3306"),
		a.Config().MustValue("mysql_db", "examination"),
	) //连接到mysql

	server.InitServer()
	fmt.Println(model.GetAllExamId())
	a.Run()
}
