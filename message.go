package solidnet

type IMessage interface {
	Data() interface{} //消息的数据
	Args() interface{} //消息的参数
}

// 网络消息
type NetMessage struct {
	packet []byte
	client IClient
}

func (m *NetMessage) Data() interface{} {
	return m.packet
}

func (m *NetMessage) Args() interface{} {
	return m.client
}

// 定时器消息
type TimerMessage struct {
	id      int32
	handler ITimerHandler
}

func (m *TimerMessage) Data() interface{} {
	return m.id
}

func (m *TimerMessage) Args() interface{} {
	return m.handler
}

// 状态消息
type StateMessage struct {
	state  int32
	client IClient
}

func (m *StateMessage) Data() interface{} {
	return m.state
}

func (m *StateMessage) Args() interface{} {
	return m.client
}
