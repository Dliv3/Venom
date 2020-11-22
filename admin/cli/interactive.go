package cli

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Dliv3/Venom/admin/dispather"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/utils"

	"github.com/mattn/go-tty"
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
	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT) //, syscall.SIGTERM)
	// shellExit := true

	// go func() {
	// 	for {
	// 		<-sigs
	// 		if !shellExit {
	// 			// ctrl c 处理函数
	// 			fmt.Println("Ctrl-C")
	// 		} else {
	// 			os.Exit(0)
	// 		}
	// 	}
	// }()
	var nodeID int
	var peerNode *node.Node
	// init
	currentPeerNodeHashID = node.CurrentNode.HashID

	t, err := tty.Open()

	if err != nil {
		fmt.Println("Failed to open tty")
		os.Exit(-1)
	}
	defer t.Close()

	for {
		var header string

		if currentPeerNodeHashID == node.CurrentNode.HashID {
			header = "(admin node) >>> "
		} else {
			header = fmt.Sprintf("(node %d) >>> ", nodeID)
		}
		var line string
		if line, err = utils.ReadLine(t, header); err != nil {
			continue
		}

		cmdStr := strings.Split(line, " ")
		if len(cmdStr) == 0 {
			fmt.Println("Unknown command")
			continue
		}

		switch cmdStr[0] {
		case "help":
			ShowUsage()
		case "show":
			dispather.SendSyncCmd()
			printNetworkMap()
		case "setdes":
			if !checkPeerNodeIsSelected() {
				break
			}

			if len(cmdStr) == 1 {
				fmt.Println("setdes [info]")
				continue
			}

			var description string
			description = strings.Join(cmdStr[1:], " ")
			node.GNodeInfo.NodeDescription[currentPeerNodeHashID] = description
		case "getdes":
			if !checkPeerNodeIsSelected() {
				break
			}
			fmt.Println(node.GNodeInfo.NodeDescription[currentPeerNodeHashID])
		case "goto":
			if len(cmdStr) != 2 {
				fmt.Println("goto [id]")
				continue
			}

			var tmpNodeID int
			if tmpNodeID, err = strconv.Atoi(cmdStr[1]); err != nil {
				fmt.Printf("Bad nodeid %s\n", cmdStr[1])
				continue
			}

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
			if len(cmdStr) != 2 {
				fmt.Println("listen [lport]")
				continue
			}

			var port uint16
			var value uint64

			if value, err = strconv.ParseUint(cmdStr[1], 10, 16); err != nil {
				fmt.Printf("Bad port %s\n", cmdStr[1])
				continue
			}
			port = uint16(value)

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
			if len(cmdStr) != 3 {
				fmt.Println("connect [rhost] [rport]")
				continue
			}

			var ipString string
			var port uint16
			var value uint64

			ipString = cmdStr[1]
			if value, err = strconv.ParseUint(cmdStr[2], 10, 16); err != nil {
				fmt.Printf("Bad port %s\n", cmdStr[1])
				continue
			}
			port = uint16(value)

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

			if len(cmdStr) != 2 {
				fmt.Println("socks [lport]")
				continue
			}

			var port uint16
			var value uint64

			if value, err = strconv.ParseUint(cmdStr[1], 10, 16); err != nil {
				fmt.Printf("Bad port %s\n", cmdStr[1])
				continue
			}
			port = uint16(value)

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
			// shellExit = false
			dispather.SendShellCmd(peerNode)
			// shellExit = true
			continue
		case "upload":
			if !checkPeerNodeIsSelected() {
				break
			}

			if len(cmdStr) != 3 {
				fmt.Println("upload [local_file] [remote_file]")
				continue
			}

			var localPath string
			var remotePath string

			localPath = cmdStr[1]
			remotePath = cmdStr[2]
			fmt.Println("upload", localPath, fmt.Sprintf("to node %d:", nodeID), remotePath)
			dispather.SendUploadCmd(peerNode, localPath, remotePath)
		case "download":
			if !checkPeerNodeIsSelected() {
				break
			}

			if len(cmdStr) != 3 {
				fmt.Println("upload [local_file] [remote_file]")
				continue
			}

			var remotePath string
			var localPath string

			remotePath = cmdStr[1]
			localPath = cmdStr[2]
			fmt.Println("download", localPath, fmt.Sprintf("from node %d:", nodeID), remotePath)
			dispather.SendDownloadCmd(peerNode, remotePath, localPath)
		case "lforward":
			if !checkPeerNodeIsSelected() {
				break
			}

			if len(cmdStr) != 4 {
				fmt.Println("lforward [lhost] [sport] [dport]")
				continue
			}

			var sport uint16
			var dport uint16
			var lhostString string
			var value uint64

			if value, err = strconv.ParseUint(cmdStr[1], 10, 16); err != nil {
				fmt.Printf("Bad sport %s\n", cmdStr[1])
				continue
			}
			sport = uint16(value)

			if value, err = strconv.ParseUint(cmdStr[2], 10, 16); err != nil {
				fmt.Printf("Bad dport %s\n", cmdStr[2])
				continue
			}
			dport = uint16(value)

			lhostString = cmdStr[3]
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

			if len(cmdStr) != 4 {
				fmt.Println("rforward [rhost] [sport] [dport]")
				continue
			}

			var sport uint16
			var dport uint16
			var rhostString string
			var value uint64

			rhostString = cmdStr[1]
			if value, err = strconv.ParseUint(cmdStr[2], 10, 16); err != nil {
				fmt.Printf("Bad sport %s\n", cmdStr[2])
				continue
			}
			sport = uint16(value)

			if value, err = strconv.ParseUint(cmdStr[3], 10, 16); err != nil {
				fmt.Printf("Bad dport %s\n", cmdStr[3])
				continue
			}
			dport = uint16(value)

			rhost := net.ParseIP(rhostString)
			if rhost == nil {
				fmt.Println("invalid ip address.")
				break
			}
			fmt.Printf("forward remote network %s port %d to local port %d\n", rhostString, sport, dport)
			dispather.SendRForwardCmd(peerNode, rhostString, sport, dport)
		case "sshconnect":
			// sshconnect user:password@10.1.1.1:22 9999
			if len(cmdStr) != 3 {
				fmt.Println("sshconnect [user@ip:port] [dport]")
				continue
			}

			var sshString string
			var dport uint16
			var value uint64

			sshString = cmdStr[1]

			if value, err = strconv.ParseUint(cmdStr[2], 10, 16); err != nil {
				fmt.Printf("Bad dport %s\n", cmdStr[2])
				continue
			}
			dport = uint16(value)

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

			var choice_s string
			var choice uint16

			if choice_s, err = t.ReadString(); err != nil {
				fmt.Println("Error to read choice")
				continue
			}

			if value, err = strconv.ParseUint(choice_s, 10, 16); err != nil {
				fmt.Printf("Bad choice %s\n", choice_s)
				continue
			}
			choice = uint16(value)

			if checkPeerNodeIsVaild() {
				switch choice {
				case 1:
					fmt.Print("password: ")
					var password string

					if password, err = t.ReadString(); err != nil {
						fmt.Println("Error to read password")
						continue
					}

					fmt.Printf("connect to target host's %d through ssh tunnel (%s@%s:%d).\n", dport, sshUser, sshHost, sshPort)
					if isAdmin() {
						dispather.BuiltinSshConnectCmd(sshUser, sshHost, sshPort, dport, choice, password)
					} else {
						dispather.SendSshConnectCmd(peerNode, sshUser, sshHost, sshPort, dport, choice, password)
					}
				case 2:
					fmt.Print("file path of ssh key: ")
					var path string
					if path, err = t.ReadString(); err != nil {
						fmt.Println("Error to read path")
						continue
					}
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
					fmt.Println("Unknown choice.")
					break
				}
			}
		case "exit":
			t.Close()
			os.Exit(0)
		case "":
			continue
		default:
			fmt.Printf("Unknown command %s, use \"help\" to see all commands.\n", cmdStr[0])
		}
		utils.HandleWindowsCR()
	}
}
