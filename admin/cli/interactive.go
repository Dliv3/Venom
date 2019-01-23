package cli

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dliv3/Venom/admin/dispather"
	"github.com/Dliv3/Venom/node"
)

// admin节点想要操作的对端节点的ID，主要用于goto命令
var currentPeerNodeHashID string

func checkCurrentPeerNode() bool {
	if currentPeerNodeHashID == "" {
		fmt.Println("you should select node first")
		return false
	}
	return true
}

func printNetworkMap() {
	printed := make(map[string]bool)
	fmt.Println("A")
	printed[node.CurrentNode.HashID] = true
	printEachMap(node.CurrentNode.HashID, 0, printed)
}

func printEachMap(nodeID string, depth int, printed map[string]bool) {
	for _, value := range node.GNetworkTopology.NetworkMap[nodeID] {
		if _, ok := printed[value]; ok {
			continue
		}
		for i := 0; i < depth; i++ {
			fmt.Print("     ")
		}
		fmt.Print("+ -- ")
		fmt.Println(node.GNodeInfo.NodeUUID2Number[value])
		printed[value] = true
		printEachMap(value, depth+1, printed)
	}
}

// Interactive 交互式控制
func Interactive() {
	// 处理ctrl c的SIGINT信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT) //, syscall.SIGTERM)
	shellExit := true

	go func() {
		for {
			<-sigs
			if !shellExit {
				// ctrl c 处理函数
				fmt.Println("Ctrl-C")
			} else {
				os.Exit(0)
			}
		}
	}()
	var nodeID int
	var peerNode *node.Node
	for {
		if currentPeerNodeHashID == "" {
			fmt.Print("(admin node)>>> ")
		} else {
			fmt.Printf("(node %d)>>> ", nodeID)
		}
		var cmdStr string
		fmt.Scanf("%s", &cmdStr)
		switch cmdStr {
		case "help":
			ShowUsage()
		case "show":
			dispather.SendSyncCmd()
			printNetworkMap()
		case "setdes":
			var nodeID int
			var description string
			fmt.Scanf("%d", &nodeID)
			reader := bufio.NewReader(os.Stdin)
			descriptionBytes, _, _ := reader.ReadLine()
			description = string(descriptionBytes)
			// description = util.ScanLine()
			if _, ok := node.GNodeInfo.NodeNumber2UUID[nodeID]; ok {
				node.GNodeInfo.NodeDescription[nodeID] = description
			} else {
				fmt.Println("unknown nodeID")
			}
		case "getdes":
			var nodeID int
			fmt.Scanf("%d", &nodeID)
			if _, ok := node.GNodeInfo.NodeDescription[nodeID]; ok {
				fmt.Println(node.GNodeInfo.NodeDescription[nodeID])
			} else {
				fmt.Println("unknown nodeID")
			}
		case "goto":
			var tmpNodeID int
			fmt.Scanf("%d", &tmpNodeID)
			fmt.Println("node", tmpNodeID)
			if _, ok := node.GNodeInfo.NodeNumber2UUID[tmpNodeID]; ok {
				nodeID = tmpNodeID
			} else {
				fmt.Println("unknown nodeID")
				break
			}
			currentPeerNodeHashID = node.GNodeInfo.NodeNumber2UUID[nodeID]
			// nextNodeID := node.GNetworkTopology.RouteTable[currentPeerNodeHashID]
			// nextNode = node.Nodes[nextNodeID]
			peerNode = node.Nodes[currentPeerNodeHashID]
		case "listen":
			if !checkCurrentPeerNode() {
				break
			}
			var port uint16
			fmt.Scanf("%d", &port)
			fmt.Println("port", port)
			if port > 65535 || port < 1 {
				fmt.Println("port number error")
				break
			}
			dispather.SendListenCmd(peerNode, port)
		case "connect":
			if !checkCurrentPeerNode() {
				break
			}
			var ipString string
			var port uint16
			fmt.Scanf("%s %d", &ipString, &port)
			fmt.Println("ip port", ipString, port)
			ip := net.ParseIP(ipString)
			if ip == nil {
				fmt.Println("invalid ip address.")
				break
			}
			dispather.SendConnectCmd(peerNode, ipString, port)
		case "socks":
			if !checkCurrentPeerNode() {
				break
			}
			var port uint16
			fmt.Scanf("%d", &port)
			fmt.Println("port", port)
			if port > 65535 || port < 1 {
				fmt.Println("port number error")
				break
			}
			dispather.SendSocks5Cmd(peerNode, port)
		case "shell":
			if !checkCurrentPeerNode() {
				break
			}
			fmt.Println("You can execute dispather in this shell :D, 'exit' to exit.")
			shellExit = false
			dispather.SendShellCmd(peerNode)
			shellExit = true
		case "upload":
			if !checkCurrentPeerNode() {
				break
			}
			var localPath string
			var remotePath string

			fmt.Scanf("%s %s", &localPath, &remotePath)
			fmt.Println("path", localPath, remotePath)
			dispather.SendUploadCmd(peerNode, localPath, remotePath)
		case "download":
			if !checkCurrentPeerNode() {
				break
			}
			var remotePath string
			var localPath string
			fmt.Scanf("%s %s", &remotePath, &localPath)
			fmt.Println("path", remotePath, localPath)
			dispather.SendDownloadCmd(peerNode, remotePath, localPath)
		case "lforward":
			if !checkCurrentPeerNode() {
				break
			}
			var sport uint16
			var dport uint16
			var lhostString string
			fmt.Scanf("%s %d %d", &lhostString, &sport, &dport)
			lhost := net.ParseIP(lhostString)
			if lhost == nil {
				fmt.Println("invalid ip address.")
				break
			}
			fmt.Printf("forward %s port %d to remote port %d\n", lhostString, sport, dport)
			dispather.SendLForwardCmd(peerNode, sport, lhostString, dport)
		case "rforward":
			if !checkCurrentPeerNode() {
				break
			}
			var sport uint16
			var dport uint16
			var rhostString string
			fmt.Scanf("%s %d %d", &rhostString, &sport, &dport)
			rhost := net.ParseIP(rhostString)
			if rhost == nil {
				fmt.Println("invalid ip address.")
				break
			}
			fmt.Printf("forward %s port %d to local port %d\n", rhostString, sport, dport)
			dispather.SendRForwardCmd(peerNode, rhostString, sport, dport)
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("unknown command, use \"help\" to see all commands.")
		}
	}
}
