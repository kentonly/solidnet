package solidnet

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	DISPATCH_WAIT_TIME  = 2
	MAX_LOGIN_AUTH_TIME = 15
)

const (
	EVENT_LOGIN_AUTH_TIMER = 1
)

type TcpClient struct {
	*BaseClient
	processor      IProcessor
	stopWait       sync.WaitGroup
	closeFlag      chan int32
	loginFlag      bool
	loginAuthTimer ITimer
}

func NewTcpClient(conn *net.TCPConn, p IProcessor, f IPacketFactory) *TcpClient {
	c := &TcpClient{
		BaseClient: NewBaseClient(conn, f),
		processor:  p,
		closeFlag:  make(chan int32),
	}
	c.loginAuthTimer = NewTimer(EVENT_LOGIN_AUTH_TIMER, c)
	return c
}

// 实现 ITimerHandler
func (c *TcpClient) DoTimerAction(id int32) {
	switch id {
	case EVENT_LOGIN_AUTH_TIMER:
		// 如果没有登录验证，则关闭连接
		if !c.GetLoginFlag() {
			fmt.Printf("client[%s] loginauth timeout", c.RemoteAddr())
			c.stop()
		}
	default:

	}
}

func (c *TcpClient) SetLoginFlag(flag bool) {
	c.loginFlag = flag
}

func (c *TcpClient) GetLoginFlag() bool {
	return c.loginFlag
}

func (c *TcpClient) Run() {
	c.stopWait.Add(1)
	go c.dispatchStateMessage()
	c.stopWait.Add(1)
	go c.dispatchNetMessage()
	c.stopWait.Wait()
}

func (c *TcpClient) dispatchNetMessage() {
	defer c.stopWait.Done()
	for {
		var data []byte
		select {
		case data = <-c.input:
		case <-c.closeFlag:
			// 退出协程
			return
		case <-time.After(time.Second * DISPATCH_WAIT_TIME):
			continue
		}
		c.processor.Dispatch(&NetMessage{data, c})
	}
}

func (c *TcpClient) dispatchStateMessage() {
	defer c.stopWait.Done()
	for {
		var state int32
		select {
		case state = <-c.state:
		case <-time.After(time.Second * DISPATCH_WAIT_TIME):
			continue
		}
		c.processor.Dispatch(&StateMessage{state, c})
		switch state {
		case STATE_CONNECTED:
			// 连接后，一定时间内进行登录认证，否则视为非法用户
			c.loginAuthTimer.Start(MAX_LOGIN_AUTH_TIME*time.Second, false)
		case STATE_CLOSED:
			// 连接已经关闭，通知其他协程
			c.closeFlag <- 1
			return
		}
	}
}
