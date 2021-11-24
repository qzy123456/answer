package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"time"

	"github.com/vckai/GoAnswer/model"

	"github.com/bitly/go-simplejson"
	"log"
)

var (
	ErrNotExistsOnlineUser = errors.New("没有找到该用户")
	ErrUserInRoom          = errors.New("该用户已经在房间内")
	ErrApiParam            = errors.New("接口参数错误")
	ErrUserNotExtsisRoom   = errors.New("用户不在房间中")
	ErrRoomNotExists       = errors.New("房间不存在")
)
//初始化服务器服务器，获取所有题目
func InitServer() {
	go h.run()
}

func GetServer() *hub {
	return h
}

// 心跳检测
func (this *hub) HeartBeat() {
	fmt.Println("新设计的心跳包")
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for userId,users := range this.onlineUsers {
				c := this.getClient(userId)
				fmt.Println("心跳包",c)
				if c != nil {
					//如果当前用户有问题，发送失败，那么就删除
					if err := c.write(websocket.PingMessage, []byte{}); err != nil {
						fmt.Println("失败了，要删除在线列表，", userId,users.RoomId)
						//删除在线列表
						delete(this.onlineUsers,userId)
						//删除链接列表
						delete(this.connections,userId)
						//房间用户删除
						this.rooms[users.RoomId].Game.GameOver(userId)
					}
				}
			}
		}
	}
}

// 开始运行
func (this *hub) run() {
	//心跳包
	go h.HeartBeat()
	for {
		select {
		case c := <-this.register:  //登陆
			this.reg(c)
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

	if param.Get("Action").MustString() != "Login" {
		c := this.getClient(userId)
		if c == nil {
			fmt.Println("尚未登录, 无法使用该接口：", param)
			return
		}
	}

	switch param.Get("Action").MustString() {
	case "Login": //登录
		this.login(param)
	case "Submit": //提交答案
		this.submitAnswer(param)
	case "OutRoom": //退出房间
		this.outRoom(param)
	case "Ready": //用户准备
		this.ready(param)
	}
}

// 保存socket连接
// 该链接尚属于未登陆状态
func (this *hub) reg(c *Connection) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.connections[c.userId] = c
}

// 初始化登录
func (this *hub) login(param *simplejson.Json) {
	this.lock.Lock()
	defer this.lock.Unlock()

	userId, _ := param.Get("UserId").Int()
	userName, _ := param.Get("UserName").String()

	_, ok := this.connections[userId]
	if !ok {
		fmt.Println("该用户尚未建立连接", userId)
		return
	}
	user := model.Users{
		UserId:userId,
		UserName:userName,
	}

	u := &onlineUser{user, 0} //在线用户信息, 房间ID

	this.onlineUsers[userId] = u
	log.Println("用户ID：", userId, "，姓名：", user.UserName, "，登录成功")
	//登陆成功 分配 加入房间

	if u.RoomId > 0 { //该用户已经加入房间中了
		fmt.Println("禁止重复进入房间", userId)
		return
	}
	var err error
	var rm *room
	//查找房间
	rm  = this.findRoom()
	if rm == nil { //没有房间则新建一个
		rm, err = NewRoom()
		log.Printf("新创建一个房间%v",rm)
		if err != nil {
			fmt.Println("创建房间失败：", err)
			return
		}

		this.rooms[rm.Id] = rm
	}
	if err := rm.addPlayer(userId); err != nil {
		fmt.Println("添加用户进入房间失败：", userId, err)
		return
	}
	this.onlineUsers[userId].RoomId = rm.Id
	//下发消息
	fmt.Println("用户加入房间成功：", rm.Id)
}

// 退出房间
func (this *hub) outRoom(param *simplejson.Json) {

	userId := param.Get("UserId").MustInt()
	fmt.Println("用户", userId, "退出房间")

	this.lock.Lock()
	user, ok := this.onlineUsers[userId]
	if !ok {
		fmt.Println("获取用户失败：", userId)
		this.lock.Unlock()
		return
	}
	inRoomId := user.RoomId
	user.RoomId = 0
	this.lock.Unlock()

	if inRoomId > 0 {
		this.getRoom(inRoomId).Game.GameOver(userId)
		this.sendToClient("OutRoom", userId, map[string]interface{}{
			"OverUser":     userId,
			"OverUserName": user.UserName,
		})
	}
}

// 退出socket登录连接, 各种清理
func (this *hub) logout(c *Connection) {
	this.lock.Lock()
	defer this.lock.Unlock()

	fmt.Println("用户", c.userId, "退出登录")

	user, ok := this.onlineUsers[c.userId]
	if !ok {
		fmt.Println("获取用户失败：", c.userId)
		return
	}

	if user.RoomId > 0 {
		this.getRoom(user.RoomId).Game.GameOver(c.userId)
	}
	delete(this.connections, c.userId)
	delete(this.onlineUsers, c.userId)
	close(c.send)
}

//// 进入房间
//func (this *hub) joinRoom(param *simplejson.Json) error {
//	userId := param.Get("UserId").MustInt()
//
//	u, err := this.GetOnlineUser(userId)
//	if err != nil {
//		fmt.Println("用户UID", userId, err)
//		return err
//	}
//
//	if u.RoomId > 0 { //该用户已经加入房间中了
//		fmt.Println("禁止重复进入房间", userId)
//		return ErrUserInRoom
//	}
//
//	//查找房间
//	rm := this.findRoom()
//
//	if rm == nil { //没有房间则新建一个
//		rm, err = NewRoom()
//		if err != nil {
//			fmt.Println("创建房间失败：", err)
//			return err
//		}
//
//		this.rooms[rm.Id] = rm
//	}
//	if err := rm.addPlayer(userId); err != nil {
//		fmt.Println("添加用户进入房间失败：", userId, err)
//		return err
//	}
//	this.lock.Lock()
//	this.onlineUsers[userId].RoomId = rm.Id
//	this.lock.Unlock()
//
//	return nil
//}

// 提交答案
func (this *hub) submitAnswer(param *simplejson.Json) error {
	userId := param.Get("UserId").MustInt()
	answerId := param.Get("Params").Get("AnswerId").MustInt()
	if userId == 0 {
		fmt.Println("接口参数错误", userId, answerId)
		return ErrApiParam
	}
	fmt.Println(this.onlineUsers)
	user, err := this.GetOnlineUser(userId)
	if err != nil {
		fmt.Println("获取在线用户失败：", err)
		return err
	}
	if user.RoomId == 0 {
		fmt.Println("该用户不在房间中", user.UserId)
		return ErrUserNotExtsisRoom
	}
	r := this.getRoom(user.RoomId)
	if r == nil {
		fmt.Println("获取房间失败：", user.RoomId)
		return ErrRoomNotExists
	}
	r.Game.Answer <- userId*10000+answerId

	return nil
}

// 用户准备
func (this *hub) ready(param *simplejson.Json) error {
	userId := param.Get("UserId").MustInt()
	if userId == 0 {
		fmt.Println("接口参数错误，缺少USERID")
		return ErrApiParam
	}
	fmt.Println(this.onlineUsers)
	user, err := this.GetOnlineUser(userId)
	if err != nil {
		fmt.Println("获取在线用户失败，", err, userId)
		return ErrNotExistsOnlineUser
	}
	r := this.getRoom(user.RoomId)
	if r == nil {
		fmt.Println("获取房间失败", user.RoomId)
		return ErrRoomNotExists
	}
	err = r.userReady(userId)

	return err
}

// 获取socket链接
func (this *hub) getClient(userId int) *Connection {
	c, ok := this.connections[userId]
	if !ok {
		fmt.Println("没有获取到该socket")
		return nil
	}
	return c
}

func (this *hub) GetOnlineUser(userId int) (*onlineUser, error) {

	u, ok := this.onlineUsers[userId]
	if !ok {
		return nil, ErrNotExistsOnlineUser
	}

	return u, nil
}

func (this *hub) DelOnlineUser(userId int)  {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, ok := this.onlineUsers[userId]
	if ok {
		delete(this.onlineUsers,userId)
	}

}
// 添加用户到在线列表中
func (this *hub) addOnlineUser(user *onlineUser) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.onlineUsers[user.UserId] = user //添加到在线用户信息

	return nil
}

