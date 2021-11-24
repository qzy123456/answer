package server

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/vckai/GoAnswer/model"
)

var (
	ErrNotExam = errors.New("已经完成了所有题目")
)

type Game struct {
	rm     *room     //房间
	Users  []*player //房间用户列表
	Status int8      //房间状态(0等待1游戏中2满人)

	exam      map[int]model.KsQuestion //当前题目
	ExamList  []int      //已完成的题库列表
	Answer    chan int  //开始游戏chan
	GameStart chan bool  //chan
	Over      chan bool  //离开游戏chan
	examids   []int
	lock sync.RWMutex
}

// new game
func NewGame(r *room) *Game {
	game := &Game{
		Status:    0,
		Answer:    make(chan int),
		GameStart: make(chan bool),
		Over:      make(chan bool),
		rm:        r,
		ExamList:  make([]int, 0),
		Users:    make([]*player,0),
		exam:     make(map[int]model.KsQuestion,0),
		examids:  make([]int,0),
	}
	//查询所有的题目
    game.exam ,_ = model.GetAllExamId()

	go func(game *Game) {
		for {
			fmt.Println("33333333")
			if ok := <-game.GameStart; ok {
				game.playGame()
			}
		}
	}(game)

	return game
}

// 添加一个玩家到房间中
func (game *Game) addPlayer(userId int) error {
	log.Printf("添加用户到房间中")
	//初始化一个玩家
	player, err := NewPlayer(userId)
	if err != nil {
		fmt.Println("创建游戏用户失败：", err)
		return err
	}
	log.Printf("在线player%v",player)
	//加入到房间列表
	game.Users = append(game.Users, player)

	var users []*onlineUser
	for _, user := range game.Users { //返回给客户端的用户信息
		log.Printf("在线puserr%v",user)
		u, err := h.GetOnlineUser(user.UserId)
		log.Printf("在线？%v",u)
		if err != nil {
			fmt.Println("获取在线用户失败：", err)
			continue
		}
		users = append(users, u)
	}
   log.Printf("在线列表%v",users)
	game.send("JoinRoom", map[string]interface{}{
		"Room": map[string]interface{}{
			"Id":     game.rm.Id,
			"Name":   game.rm.Name,
			"UserId": userId,
			"Status": game.Status,
			"Time":   game.rm.Time,
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
    if len(game.Users) >= 2{
    	isStart = true
	}
	game.send("Ready", map[string]interface{}{
		"UserId": userId,
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
	return true
}

// 开始游戏
func (game *Game) playGame() {
	//检查是不是够2人
	if !game.checkIsReady() {
		return
	}
	exam, err := game.getExam()
	if err != nil {
		fmt.Println("没有找到题目",err)
		game.endGame()
		return
	}
	// 添加该题进入过滤slice
	game.ExamList = append(game.ExamList, exam.Id)


	if game.Status == 0 { // 开始一局游戏
		game.Status = 1 //更改房间游戏状态
	}
	//继续游戏
	for _, user := range game.Users {
		fmt.Println(user)
	}
	//发送消息正在玩游戏
	game.send("PlayGame", map[string]interface{}{
		"Exam":     exam,
		"Users":    game.Users,
		"GameTime": playGameTime,
	})
   //20秒倒计时答题
	GameOutTime := playGameTime

	timer := time.NewTimer(GameOutTime * time.Second)
	for { //wait submit answer
		select {
		case answer := <-game.Answer:
			userId := answer%10000
			answerId := answer - userId
			fmt.Println("提交答案", answerId)
			fmt.Println("用户id", userId)
			game.submit(userId,answerId)
			return
		case <-game.Over: //游戏结束
			fmt.Println("游戏结束")
			game.endGame()
			return
		case <-timer.C: //游戏超时未提交答案，随机一个题？
			fmt.Println("超时未答题")
			game.endGame()
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
			log.Printf("退出房间找到了%v"	,user.UserId)
			delkey = key
		}
		//else {
			//log.Printf("退出房间没有找到")
			//u, err := h.GetOnlineUser(userId)
			//if err != nil {
			//	fmt.Println("获取用户失败：", err)
			//	continue
			//}
			//
			//if game.Status == 1 { // 游戏中的状态
			//	game.Over <- true
			//	h.sendToClient("GameOver", user.UserId, map[string]interface{}{
			//		"OverUser":     userId,
			//		"OverUserName": u.Users.UserName,
			//	})
			//} else {
			//	h.sendToClient("OutRoom", user.UserId, map[string]interface{}{
			//		"OverUser":     userId,
			//		"OverUserName": u.Users.UserName,
			//	})
			//}
		//}
	}
	if delkey >= 0 && game.Users != nil{
		fmt.Println("将用户", userId, "从房间", game.rm.Id, "中移除")
		game.Users = append(game.Users[:delkey], game.Users[delkey+1:]...)
		if len(game.Users) == 0 {
			game.ExamList = make([]int, 0)
		}
	}
}

// 提交答案
func (game *Game) submit(userId,answer int) {
	if game.Status != 1 {
		fmt.Println("不在游戏中，请勿随便提交答案")
		return
	}
	isOk := false
	// 是否答对
	//TODO
	if game.exam[1].AnswerId == answer {
		isOk = true
	}
	isEnd := false //是否结束了

	for _, user := range game.Users {
			if isOk == true && userId == user.UserId{ //答对，并且是自己，那么自己答对的题目+1
				user.count()
			} else { //对方加1
				user.count()
			}
		game.send("GameResult", map[string]interface{}{
			"Answer":     game.exam[1].AnswerId,
			"IsOk":       isOk,
			"UserId":     userId,
			"UserAnswer": answer,
		})
	}



	time.Sleep(3 * time.Second) //延时3秒, 让客户端等待缓冲

	if isEnd == true { //结束一局游戏
		game.endGame()
	} else { //重新开始游戏
		game.playGame()
	}
}

// 获取当前答题的题目
// 全部放到内存中
func (game *Game) getExam() (model.KsQuestion, error) {
	exam := game.getRandExamId()
	if exam.AnswerId == 0 { //已经完成所有题目
		return model.KsQuestion{}, ErrNotExam
	}
	return exam, nil
}

// 随机获取题目ID
// 过滤掉该房间已经完成过的题目
func (game *Game) getRandExamId() model.KsQuestion {
	isNotList := false
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := len(game.examids)
	if num == 0 || len(game.examids) == len(game.ExamList) {
		return model.KsQuestion{}
	}
	for {
		isNotList = false
		examId := game.examids[r.Intn(num)]
		for _, inExamId := range game.ExamList {
			if examId == inExamId {
				isNotList = true
			}
		}
		if isNotList == false {
			return game.exam[examId]
		}
	}
	return model.KsQuestion{}
}

// 游戏结束
func (game *Game) endGame() {
	game.clearGame()

	game.send("EndGame", map[string]interface{}{
		"Users": game.Users,
	})
}

// 游戏结束, 清空状态
func (game *Game) clearGame() {
	fmt.Println("清空房间信息")

	// 重置用户游戏状态
	for _, u := range game.Users {
		u.Count = 0
		u.Win = false
	}

	// 重置游戏状态
	game.exam = make(map[int]model.KsQuestion)
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
		log.Printf("发消息%v",user)
		h.sendToClient(action, user.UserId, res)
	}
}
