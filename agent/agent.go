package main

import (
	"runtime"

	"github.com/Dliv3/Venom/agent/cli"
	"github.com/Dliv3/Venom/agent/dispather"
	"github.com/Dliv3/Venom/agent/initnode"
	"github.com/Dliv3/Venom/node"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cli.ParseArgs()

	node.CurrentNode.IsAdmin = 0
	node.CurrentNode.InitCommandBuffer()
	node.CurrentNode.InitDataBuffer()

	// fmt.Println(node.CurrentNode.HashID)

	dispather.InitAgentHandler()

	initnode.InitNode()
}
