package server

import (
	"sync"
)

// 房间用户信息
type player struct {
	UserId int
	Count  int //答对题数
	Status bool  //用户准备状态(true准备中, false尚未准备)
	lock    sync.RWMutex
}

// 创建一个玩家
// userId  玩家ID
func NewPlayer(userId int) (*player, error) {
	return &player{
		UserId:  userId,
		Count:   0,
		Status:  false,
	}, nil
}

// 答对题数
func (p *player) count() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Count++
}

// 用户准备状态处理
func (p *player) ready(status bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Status = status
}

