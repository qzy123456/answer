package control

import (
	"github.com/chaoyang/answer/app"
	"time"
)

// 首页
func Index(context *app.Context) {
	var userId int
	userId = int(time.Now().Unix())

	context.Render("index", map[string]interface{}{
		"userId":   userId,
		"Host": context.Host,
	})

}
