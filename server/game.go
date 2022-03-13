package server

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chaoyang/answer/model"
)

var (
	ErrNotExam = errors.New("已经完成了所有题目")
)

type Game struct {
	rm     *room     //房间
	Users  []*player //房间用户列表
	Status int8      //房间状态(0等待1游戏中2满人)

	exam      map[int]model.KsQuestion //当前题目
	ExamList  []int                    //已完成的题库列表
	Answer    chan string              //开始游戏chan
	GameStart chan bool                //chan
	Over      chan int                 //离开游戏chan
	Winner    int                      //房间内哪个uuid赢了
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
		ExamList:  make([]int, 0),
		Users:     make([]*player, 0),
		exam:      make(map[int]model.KsQuestion, 0),
		Winner:    -1,
	}
	//查询所有的题目
	game.exam, _ = model.GetAllExamId()

	go func(game *Game) {
		for {
			fmt.Println("开始游戏了")
			if ok := <-game.GameStart; ok {
				//没题目就要新随机题目
				if len(game.exam) == 0 {
					game.exam, _ = model.GetAllExamId()
					fmt.Println("从新加载题目了", game.exam)
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
	//初始化一个玩家
	player, err := NewPlayer(userId)
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
	//一个人准备
	//h.sendToClient("Ready", userId, map[string]interface{}{
	//	"UserId": userId,
	//})

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
	exam, err := game.getExam()
	if err != nil {
		fmt.Println("没有找到题目", err)
		game.endGame(0)
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
			ans := strings.Split(answer, ",")
			userId, _ := strconv.Atoi(ans[0])
			questionId, _ := strconv.Atoi(ans[2])
			answerId, _ := strconv.Atoi(ans[1])
			fmt.Println("提交答案", answerId)
			fmt.Println("用户id", userId)
			fmt.Println("答案id", questionId)

			game.submit(userId, questionId, answerId)
			return
		case userId := <-game.Over: //游戏结束
			fmt.Println("游戏结束")
			game.endGame(userId)
			return
		case <-timer.C: //游戏超时未提交答案，随机一个题？
			fmt.Println("超时未答题")
			//game.endGame(0)
			//继续游戏，直到没题，或者人退出
			game.playGame()
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
			game.ExamList = make([]int, 0)
		}
	}
}

// 提交答案
func (game *Game) submit(userId, questionId, answerId int) {
	log.Println(userId, questionId, answerId)
	if game.Status != 1 {
		fmt.Println("不在游戏中，请勿随便提交答案")
		return
	}
	isOk := false
	// 是否答对
	trueAnwer := -1
	//循环题库
	for _, value := range game.exam {
		if questionId == value.Id && answerId == value.AnswerId {
			isOk = true
			trueAnwer = answerId
		}
	}
	isEnd := false //是否结束了
	//把答过的题放到已完成列表，之前随机的时候已经加过了
	//game.ExamList = append(game.ExamList, questionId)
	//下发消息，谁答了，是否答对
	for _, user := range game.Users {
		//答对，并且是自己，那么自己答对的题目+1，答错，对方+1
		if isOk == true {
			if userId == user.UserId {
				user.count()
			}
		} else { //对方加1
			if userId != user.UserId {
				user.count()
			}
		}
	}
	//根据房间的人数循环发
	game.send("GameResult", map[string]interface{}{
		"Answer":     trueAnwer,
		"IsOk":       isOk,
		"UserId":     userId,
		"UserAnswer": answerId,
	})

	time.Sleep(2 * time.Second) //延时3秒, 让客户端等待缓冲

	//检测是否要结束,题目都已经答完了，当前局结束
	if len(game.ExamList) >= len(game.exam) {
		isEnd = true
	}

	if isEnd == true { //结束一局游戏
	  fmt.Println("结束了？",game.ExamList ,game.exam)
		game.endGame(0)
	} else { //重新开始游戏
		game.playGame()
	}
}

// 获取当前答题的题目
// 全部放到内存中
func (game *Game) getExam() (model.KsQuestion, error) {
	exam := game.getRandExamId()
	fmt.Println("题目", exam)
	if exam.Id == 0 { //已经完成所有题目
		return model.KsQuestion{}, ErrNotExam
	}
	return exam, nil
}

// 随机获取题目ID
// 过滤掉该房间已经完成过的题目
func (game *Game) getRandExamId() model.KsQuestion {
	isNotList := false
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(game.ExamList) >= len(game.exam) {
		fmt.Println("题目答玩了")
		return model.KsQuestion{}
	}
	for {
		isNotList = false
		ranId := r.Intn(len(game.exam))
		fmt.Println("随机到了id",ranId)
		exam := game.exam[ranId]
		for _, inExamId := range game.ExamList {
			if exam.Id == inExamId {
				isNotList = true
			}
		}
		if isNotList == false {
			fmt.Println("随机一个题目？", ranId, exam)
			return game.exam[ranId]
		}
	}
	return model.KsQuestion{}
}

// 游戏结束
func (game *Game) endGame(userId int) {
	winner := 0
	if userId == 0 {
		winner = game.CheckWinner()

		game.send("EndGame", map[string]interface{}{
			"Users":  game.Users,
			"Winner": winner,
		})
	}
	//清空
	game.clearGame()
}

// 游戏结束, 清空状态
func (game *Game) clearGame() {
	fmt.Println("清空房间信息")

	// 重置用户游戏状态
	for _, u := range game.Users {
		u.Count = 0
		u.Status = false
	}

	// 重置游戏状态,清空房间
	game.exam = make(map[int]model.KsQuestion, 0) //从新加载的题
	game.ExamList = make([]int, 0)                //历史完成的题
	game.Status = 0
	game.Winner = -1
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

//检测谁赢了
func (game *Game) CheckWinner() int {
	fmt.Println("检测睡赢哦",len(game.exam),len(game.Users))
	if len(game.Users) >= 2 {
		if game.Users[0].Count > len(game.exam)/2 {
			game.Winner = game.Users[0].UserId
		} else if game.Users[1].Count > len(game.exam)/2 {
			game.Winner = game.Users[1].UserId
		} else { //平局
			game.Winner = 0
		}
	}
	return game.Winner
}
