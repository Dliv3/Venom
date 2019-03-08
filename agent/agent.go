package main

import (
	"os/signal"
	"runtime"
	"syscall"

	"github.com/Dliv3/Venom/agent/cli"
	"github.com/Dliv3/Venom/agent/dispather"
	initnode "github.com/Dliv3/Venom/agent/init"
	"github.com/Dliv3/Venom/node"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cli.ParseArgs()

	// ignore sighup
	signal.Ignore(syscall.SIGHUP)

	node.CurrentNode.IsAdmin = 0
	node.CurrentNode.InitCommandBuffer()
	node.CurrentNode.InitDataBuffer()

	dispather.InitAgentHandler()

	initnode.InitNode()
}
