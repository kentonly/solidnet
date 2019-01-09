package solidnet

import (
	"time"
)

// 这里必须初始化定时消息处理器
var (
	TimerMsgprocessor IProcessor = GetProcessor()
)

type ITimer interface {
	Start(time.Duration, bool)
	Stop()
}

type ITimerHandler interface {
	DoTimerAction(int32)
}

type Timer struct {
	timer     *time.Timer
	id        int32
	isLoop    bool
	timeout   time.Duration
	isRunning bool
	handler   ITimerHandler
}

func NewTimer(id int32, h ITimerHandler) *Timer {
	t := new(Timer)
	t.id = id
	t.handler = h
	return t
}

func (t *Timer) Start(timeout time.Duration, isLoop bool) {
	t.isLoop = isLoop
	t.timeout = timeout
	t.timer = time.AfterFunc(timeout, t.procTimeout)
}

func (t *Timer) Stop() {
	if nil != t.timer {
		t.timer.Stop()
	}
}

func (t *Timer) procTimeout() {
	message := &TimerMessage{t.id, t.handler}
	TimerMsgprocessor.Dispatch(message)

	if t.isLoop {
		t.timer.Reset(t.timeout)
	}
}
