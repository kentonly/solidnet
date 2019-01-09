package solidnet

import (
	"sync"
	"time"

	logger "github.com/idakun/tinylog"
)

var (
	p    IProcessor // 全局单例
	once sync.Once
)

type IProcessor interface {
	Dispatch(IMessage)
	Epoll() IMessage
}

func GetProcessor() IProcessor {
	once.Do(func() {
		p = &ChannelProcessor{make(chan IMessage, MAX_CHANNEL_LEN)}
	})
	return p
}

type ChannelProcessor struct {
	messageChannel chan IMessage
}

func (p *ChannelProcessor) Dispatch(message IMessage) {

	select {
	case p.messageChannel <- message:

	case <-time.After(time.Second * MAX_SEND_TIMEOUT):
		//超时，导致数据丢弃
		logger.Error("send to packet channel timeout!!!")
	}
}

func (p *ChannelProcessor) Epoll() IMessage {
	return <-p.messageChannel
}
