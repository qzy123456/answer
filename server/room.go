package server

import "github.com/gorilla/websocket"

// 房间相关信息
type room struct {
	Id     uint32 //房间ID
	Game *Game
}

// 创建房间
func NewRoom() (*room, error) {
	rid := newRoomId()
	r := &room{
		Id:     rid,
	}
	r.Game = NewGame(r)

	return r, nil
}

// 用户准备
func (r *room) userReady(userId int) error {
	return r.Game.userReady(userId)
}

// 添加一个玩家到房间中
func (r *room) addPlayer(userId int, conn *websocket.Conn) error {
	err := r.Game.addPlayer(userId,conn)
	return err
}

func (r *room) Close() {
	r.Game.Close()
}
