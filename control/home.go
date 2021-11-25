package control

import (
	"github.com/vckai/GoAnswer/app"
	"time"
)

// 首页
func Index(context *app.Context) {
	var user string
	var userId int
	userId = int(time.Now().Unix())

	user = context.String("user")
	context.Render("index", map[string]interface{}{
		"userName": user,
		"userId":   userId,
		"Host": context.Host,
	})

}

// 首页
func Login(context *app.Context) {
	context.Render("login", map[string]interface{}{})
}
