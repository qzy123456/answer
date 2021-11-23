package control

import (
	"fmt"
	"github.com/vckai/GoAnswer/app"
	"github.com/vckai/GoAnswer/server"

	"github.com/gorilla/websocket"
)

// 建立websocket连接
func Ws(context *app.Context) {

	nUserId := 1
	ws, err := websocket.Upgrade(context.Response, context.Request, nil, 1024, 1024)
	if errmsg, ok := err.(websocket.HandshakeError); ok {
		fmt.Println("websocket连接失败，---ERROR：", errmsg)
		context.Throw(403, "Websocket Not handler")
		return
	} else if err != nil {
		fmt.Println("Websocket连接失败，---ERROR：", err)
		context.Throw(403, "WEBSOCKET连接失败")
		return
	}

	c := server.NewConn(nUserId, ws)

	go c.WritePump()
	c.ReadPump()
}


