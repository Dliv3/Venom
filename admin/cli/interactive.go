package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
			fmt.Print("(admin node) >>> ")
		} else {
			fmt.Printf("(node %d) >>> ", nodeID)
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
			if !checkCurrentPeerNode() {
				break
			}
			var description string
			reader := bufio.NewReader(os.Stdin)
			descriptionBytes, _, _ := reader.ReadLine()
			description = string(descriptionBytes)
			node.GNodeInfo.NodeDescription[currentPeerNodeHashID] = description
		case "getdes":
			if !checkCurrentPeerNode() {
				break
			}
			fmt.Println(node.GNodeInfo.NodeDescription[currentPeerNodeHashID])
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
		case "sshconnect":
			// sshconnect user:password@10.1.1.1:22 9999
			if !checkCurrentPeerNode() {
				break
			}
			var sshString string
			var dport uint16
			fmt.Scanf("%s %d", &sshString, &dport)
			var sshUser string
			var sshHost string
			var sshPort uint16
			if parts := strings.Split(sshString, "@"); len(parts) > 1 {
				sshUser = parts[0]
				sshHost = parts[1]
			}
			parts := strings.Split(sshHost, ":")
			sshHost = parts[0]
			if len(parts) > 1 {
				tmp, _ := strconv.Atoi(parts[1])
				sshPort = uint16(tmp)
			} else if len(parts) == 1 {
				sshPort = 22
			}
			if net.ParseIP(sshHost) == nil {
				fmt.Println("invalid ssh server ip address.")
				break
			}
			fmt.Print("use password (1) / ssh key (2)? ")
			var choice uint16
			fmt.Scanf("%d", &choice)
			switch choice {
			case 1:
				fmt.Print("password: ")
				var password string
				fmt.Scanf("%s", &password)
				fmt.Printf("connect to target host's %d through ssh tunnel (%s@%s:%d).\n", dport, sshUser, sshHost, sshPort)
				dispather.SendSshConnectCmd(peerNode, sshUser, sshHost, sshPort, dport, choice, password)
			case 2:
				fmt.Print("file path of ssh key: ")
				var path string
				fmt.Scanf("%s", &path)
				sshKey, err := ioutil.ReadFile(path)
				if err != nil {
					fmt.Println("ssh key error:", err)
					break
				}
				fmt.Printf("connect to target host's %d through ssh tunnel (%s@%s:%d).\n", dport, sshUser, sshHost, sshPort)
				dispather.SendSshConnectCmd(peerNode, sshUser, sshHost, sshPort, dport, choice, string(sshKey))
			default:
				fmt.Println("unknown choice.")
				break
			}
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("unknown command, use \"help\" to see all commands.")
		}
	}
}
