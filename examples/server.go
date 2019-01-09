package main

import (
	"fmt"
	"os"
	"runtime"

	solidnet "github.com/idakun/solidnet"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:./server 0.0.0.0:8000")
		return
	}
	addr := os.Args[1]
	runtime.GOMAXPROCS(1)
	solidnet.Run(solidnet.NewGame(addr, "testserver", ".", NewHandler(), NewPacketFactory()))
}
