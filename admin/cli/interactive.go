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
	"github.com/Dliv3/Venom/utils"
)

// admin节点想要操作的对端节点的ID，主要用于goto命令
var currentPeerNodeHashID string

// need code refactoring
func isAdmin() bool {
	if currentPeerNodeHashID == node.CurrentNode.HashID {
		return true
	}
	return false
}

// need code refactoring
// checkPeerNodeIsVaild 检查对端节点是否合法
func checkPeerNodeIsVaild() bool {
	if currentPeerNodeHashID == node.CurrentNode.HashID {
		return true
	} else if node.Nodes[currentPeerNodeHashID] == nil {
		fmt.Println("the node is disconnected.")
		return false
	}
	return true
}

// need code refactoring
// checkPeerNodeIsSelected 检查是否选择对端节点
func checkPeerNodeIsSelected() bool {
	if currentPeerNodeHashID == node.CurrentNode.HashID {
		fmt.Println("you should choose the node first")
		return false
	} else if node.Nodes[currentPeerNodeHashID] == nil {
		fmt.Println("the node is disconnected.")
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
	// init
	currentPeerNodeHashID = node.CurrentNode.HashID
	for {
		if currentPeerNodeHashID == node.CurrentNode.HashID {
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
			if !checkPeerNodeIsSelected() {
				break
			}
			var description string
			reader := bufio.NewReader(os.Stdin)
			descriptionBytes, _, _ := reader.ReadLine()
			description = string(descriptionBytes)
			node.GNodeInfo.NodeDescription[currentPeerNodeHashID] = description
		case "getdes":
			if !checkPeerNodeIsSelected() {
				break
			}
			fmt.Println(node.GNodeInfo.NodeDescription[currentPeerNodeHashID])
		case "goto":
			// need code refactoring
			var tmpNodeID int
			fmt.Scanf("%d", &tmpNodeID)
			if tmpNodeID == 0 {
				// admin
				currentPeerNodeHashID = node.CurrentNode.HashID
				break
			} else if _, ok := node.GNodeInfo.NodeNumber2UUID[tmpNodeID]; ok {
				nodeID = tmpNodeID
			} else {
				fmt.Println("unknown node id.")
				break
			}
			currentPeerNodeHashID = node.GNodeInfo.NodeNumber2UUID[nodeID]
			// nextNodeID := node.GNetworkTopology.RouteTable[currentPeerNodeHashID]
			// nextNode = node.Nodes[nextNodeID]
			peerNode = node.Nodes[currentPeerNodeHashID]
		case "listen":
			var port uint16
			fmt.Scanf("%d", &port)
			fmt.Println("listen port", port)
			if port > 65535 || port < 1 {
				fmt.Println("port number error.")
				break
			}
			if checkPeerNodeIsVaild() {
				if isAdmin() {
					dispather.BuiltinListenCmd(port)
				} else {
					dispather.SendListenCmd(peerNode, port)
				}
			}
		case "connect":
			var ipString string
			var port uint16
			fmt.Scanf("%s %d", &ipString, &port)
			fmt.Println("connect to", ipString, port)
			ip := net.ParseIP(ipString)
			if ip == nil {
				fmt.Println("invalid ip address.")
				break
			}
			if checkPeerNodeIsVaild() {
				if isAdmin() {
					dispather.BuiltinConnectCmd(ipString, port)
				} else {
					dispather.SendConnectCmd(peerNode, ipString, port)
				}
			}
		case "socks":
			if !checkPeerNodeIsSelected() {
				break
			}
			var port uint16
			fmt.Scanf("%d", &port)
			if port > 65535 || port < 1 {
				fmt.Println("port number error.")
				break
			}
			dispather.SendSocks5Cmd(peerNode, port)
		case "shell":
			if !checkPeerNodeIsSelected() {
				break
			}
			utils.HandleWindowsCR()
			fmt.Println("You can execute commands in this shell :D, 'exit' to exit.")
			shellExit = false
			dispather.SendShellCmd(peerNode)
			shellExit = true
			continue
		case "upload":
			if !checkPeerNodeIsSelected() {
				break
			}
			var localPath string
			var remotePath string

			fmt.Scanf("%s %s", &localPath, &remotePath)
			fmt.Println("upload", localPath, fmt.Sprintf("to node %d:", nodeID), remotePath)
			dispather.SendUploadCmd(peerNode, localPath, remotePath)
		case "download":
			if !checkPeerNodeIsSelected() {
				break
			}
			var remotePath string
			var localPath string
			fmt.Scanf("%s %s", &remotePath, &localPath)
			fmt.Println("download", localPath, fmt.Sprintf("from node %d:", nodeID), remotePath)
			dispather.SendDownloadCmd(peerNode, remotePath, localPath)
		case "lforward":
			if !checkPeerNodeIsSelected() {
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
			fmt.Printf("forward local network %s port %d to remote port %d\n", lhostString, sport, dport)
			dispather.SendLForwardCmd(peerNode, sport, lhostString, dport)
		case "rforward":
			if !checkPeerNodeIsSelected() {
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
			fmt.Printf("forward remote network %s port %d to local port %d\n", rhostString, sport, dport)
			dispather.SendRForwardCmd(peerNode, rhostString, sport, dport)
		case "sshconnect":
			// sshconnect user:password@10.1.1.1:22 9999
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
			if checkPeerNodeIsVaild() {
				switch choice {
				case 1:
					fmt.Print("password: ")
					var password string
					fmt.Scanf("%s", &password)
					fmt.Printf("connect to target host's %d through ssh tunnel (%s@%s:%d).\n", dport, sshUser, sshHost, sshPort)
					if isAdmin() {
						dispather.BuiltinSshConnectCmd(sshUser, sshHost, sshPort, dport, choice, password)
					} else {
						dispather.SendSshConnectCmd(peerNode, sshUser, sshHost, sshPort, dport, choice, password)
					}
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
					if isAdmin() {
						dispather.BuiltinSshConnectCmd(sshUser, sshHost, sshPort, dport, choice, string(sshKey))
					} else {
						dispather.SendSshConnectCmd(peerNode, sshUser, sshHost, sshPort, dport, choice, string(sshKey))
					}
				default:
					fmt.Println("unknown choice.")
					break
				}
			}
		case "exit":
			os.Exit(0)
		case "":
			continue
		default:
			fmt.Println("unknown command, use \"help\" to see all commands.")
		}
		utils.HandleWindowsCR()
	}
}
