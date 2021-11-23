package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"

	"github.com/vckai/GoAnswer/model"

	"github.com/bitly/go-simplejson"
)

var (
	ErrNotExistsOnlineUser = errors.New("没有找到该用户")
	ErrUserInRoom          = errors.New("该用户已经在房间内")
	ErrApiParam            = errors.New("接口参数错误")
	ErrUserNotExtsisRoom   = errors.New("用户不在房间中")
	ErrRoomNotExists       = errors.New("房间不存在")
)
const (
	writeWait                     = 10 * time.Second
	pongWait                      = 30 * time.Second
	pingPeriod                    =  (pongWait *9) /10
	messageSize                   = 512
	roomMaxUser                   = 2
	playGameTime    time.Duration = 20 //每局游戏时间
)

type Room struct {
	id      int
	Clients map[int]*websocket.Conn
	timer   int //每步20秒
	status  int //0空，1，有1个人，2，有2个人,为2个人就准备好了
	battle  *player
	mu      sync.Mutex
}
// 房间用户信息
type player struct {
	UserId int
	Status bool  //用户准备状态(true准备中, false尚未准备)
	Win    bool  //用户是否胜利
	Count  int16 //答对题数
	Conn   *websocket.Conn
	lock    sync.Mutex
}

// api返回
type apiParam struct {
	Action string
	UserId int
	Params map[string]interface{}
	Time   int64
}

// 连接管理，在线用户
type Connection struct {
	UserId int
	UserName string
	ws     *websocket.Conn
	send   chan []byte
	RoomId int
}
//全局游戏处理
type hub struct {
	register      chan *Connection   //登陆时候的channel
	unregister    chan *Connection   //退出登陆的chanel
	connections   map[int]*Connection
	broadcast     chan *simplejson.Json  //消息
	examids       []int                //所有试卷题目id
	rooms         []*Room
	r             *Room
	mu            sync.RWMutex
	lock          sync.Mutex
}

var h = &hub{
	register:      make(chan *Connection,100),
	unregister:    make(chan *Connection,100),
	connections:    make(map[int]*Connection),
	broadcast:     make(chan *simplejson.Json),

}

//初始化，生成房间
func init()  {

	h.rooms = make([]*Room, 8)
	for i := range h.rooms {
		fmt.Println("生成房间",i)
		h.rooms[i] = &Room{
			id:      i+1,
			Clients: make(map[int]*websocket.Conn, 2),
		}
	}
}

//根据房间id，查询信息
func (w *hub) GetById(id int) *Room {
	return w.rooms[id]
}

func (w *hub) addRooms() {

	w.rooms = append(w.rooms, &Room{id: len(w.rooms), Clients: make(map[int]*websocket.Conn, 2)})

}

//查找正在等待匹配的房间
func (w *hub) WaitingRoom() int {
	w.mu.RLock()
	for _, v := range w.rooms {
		if v.status == 1 {
			w.mu.RUnlock()
			return v.id
		}
	}
	return -1
}
//找到第一个空房间
func (w *hub) FirstEmptyRoom() int {
	w.mu.RLock()
	for _, v := range w.rooms {
		if v.status == 0 {
			w.mu.RUnlock()
			return v.id
		}
	}
	return -1
}

//当前用户进房间
func (w *hub) inRoom(c *Connection) (roomid int, err error) {
	waitingRoom := w.WaitingRoom()
	fmt.Println("走到了进房间的逻辑，",waitingRoom)
	if waitingRoom == -1 {
		//没有等待的房间
		emptyRoom := w.FirstEmptyRoom()
		fmt.Println("emptyRoom，",emptyRoom)
		if emptyRoom == -1 {
			//人居然满了我是不信的
			//添加房间
			w.addRooms()
			return w.inRoom(c)
		} else {
			//用户触发进入房间的逻辑
			_, err = w.rooms[emptyRoom].ClientIn(c)
		}
		roomid = emptyRoom
		//把自己的房间id更新
		c.RoomId = roomid

	} else {
		fmt.Println("waitingRoom，",waitingRoom)
		//用户触发进入房间的逻辑
		_, err = w.rooms[waitingRoom].ClientIn(c)
		roomid = waitingRoom
		//把自己的房间id更新
		c.RoomId = roomid

	}
	return roomid, err
}
//用户进入
func (r *Room) ClientIn(c *Connection) (bool, error) {
	ok := false
	if r.status > 2 {
		return false, errors.New("this room is full")
	}

	r.mu.Lock()
	r.Clients[c.UserId] = c.ws
	r.mu.Unlock()
	r.UpdateStatus()
	if r.status == 2 {
		ok = true
	}
	return ok, nil
}
//是否满2人  满2人就可以开展
func (r *Room) UpdateStatus() {
	r.mu.Lock()
	r.status = len(r.Clients)
	r.mu.Unlock()
}
func (r *Room) initClient() {
	r.mu.Lock()
	r.Clients = make(map[int]*websocket.Conn, 2)
	r.mu.Unlock()
}
//重置房间信息
func (r *Room) Clear() {
	r.initClient()
	r.UpdateStatus()
}

// 初始化一个新连接
func NewConn(userId int, ws *websocket.Conn) *Connection {
	//存在就替换
	c := &Connection{
		UserId: userId,
		ws:     ws,
		send:   make(chan []byte, 256),
		RoomId: 0 ,
	}

	//每次新登陆都替换成最新的链接
	if  _, ok  := h.connections[userId];ok{
		fmt.Println("存在",userId)
      delete(h.connections,userId)
	}
	h.connections[userId] = c


	//chan通知有人注册
	h.register <- c

	return c
}


