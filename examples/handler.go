package main

import (
	"time"

	solidnet "github.com/idakun/solidnet"

	logger "github.com/idakun/tinylog"
)

type Handler struct {
	router map[uint16]func(*Packet, solidnet.IClient) int
}

func NewHandler() solidnet.IHandler {
	h := &Handler{router: make(map[uint16]func(*Packet, solidnet.IClient) int)}
	h.RegisterRouter(CLIENT_COMMAND_TIME_REQ, h.HandleTimeReq)
	h.RegisterRouter(CLIENT_COMMAND_LOGIN_AUTH, h.HandleLoginAuth)
	return h
}

// 派发定时器消息
func (h *Handler) HandleTimer(message solidnet.IMessage) {
	id := (message.Data()).(int32)
	handler := (message.Args()).(solidnet.ITimerHandler)
	handler.DoTimerAction(id)
}

// 派发网络消息
func (h *Handler) HandleNet(message solidnet.IMessage) {
	buf := (message.Data()).([]byte)
	c := (message.Args()).(solidnet.IClient)

	p := NewPacket()
	p.Refer(buf)
	cmd := uint16(p.GetCmd())

	h.handle(cmd, p, c)
}

// 派发状态消息
func (h *Handler) HandleState(message solidnet.IMessage) {
	state := (message.Data()).(int32)
	client := (message.Args()).(solidnet.IClient)
	_ = client
	if solidnet.STATE_CLOSED == state {
		// 处理断线
		//logger.Debug("client[%s] closed.", client.RemoteAddr())
		// ......
	} else if solidnet.STATE_CONNECTED == state {
		// 这里可以限制时间，connect之后，不登陆则关闭连接
		//logger.Debug("client[%s] connected.", client.RemoteAddr())
		// ......
	}
}

func (h *Handler) handle(cmd uint16, in *Packet, c solidnet.IClient) {
	if handle, ok := h.router[cmd]; ok {
		handle(in, c)
	} else {
		logger.Error("Not find handler of cmd:%x", cmd)
	}
}

func (h *Handler) RegisterRouter(cmd uint16, handle func(*Packet, solidnet.IClient) int) {
	h.router[cmd] = handle
}

func (h *Handler) HandleTimeReq(packet *Packet, c solidnet.IClient) int {
	p := NewPacket()
	p.WriteBegin(SERVER_COMMAND_TIME_RESP, 0)
	p.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	p.WriteEnd()
	c.Send(p.GetData())
	return 0
}

func (h *Handler) HandleLoginAuth(packet *Packet, c solidnet.IClient) int {
	authKey := packet.ReadInt32()
	var authSuccess int32
	if AUTH_KEY == authKey {
		// 设置客户端登录标记
		c.SetLoginFlag(true)
		authSuccess = 1
	}

	// 回复认证结果
	p := NewPacket()
	p.WriteBegin(SERVER_COMMAND_AUTH_SUCCESS, 0)
	p.WriteInt32(authSuccess)
	p.WriteEnd()
	c.Send(p.GetData())

	return 0
}
