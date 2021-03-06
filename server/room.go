package server

import (
	"fmt"
	"time"
)

// 房间相关信息
type room struct {
	Id     uint32 //房间ID
	Name   string //房间名称
	Time   int64  //房间创建时间
	Game *Game
}

// 创建房间
func NewRoom() (*room, error) {
	rid := newRoomId()
	r := &room{
		Id:     rid,
		Name:   fmt.Sprintf("room_%d", rid),
		Time:   time.Now().Unix(),
	}
	r.Game = NewGame(r)

	return r, nil
}

// 用户准备
func (r *room) userReady(userId int) error {
	return r.Game.userReady(userId)
}

// 添加一个玩家到房间中
func (r *room) addPlayer(userId int, arg ...bool) error {
	err := r.Game.addPlayer(userId)
	return err
}

func (r *room) Close() {
	r.Game.Close()
}