// 创建一个玩家
// userId  玩家ID
func NewPlayer(userId int,conn *websocket.Conn) (*player, error) {
	return &player{
		UserId:  userId,
		Status:  false,
		Win:     false,
		Conn:    conn,
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
//初始化服务器服务器，获取所有题目
func InitServer() {
	h.examids, _ = model.GetAllExamId()

	if len(h.examids) == 0 {
		fmt.Println("没有找到题目")
	}
	fmt.Println(h.examids)
	go h.run()
}

func GetServer() *hub {
	return h
}

// 开始运行
func (this *hub) run() {
	for {
		select {
		case c := <-this.register:  //登陆
			this.login(c)
		case c := <-this.unregister:  //退出
			this.logout(c)
		case param := <-this.broadcast: //消息
			this.handle(param)
		}
	}
}

// 接口执行
func (this *hub) handle(param *simplejson.Json) {
	userId, err := param.Get("UserId").Int()
	if err != nil || userId < 1 {
		fmt.Println("传入参数错误，UserID：", userId, "---ERROR：", err)
		return
	}

	switch param.Get("Action").MustString() {
	case "Submit": //提交答案
		this.submitAnswer(param)
	case "OutRoom": //退出房间
		this.outRoom(param)
	}
}


// 初始化登录
func (this *hub) login(c *Connection) {
	this.lock.Lock()
	defer this.lock.Unlock()

	c, ok := this.connections[c.UserId]
	if !ok {
		fmt.Println("该用户尚未建立连接", c.UserId)
		return
	}
	//进房间
	roomid,_ :=this.inRoom(c)

	fmt.Println("用户ID：", c.UserId, "，姓名：", c.UserName, "，登录成功;房间id：",roomid)
	//下发消息
	this.sendToClient("JoinRoom", c.UserId, map[string]interface{}{
			"OverUser":     c.UserId,
			"OverUserName": c.UserId,
		})

}

// 退出房间
func (this *hub) outRoom(param *simplejson.Json) {

	userId := param.Get("UserId").MustInt()
	fmt.Println("用户", userId, "退出房间")

	this.lock.Lock()
	user, ok := this.connections[userId]
	if !ok {
		fmt.Println("获取用户失败：", userId)
		this.lock.Unlock()
		return
	}
	inRoomId := user.RoomId
	user.RoomId = 0
	this.lock.Unlock()

	if inRoomId > 0 {
		//this.getRoom(inRoomId).Game.GameOver(userId)
		//this.sendToClient("OutRoom", userId, map[string]interface{}{
		//	"OverUser":     userId,
		//	"OverUserName": user.UserName,
		//})
	}
}

// 退出socket登录连接, 各种清理
func (this *hub) logout(c *Connection) {
	this.lock.Lock()
	defer this.lock.Unlock()

	fmt.Println("用户", c.UserId, "退出登录")

	_, ok := this.connections[c.UserId]
	if !ok {
		fmt.Println("获取用户失败：", c.UserId)
		return
	}
	delete(this.connections, c.UserId)
	close(c.send)
}


// 提交答案
func (this *hub) submitAnswer(param *simplejson.Json) error {
	userId := param.Get("UserId").MustInt()
	answerId := param.Get("Params").Get("AnswerId").MustInt()
	if userId == 0 {
		fmt.Println("接口参数错误", userId, answerId)
		return ErrApiParam
	}
	user, err := this.GetOnlineUser(userId)
	if err != nil {
		fmt.Println("获取在线用户失败：", err)
		return err
	}
	if user.RoomId == 0 {
		fmt.Println("该用户不在房间中", user.UserId)
		return ErrUserNotExtsisRoom
	}
	r := this.GetById(user.RoomId)
	if r == nil {
		fmt.Println("获取房间失败：", user.RoomId)
		return ErrRoomNotExists
	}
	//r.Answer <- userId*10000+answerId

	return nil
}


// 获取socket链接
func (this *hub) getClient(userId int) *Connection {
	this.lock.Lock()
	defer this.lock.Unlock()

	if _,ok := this.connections[userId];!ok{
			fmt.Println("没有获取到该socket")
			return nil
	}else {
		fmt.Println("获取到该socket",this.connections[userId])
	}

	return this.connections[userId]
}

func (this *hub) GetOnlineUser(userId int) (*Connection, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	u, ok := this.connections[userId]
	if !ok {
		return nil, ErrNotExistsOnlineUser
	}

	return u, nil
}






// 获取房间信息
func (this *hub) getRoom(roomId int) *Room {
	for _, val := range this.rooms {
		if val.id == roomId {
			return val
		}
	}
	return nil
}

// 发送消息到指定客户端
func (this *hub) sendToClient(action string, userId int, data map[string]interface{}) {
	c := this.getClient(userId)
	fmt.Println("下发通知",c.ws)
	if c == nil {
		fmt.Println("该用户不存在", userId)
		return
	}
	c.send <- this.genRes(action, userId, data)
}

// 统一返回的数据结构
func (this *hub) genRes(action string, userId int, params map[string]interface{}) []byte {
	v := &apiParam{Action: action, UserId: userId, Params: params, Time: time.Now().Unix()}
	s, err := json.Marshal(v)
	if err != nil {
		fmt.Println("生成JSON数据错误, ERR: ", err)
		return []byte{}
	}
	return []byte(s)
}
