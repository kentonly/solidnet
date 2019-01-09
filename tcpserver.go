package solidnet

import (
	"net"
	"sync"

	logger "github.com/idakun/tinylog"
)

const (
	MAX_CLIENT_NUM = 10000 // 最多客户端数量
)

type TcpServer struct {
	Addr         string
	Clients      map[net.Conn]*TcpClient
	clientsWait  sync.WaitGroup
	lsn          *net.TCPListener
	clientsMutex sync.Mutex

	Processor IProcessor
	Factory   IPacketFactory
}

func NewTcpServer(addr string, processor IProcessor, f IPacketFactory) *TcpServer {
	s := new(TcpServer)
	s.Addr = addr
	s.Processor = processor
	s.Factory = f
	s.Clients = make(map[net.Conn]*TcpClient)
	return s
}

func (s *TcpServer) Start() bool {
	addr, _ := net.ResolveTCPAddr("tcp", s.Addr)
	var err error
	s.lsn, err = net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Error("net.Listen() error: %s", err.Error())
		return false
	}
	go s.listen()
	return true
}

func (s *TcpServer) AddClient(conn net.Conn, client *TcpClient) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	s.Clients[conn] = client
	logger.Debug("num of clients is : %d", len(s.Clients))
}

func (s *TcpServer) DelClient(conn net.Conn) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	delete(s.Clients, conn)
	logger.Debug("num of clients is : %d", len(s.Clients))
}

func (s *TcpServer) stop() {
	s.lsn.Close()
	// 关闭所有客户端
	for _, client := range s.Clients {
		client.stop()
	}
	s.clientsWait.Wait()
}

func (s *TcpServer) listen() {
	defer s.stop()
	for {
		conn, err := s.lsn.AcceptTCP()
		if err != nil {
			logger.Error("Listener.Accept() error: %s", err.Error())
			return
		}
		s.clientsMutex.Lock()
		if len(s.Clients) >= MAX_CLIENT_NUM {
			conn.Close()
			logger.Fatal("len[%d] of clients More than maxClientNum!!!", len(s.Clients))
		} else {
			s.clientsWait.Add(1)
			go s.runClient(conn)
		}
		s.clientsMutex.Unlock()
	}
}

func (s *TcpServer) runClient(conn *net.TCPConn) {
	defer s.clientsWait.Done()

	tcpClient := NewTcpClient(conn, s.Processor, s.Factory)
	s.AddClient(conn, tcpClient)
	logger.Debug("client[%s] connected", conn.RemoteAddr().String())

	// 这里会阻塞，直到tcpClient结束才返回
	tcpClient.Run()

	// 这里说明客户端已经关闭，删除结束的客户端
	s.DelClient(conn)
	logger.Debug("client[%s] closed", conn.RemoteAddr().String())
}
