package server

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ROW = 15  //行
	COL = 15  //列
)

type Game struct {
	rm        *room     //房间
	Users     []*player //房间用户列表
	Status    int8      //房间状态(0等待1游戏中2满人)
	Answer    chan string              //开始游戏chan
	GameStart chan bool                //chan
	Over      chan int                 //离开游戏chan
	Grid      *Grid                      //棋盘
	lock      sync.RWMutex
}

// new game
func NewGame(r *room) *Game {
	game := &Game{
		Status:    0,
		Answer:    make(chan string),
		GameStart: make(chan bool),
		Over:      make(chan int),
		rm:        r,
		Users:     make([]*player, 0),
	}
	//查询所有的题目
	game.Grid = InitGrid(ROW, COL, &Grid{})

	go func(game *Game) {
		for {
			fmt.Println("开始游戏了")
			if ok := <-game.GameStart; ok {
				//没题目就要新随机题目
				if game.Grid == nil {
					game.Grid = InitGrid(ROW, COL, &Grid{})
					fmt.Println("从新加载棋盘", game.Grid)
				}
				game.playGame()
			}
		}
	}(game)

	return game
}

// 添加一个玩家到房间中
func (game *Game) addPlayer(userId int) error {
	log.Printf("添加用户到房间中")
	var hand uint = 1
	//初始化一个玩家
	for _, user := range game.Users { //返回给客户端的用户信息
		if user.Hand == 1 {
			hand = 2
		}
	}
	player, err := NewPlayer(userId,hand)
	if err != nil {
		fmt.Println("创建游戏用户失败：", err)
		return err
	}
	log.Printf("在线player%v", player)
	//加入到房间列表
	game.Users = append(game.Users, player)

	var users []*onlineUser
	for _, user := range game.Users { //返回给客户端的用户信息
		log.Printf("在线puserr%v", user)
		u, err := h.GetOnlineUser(user.UserId)
		log.Printf("在线？%v", u)
		if err != nil {
			fmt.Println("获取在线用户失败：", err)
			continue
		}
		users = append(users, u)
	}
	log.Printf("在线列表%v", users)
	game.send("JoinRoom", map[string]interface{}{
		"Room": map[string]interface{}{
			"Id":     game.rm.Id,
			"Name":   game.rm.Name,
			"UserId": userId,
			"Status": game.Status,
			"Users":  game.Users,
		},
		"Users": users,
	})
	log.Printf("理论上发消息了")
	// 用户进入房间首次自动准备
	game.userReady(userId)

	return nil
}

// 用户准备，进房间自动准备
func (game *Game) userReady(userId int) error {
	isStart := false
	//进房间自动准备
	for _, user := range game.Users {
		//自己准备
		if userId == user.UserId {
			user.ready(true)
		} else {
			isStart = user.Status
		}
	}
	//判断人数
	if len(game.Users) >= 2 && isStart {
		isStart = true
	}
	u, _ := h.GetOnlineUser(userId)

	game.send("Ready", map[string]interface{}{
		"UserId":   userId,
		"UserName": u.UserName,
	})

	if isStart {
		game.GameStart <- true
	}
	return nil
}

// 判断用户是否准备状态,也就是房间够不够2人
func (game *Game) checkIsReady() bool {
	//进房间自动准备，不存在还需要准备
	if len(game.Users) != roomMaxUser {
		fmt.Println("人数不足", roomMaxUser, "人")
		return false
	}
	for _, user := range game.Users {
		if user.Status == false {
			fmt.Println("用户", user.UserId, "还未准备")
			return false
		}
	}
	return true
}

