package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

//客户端测试程序：
//1、服务端的地址和创建客户端的数量通关参数传递 Usage:./client 127.0.0.1:8000 200
//2、每个客户端启动两个协程，一个发送，一个接受
//3、每5秒发送一次，请求当前服务端时间

func Send(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		fmt.Printf("Send() done!!\n")
	}()
	for {
		p := NewPacket()
		p.WriteBegin(CLIENT_COMMAND_TIME_REQ, 0)
		p.WriteEnd()
		_, err := conn.Write(p.GetData())
		if nil != err {
			fmt.Printf("c.Write() failed, err:%s", err.Error())
			return
		}
		time.Sleep(5 * time.Second)
	}

}

func Recv(conn net.Conn, wg *sync.WaitGroup, authFlag chan int32) {
	defer func() {
		wg.Done()
		fmt.Printf("Recv() done!!\n")
	}()

	for {
		p := NewPacket()
		// 读取包头
		head := make([]byte, p.GetHeadLen())
		_, err := io.ReadFull(conn, head)
		if nil != err {
			fmt.Printf("io.ReadFull() failed, error[%s]", err.Error())
			return
		}
		p.WriteBytes(head)

		// 读取包体
		data := make([]byte, p.GetBodyLen())
		_, err = io.ReadFull(conn, data)
		if nil != err {
			fmt.Printf("io.ReadFull() failed, error[%s]", err.Error())
			return
		}
		p.WriteBytes(data)
		p.WriteEnd()

		// 解析包的内容
		cmd := p.GetCmd()
		switch cmd {
		case SERVER_COMMAND_TIME_RESP:
			curTime := p.ReadString()
			fmt.Printf("[%s] Server resp time=%s\n", conn.LocalAddr(), curTime)
		case SERVER_COMMAND_AUTH_SUCCESS:
			authResult := p.ReadInt32()
			authFlag <- authResult
			if 1 == authResult {
				fmt.Printf("[%s] login auth success\n", conn.LocalAddr())
			} else {
				fmt.Printf("[%s] login auth failed\n", conn.LocalAddr())
			}
		default:
			fmt.Printf("cmd[%x] is error", cmd)
		}
	}
}

func CreateConnect(num int, addr string, wgs *sync.WaitGroup, isAuth bool) {
	defer func() {
		wgs.Done()
		fmt.Printf("CreateConnect() done!!\n")
	}()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("net.Dail(%s) failed, err:%s\n", addr, err.Error())
		return
	}
	fmt.Printf("client[%d] connect success\n", num)

	// 发送登录包
	p := NewPacket()
	p.WriteBegin(CLIENT_COMMAND_LOGIN_AUTH, 0)
	if isAuth {
		p.WriteInt32(AUTH_KEY)
	} else {
		// 如果不认证，这里写一个错误的key，服务端认证不成功，将关闭连接
		p.WriteInt32(0)
	}
	p.WriteEnd()
	_, err = c.Write(p.GetData())
	if nil != err {
		fmt.Printf("c.Write() failed, err:%s\n", err.Error())
		return
	}

	var wg sync.WaitGroup
	authFlag := make(chan int32)

	// 启动接受协程
	wg.Add(1)
	go Recv(c, &wg, authFlag)

	// 认证成功后，启动发送协程
	authResult := <-authFlag
	if 1 == authResult {
		wg.Add(1)
		go Send(c, &wg)
	}

	wg.Wait()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:./client 127.0.0.1:8000 200")
		return
	}
	clientNum, _ := strconv.Atoi(os.Args[2])
	var wg sync.WaitGroup
	for i := 0; i < clientNum; i++ {
		wg.Add(1)
		go CreateConnect(i+1, os.Args[1], &wg, true)
	}
	wg.Wait()
}
