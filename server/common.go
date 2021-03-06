package server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/vckai/GoAnswer/model"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
)

const (
	writeWait                     = 10 * time.Second
	pongWait                      = 30 * time.Second
	pingPeriod                    =  (pongWait *9) /10
	messageSize                   = 512
	roomMaxUser                   = 2
	playGameTime    time.Duration = 20 //每局游戏时间
)

var (
	roomInc     uint32 = 1        //房间ID流水号
)

// 连接管理
type Connection struct {
	userId int
	ws     *websocket.Conn
	send   chan []byte
}

// api返回
type apiParam struct {
	Action string
	UserId int
	Params map[string]interface{}
	Time   int64
}

// 在线用户
type onlineUser struct {
	model.Users
	RoomId uint32
}

type hub struct {
	register      chan *Connection   //登陆时候的channel
	unregister    chan *Connection   //退出登陆的chanel
	connections   map[int]*Connection  //登陆成功的连接
	notloginconns map[int]*Connection  //还未登陆成功
	broadcast     chan *simplejson.Json  //消息
	onlineUsers   map[int]*onlineUser   //在线用户
	rooms         map[uint32]*room     //room 列表
	examids       []int                //所有试卷题目id

	lock sync.Mutex
}

var h = &hub{
	register:      make(chan *Connection),
	unregister:    make(chan *Connection),
	connections:   make(map[int]*Connection),
	notloginconns: make(map[int]*Connection),
	onlineUsers:   make(map[int]*onlineUser),
	broadcast:     make(chan *simplejson.Json),
	rooms:         make(map[uint32]*room),
}

// 生成房间ID
func newRoomId() uint32 {
	atomic.AddUint32(&roomInc, 1)

	return roomInc
}

// 初始化一个新连接
func NewConn(userId int, ws *websocket.Conn) (*Connection, error) {
	//如果当前用户拥有连接，那么替换

	c := &Connection{
		userId: userId,
		ws:     ws,
		send:   make(chan []byte, 256),
	}

	h.register <- c

	return c, nil
}
