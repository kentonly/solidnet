package solidnet

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	logger "github.com/idakun/tinylog"
)

const (
	MAX_CHANNEL_LEN     = 1000
	MAX_USER_PACKET_LEN = 20 * 1024 // 业务层包体最大20KB
	MAX_SEND_TIMEOUT    = 1
	MAX_RECV_TIMEOUT    = 1
)

// 状态通知
const (
	STATE_CONNECTED = 0 // 连接状态
	STATE_CLOSED    = 1 // 断线状态
)

type IClient interface {
	Send(data []byte) bool               //异步发送
	SendSync(data []byte) (int32, error) //同步发送
	LocalAddr() string
	RemoteAddr() string
	SetLoginFlag(bool)
	GetLoginFlag() bool
}

type BaseClient struct {
	conn       net.Conn
	input      chan []byte // 接受数据
	output     chan []byte // 发送数据
	state      chan int32  // 状态通知
	mutex      sync.Mutex
	remoteAddr string
	localAddr  string

	factory IPacketFactory
}

func NewBaseClient(conn net.Conn, f IPacketFactory) *BaseClient {
	c := new(BaseClient)
	c.output = make(chan []byte, MAX_CHANNEL_LEN)
	c.input = make(chan []byte, MAX_CHANNEL_LEN)
	c.state = make(chan int32, MAX_CHANNEL_LEN)
	c.factory = f
	c.conn = conn
	c.remoteAddr = conn.RemoteAddr().String()
	c.localAddr = conn.LocalAddr().String()

	c.run()
	return c
}

func (c *BaseClient) Send(data []byte) bool {
	select {
	case c.output <- data:
		return true
	case <-time.After(time.Second * MAX_SEND_TIMEOUT):
		logger.Fatal("Send() timeout!!!")
		return false
	}
}

func (c *BaseClient) SendSync(data []byte) (int32, error) {
	defer func() {
		err := recover()
		if nil != err {
			logger.Fatal("panic[%v]", err)
		}
	}()
	if !c.isRunning() {
		return 0, errors.New("client have stop")
	}
	n, err := c.conn.Write(data)
	if nil != err {
		logger.Error("c.conn.Write() failed, error[%s]", err.Error())
		c.stop()
	}
	return int32(n), err
}

func (c *BaseClient) LocalAddr() string {
	return c.localAddr
}

func (c *BaseClient) RemoteAddr() string {
	return c.remoteAddr
}

// 启动2个协程，如果对端关闭或者出错，2个协程会先后退出
func (c *BaseClient) run() {
	// 启动接受数据的协程
	go c.recv()

	// 启动发送数据的协程
	go c.send()

	// 向应用层通知有新的连接
	c.notifyState(STATE_CONNECTED)
}

func (c *BaseClient) isRunning() bool {
	return nil != c.conn
}

func (c *BaseClient) stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isRunning() {
		return
	}
	c.conn.Close()
	c.conn = nil

	// 向应用层通知断线
	c.notifyState(STATE_CLOSED)
}

func (c *BaseClient) notifyState(state int32) {
	select {
	case c.state <- state:
	default:
		logger.Fatal("notifyState failed, channel is already full!!!")
	}
}

func (c *BaseClient) send() {
	defer func() {
		err := recover()
		if nil != err {
			logger.Fatal("panic[%v]", err)
		}
	}()

	for {
		// 退出协程
		if !c.isRunning() {
			return
		}

		select {

		case data := <-c.output:
			_, err := c.conn.Write(data)
			if nil != err {
				logger.Error("c.conn.Write() failed, error[%s]", err.Error())
				c.stop()
			}
		case <-time.After(time.Second * MAX_RECV_TIMEOUT):
			// Do nothing
		}
	}
}

func (c *BaseClient) recv() {
	defer func() {
		err := recover()
		if nil != err {
			logger.Fatal("panic[%v]", err)
		}
	}()

	for {
		// 退出协程
		if !c.isRunning() {
			return
		}
		p := c.factory.NewPacket()

		// 读取包头
		head := make([]byte, p.GetHeadLen())
		_, err := io.ReadFull(c.conn, head)
		if nil != err {
			logger.Error("io.ReadFull() failed, error[%s]", err.Error())
			c.stop()
			continue
		}

		p.WriteBytes(head)
		bodyLen := p.GetBodyLen()
		if bodyLen >= MAX_USER_PACKET_LEN {
			// 包体太长，有可能是网络攻击包或者错误包，丢掉不处理
			logger.Error("length of uesr packet more than MAX_USER_PACKET_LEN, bodyLen=%d", bodyLen)
			continue
		}

		// 读取包体
		bodyData := make([]byte, bodyLen)
		_, err = io.ReadFull(c.conn, bodyData)
		if nil != err {
			logger.Error("io.ReadFull() failed, error[%s]", err.Error())
			c.stop()
			continue
		}
		p.WriteBytes(bodyData)
		select {
		case c.input <- p.GetData():
		case <-time.After(time.Second * MAX_RECV_TIMEOUT):
			logger.Fatal("input channel is already full!!!")
		}
	}
}