// 查找未满人的房间
func (this *hub) findRoom() *room {
    log.Printf("房间信息%v",this.rooms)
	for _, val := range this.rooms {
		if val.Game.Status == 0 && len(val.Game.Users) < roomMaxUser{
			return val
		}
	}
	return nil
}

// 获取房间信息
func (this *hub) getRoom(roomId uint32) *room {
	for _, val := range this.rooms {
		if val.Id == roomId {
			return val
		}
	}
	return nil
}

// 房间列表
func (this *hub) getRooms() map[uint32]*room {
	this.lock.Lock()
	defer this.lock.Unlock()

	return this.rooms
}

// 删除某个房间
func (this *hub) delRoom(roomId uint32, rebUserId int) {
	this.lock.Lock()
	defer this.lock.Unlock()

	fmt.Println("删除房间", roomId, "删除机器人信息", roomId)
	delete(this.rooms, roomId)
	delete(this.onlineUsers, rebUserId) //删除机器人用户信息
}

// 发送消息到指定客户端
func (this *hub) sendToClient(action string, userId int, data map[string]interface{}) {
	c := this.getClient(userId)
	if c == nil {
		fmt.Println("该用户不存在", userId)
		return
	}
	c.send <- this.genRes(action, userId, data)
}

// 回收房间
func (this *hub) gcRoom() {
	for {
		this.clearRoom()
		time.Sleep(5 * time.Second)
	}
}

// 删除空房间
func (this *hub) clearRoom() {
	for _, room := range this.rooms {
		if room.Game.Status == 0 && len(room.Game.Users) == 1 {
			room.Close()
			delete(this.rooms, room.Id)
		}
	}
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
