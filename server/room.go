package server

import (
	"fmt"
	"sync"
)

// 房间相关信息
type room struct {
	Id   uint32 //房间ID
	Name string //房间名称
	Game *Game
	lock   sync.RWMutex
}

// 创建房间
func NewRoom() (*room, error) {
	rid := newRoomId()
	r := &room{
		Id:   rid,
		Name: fmt.Sprintf("room_%d", rid),
	}
	//新创建一个房间，就会初始化一个新游戏
	r.Game = NewGame(r)

	return r, nil
}

// 用户准备,更改状态为true
func (r *room) userReady(userId int) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Game.userReady(userId)
}

// 添加一个玩家到房间中
func (r *room) addPlayer(userId int) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Game.addPlayer(userId)
	return err
}

func (r *room) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Game.Close()
}
