package server

import (
	"sync"
)

// 房间用户信息
type player struct {
	UserId int
	Hand   uint  //1：黑子 2：白子
	Status bool //用户准备状态(true准备中, false尚未准备)
	IsAct  bool //当前答题用户
	lock   sync.RWMutex
}

// 创建一个玩家
// userId  玩家ID
func NewPlayer(userId int,hand uint) (*player, error) {
	return &player{
		UserId: userId,
		Status: false,
		Hand:hand,
	}, nil
}

// 用户准备状态处理
func (p *player) ready(status bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Status = status
}

// 设置用户答题状态
func (p *player) act(act bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.IsAct = act
}