// 开始游戏
func (game *Game) playGame() {
	//检查是不是够2人
	if !game.checkIsReady() {
		return
	}

	if game.Status == 0 { // 开始一局游戏
		game.Status = 1 //更改房间游戏状态
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		uid := r.Intn(len(game.Users)) //随机抽选用户
		game.Users[uid].act(true)      //设置当前用户为答题用户
	}else { // 一局游戏中的继续游戏
		isAct := false
		for _, user := range game.Users {
			if user.IsAct == false && isAct == false { //当前答题用户
				user.act(true)
				isAct = true
			} else if user.IsAct == true {
				user.act(false)
			}
		}
	}
	//继续游戏
	for _, user := range game.Users {
		fmt.Println("游戏内的用户",user)
	}
	//发送消息正在玩游戏
	game.send("PlayGame", map[string]interface{}{
		"Users":    game.Users,
		"GameTime": playGameTime,
	})
	//20秒倒计时答题
	GameOutTime := playGameTime

	timer := time.NewTimer(GameOutTime * time.Second)
	for { //wait submit answer
		select {
		case answer := <-game.Answer:
			ans := strings.Split(answer, ",")
			userId, _ := strconv.Atoi(ans[0])
			x, _ := strconv.Atoi(ans[1])
			y, _ := strconv.Atoi(ans[2])
			fmt.Println("x轴", x)
			fmt.Println("用户id", userId)
			fmt.Println("y轴", y)

			game.submit(userId, x, y)
			return
		case userId := <-game.Over: //游戏结束
			fmt.Println("游戏结束")
			game.endGame(userId)
			return
		case <-timer.C: //游戏超时未提交答案，随机一个题？
			fmt.Println("超时未答题")
			//继续游戏，直到没题，或者人退出
			game.CheckWinner()
			return
		}
	}
}

// 中途退出房间
func (game *Game) GameOver(userId int) {
	fmt.Println("退出房间", userId)
	var delkey int
	for key, user := range game.Users {
		if user.UserId == userId {
			log.Printf("退出房间找到了%v", user.UserId)
			delkey = key
		} else {
			u, err := h.GetOnlineUser(userId)
			if err != nil {
				fmt.Println("获取用户失败：", err)
				continue
			}

			if game.Status == 1 { // 游戏中的状态，如果是正在游戏中，那么直接判断对方赢
				game.Over <- userId
				h.sendToClient("GameOver", user.UserId, map[string]interface{}{
					"OverUser":     userId,
					"OverUserName": u.Users.UserName,
				})
			} else {
				h.sendToClient("OutRoom", user.UserId, map[string]interface{}{
					"OverUser":     userId,
					"OverUserName": u.Users.UserName,
				})
			}
		}
	}
	if delkey >= 0 && game.Users != nil {
		fmt.Println("将用户", userId, "从房间", game.rm.Id, "中移除")
		game.Users = append(game.Users[:delkey], game.Users[delkey+1:]...)
		if len(game.Users) == 0 {
			game.Grid = InitGrid(ROW,COL,&Grid{})
		}
	}
}

// 落子
func (game *Game) submit(userId, x, y int) {
	log.Println(userId, x, y)
	if game.Status != 1 {
		fmt.Println("不在游戏中，请勿随便提交答案")
		return
	}

	//下发消息，谁答了，是否答对
	var hand1 uint
	for _, user := range game.Users {
		//答对，并且是自己，那么自己答对的题目+1，答错，对方+1
		if userId == user.UserId {
			hand1 = user.Hand
		}
	}
	//落子
	game.Grid.Set(x,y,hand(hand1))
	//根据房间的人数循环发
	game.send("GameResult", map[string]interface{}{
		"X":       x,
		"Y":       y,
		"UserId":  userId,
		"Hand":    hand1,
	})

    //判断谁赢了
    res := game.Grid.IsWin(x,y)

	if res > 0 { //结束一局游戏
	  fmt.Println("结束了？")
		game.endGame(userId)
	} else { //重新开始游戏
		game.playGame()
	}
}

// 游戏结束
func (game *Game) endGame(userId int) {

	game.send("EndGame", map[string]interface{}{
		"Users":  game.Users,
		"Winner": userId,
	})

	//清空
	game.clearGame()
}

// 游戏结束, 清空状态
func (game *Game) clearGame() {
	fmt.Println("清空房间信息")

	// 重置用户游戏状态
	for _, u := range game.Users {
		//u.Hand = 0
		u.Status = false
		u.IsAct = false
	}

	// 重置游戏状态,清空房间
	game.Grid = InitGrid(ROW,COL,&Grid{})
	game.Status = 0
}

// 关闭游戏, 清除数据
func (game *Game) Close() {
	close(game.Answer)
	close(game.GameStart)
	close(game.Over)
}

// 往房间用户的客户端发送消息
func (game *Game) send(action string, res map[string]interface{}) {
	for _, user := range game.Users {
		if user.UserId == 0 {
			continue
		}
		log.Printf("发消息%v", user)
		h.sendToClient(action, user.UserId, res)
	}
}

// 游戏结束
func (game *Game) CheckWinner(){
	for _, user := range game.Users {
		//
		if user.IsAct == false {
			game.endGame(user.UserId)
		}
	}
}
