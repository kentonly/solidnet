package solidnet

import (
	"fmt"

	logger "github.com/idakun/tinylog"
)

// 程序入口
func Run(game IGame) {
	if game.Init() {
		game.Run()
	}
}

type IGame interface {
	Init() bool
	Run()
}

type Game struct {
	name   string
	logDir string
	addr   string

	processor IProcessor
	handler   IHandler
	factory   IPacketFactory
}

func NewGame(addr string, name string, logDir string, h IHandler, f IPacketFactory) *Game {
	game := new(Game)
	game.addr = addr
	game.name = name
	game.logDir = logDir
	game.processor = GetProcessor()
	game.handler = h
	game.factory = f
	return game
}

func (g *Game) Run() {
	// 处理消息
	for {
		message := g.processor.Epoll()
		switch message.(type) {
		case *TimerMessage:
			g.handler.HandleTimer(message)
		case *NetMessage:
			g.handler.HandleNet(message)
		case *StateMessage:
			g.handler.HandleState(message)
		default:
			logger.Error("type of message is error!")
		}
	}
}

func (g *Game) Init() bool {
	// 初始化日志
	result := logger.Init(g.name, g.logDir, 30*1024*1024, 10, logger.DEBUG_LEVEL, false, logger.PUT_CONSOLE|logger.WRITE_FILE)
	if !result {
		fmt.Print("logger.Init() failed!!!")
		return false
	}

	// 开启tcp服务
	s := NewTcpServer(g.addr, g.processor, g.factory)
	if !s.Start() {
		logger.Error("TcpServer start failed.")
		return false
	}
	return true
}
