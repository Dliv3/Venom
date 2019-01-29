// +build linux
// +build amd64 386

package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// reference: https://www.freebuf.com/articles/network/137683.html

// sudo iptables -t nat -A PREROUTING -p tcp --dst 192.168.204.134 --dport 80 -j REDIRECT --to-port 9999
// sudo iptables -t nat -D PREROUTING -p tcp --dst 192.168.204.134 --dport 80 -j REDIRECT --to-port 9999
// sudo iptables -L -t nat

// iptables -t nat -N VENOM
// iptables -t nat -A VENOM -p tcp -j REDIRECT --to-port 8080
// iptables -A INPUT -p tcp -m string --string 'venomcoming' --algo bm -m recent --set --name venom --rsource -j ACCEPT
// iptables -A INPUT -p tcp -m string --string 'venomleaving' --algo bm -m recent --name venom --remove -j ACCEPT
// iptables -t nat -A PREROUTING -p tcp --dst 192.168.1.18 --dport 80 --syn -m recent --rcheck --seconds 3600 --name venom --rsource -j VENOM

// iptables -t nat -D PREROUTING -p tcp --dst 192.168.1.18 --dport 80 --syn -m recent --rcheck --seconds 3600 --name venom --rsource -j VENOM
// iptables -D INPUT -p tcp -m string --string 'venomleaving' --algo bm -m recent --name venom --remove -j ACCEPT
// iptables -D INPUT -p tcp -m string --string 'venomcoming' --algo bm -m recent --set --name venom --rsource -j ACCEPT
// iptables -t nat -F VENOM
// iptables -t nat -X VENOM

const CHAIN_NAME = "VENOM"
const START_FORWARDING = "venomleaving"
const STOP_FORWARDING = "venomcoming"

var INVALID_IP_ADDR = errors.New("invalid ip address.")
var CMD_EXEC_FAIDED = errors.New("iptables command exec failed.")

func DeletePortReuseRules(localPort uint16, reusedPort uint16) error {

	var cmds []string
	cmds = append(cmds, fmt.Sprintf("iptables -t nat -D PREROUTING -p tcp --dport %d --syn -m recent --rcheck --seconds 3600 --name %s --rsource -j %s", reusedPort, strings.ToLower(CHAIN_NAME), CHAIN_NAME))
	cmds = append(cmds, fmt.Sprintf("iptables -D INPUT -p tcp -m string --string %s --algo bm -m recent --name %s --remove -j ACCEPT", STOP_FORWARDING, strings.ToLower(CHAIN_NAME)))
	cmds = append(cmds, fmt.Sprintf("iptables -D INPUT -p tcp -m string --string %s --algo bm -m recent --set --name %s --rsource -j ACCEPT", START_FORWARDING, strings.ToLower(CHAIN_NAME)))
	cmds = append(cmds, fmt.Sprintf("iptables -t nat -F %s", CHAIN_NAME))
	cmds = append(cmds, fmt.Sprintf("iptables -t nat -X %s", CHAIN_NAME))

	for _, each := range cmds {
		cmd := strings.Split(each, " ")
		err := exec.Command(cmd[0], cmd[1:]...).Run()
		if err != nil {
			log.Println("[!]Use '" + each + "' to delete iptables rules.")
		}
	}

	fmt.Println("[+]Delete iptables port reuse rules.")
	return nil
}

func SetPortReuseRules(localPort uint16, reusedPort uint16) error {

	sigs := make(chan os.Signal, 1)

	// 处理不了sigkill, 所以如果程序被kill -9杀掉，需要手动删除iptables规则
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			<-sigs
			DeletePortReuseRules(localPort, reusedPort)
			os.Exit(0)
		}
	}()

	var cmds []string
	cmds = append(cmds, fmt.Sprintf("iptables -t nat -N %s", CHAIN_NAME))
	cmds = append(cmds, fmt.Sprintf("iptables -t nat -A %s -p tcp -j REDIRECT --to-port %d", CHAIN_NAME, localPort))
	cmds = append(cmds, fmt.Sprintf("iptables -A INPUT -p tcp -m string --string %s --algo bm -m recent --set --name %s --rsource -j ACCEPT", START_FORWARDING, strings.ToLower(CHAIN_NAME)))
	cmds = append(cmds, fmt.Sprintf("iptables -A INPUT -p tcp -m string --string %s --algo bm -m recent --name %s --remove -j ACCEPT", STOP_FORWARDING, strings.ToLower(CHAIN_NAME)))
	cmds = append(cmds, fmt.Sprintf("iptables -t nat -A PREROUTING -p tcp --dport %d --syn -m recent --rcheck --seconds 3600 --name %s --rsource -j %s", reusedPort, strings.ToLower(CHAIN_NAME), CHAIN_NAME))

	for _, each := range cmds {
		fmt.Println(each)
		cmd := strings.Split(each, " ")
		err := exec.Command(cmd[0], cmd[1:]...).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
