package dispather

import (
	"fmt"
	"net"
	"time"

	"github.com/Dliv3/Venom/netio"
	"golang.org/x/crypto/ssh"
)

const TIMEOUT = 5

func BuiltinSshConnectCmd(sshUser string, sshHost string, sshPort uint16, dport uint16, sshAuthMethod uint16, sshAuthData string) {
	var auth ssh.AuthMethod
	if sshAuthMethod == 1 {
		auth = ssh.Password(sshAuthData)
	} else if sshAuthMethod == 2 {
		key, err := ssh.ParsePrivateKey([]byte(sshAuthData))
		if err != nil {
			fmt.Println("ssh failed to connects to the remote node !")
			fmt.Println("ssh key error!")
			return
		}
		auth = ssh.PublicKeys(key)
	}

	config := ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Duration(time.Second * TIMEOUT),
	}

	sshClient, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", sshHost, sshPort),
		&config,
	)
	if err != nil {
		fmt.Println("ssh failed to connects to the remote node !")
		fmt.Printf("ssh connection error: %s\n", err)
		return
	}

	nodeConn, err := sshClient.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dport))
	if err != nil {
		fmt.Println("ssh failed to connects to the remote node !")
		fmt.Printf("ssh connect to target node error: %s\n", err)
		return
	}
	AdminClient(nodeConn)
}

func BuiltinListenCmd(port uint16) {
	err := netio.InitNode(
		"listen",
		fmt.Sprintf("0.0.0.0:%d", port),
		AdminServer, false, 0)
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to open the port %d!", port))
		return
	}
	fmt.Printf("the port %d is successfully listening on the remote node!\n", port)
}

func BuiltinConnectCmd(ipString string, port uint16) {
	err := netio.InitNode(
		"connect",
		fmt.Sprintf("%s:%d", ipString, port),
		AdminClient, false, 0)
	if err != nil {
		fmt.Println("failed to connect to the remote node!")
		fmt.Println(err)
		return
	}
	fmt.Println("successfully connect to the remote node!")
}
