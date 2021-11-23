package control

import (
	"github.com/vckai/GoAnswer/app"
	"time"
)

const (
	EXPTIME = 3600 * 24
)

// 首页
func Index(context *app.Context) {
	var user string
	var userId int
	userId = int(time.Now().Unix())
	if context.Method == "POST" {
		user = context.String("user")
		context.Render("index", map[string]interface{}{
			"userName": user,
			"userId": userId,
		})
	}else {
		context.Render("login", map[string]interface{}{})
	}

}

