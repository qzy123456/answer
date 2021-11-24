package server

import (
	"sync"
)

// 房间用户信息
type player struct {
	UserId int
	Win    bool  //用户是否胜利
	Count  int16 //答对题数
	lock    sync.Mutex
}

// 创建一个玩家
// userId  玩家ID
func NewPlayer(userId int) (*player, error) {
	return &player{
		UserId:  userId,
		Win:     false,
		Count:   0,
	}, nil
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

