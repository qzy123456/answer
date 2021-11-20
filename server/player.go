package server

import (
	"sync"
)

// 房间用户信息
type player struct {
	UserId int
	Status bool  //用户准备状态(true准备中, false尚未准备)
	Win    bool  //用户是否胜利
	Count  int16 //答对题数

	lock    sync.Mutex
}

// 创建一个玩家
// userId  玩家ID
func NewPlayer(userId int) (*player, error) {
	return &player{
		UserId:  userId,
		Status:  false,
		Win:     false,
		Count:   0,
	}, nil
}


// 用户准备状态处理
func (p *player) ready(status bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Status = status
}

// 答对题数
func (p *player) count() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Count++
}

// 是否胜利
func (p *player) win(isWin bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Win = isWin
}

