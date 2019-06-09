package cli

const (
	LISTEN_MODE  = 1
	CONNECT_MODE = 2
)

const (
	SOCKET_REUSE_METHOD = 1
	IPTABLES_METHOD     = 2
)

type Option struct {
	LocalPort  int
	LocalIP    string
	RemoteIP   string
	RemotePort int
	ReusedPort int
	// 0 默认值，表示参数解析错误，无法设置模式
	// mode 1 listen a local port
	// mode 2 connect to remote port
	Mode int
	// 端口复用方法
	// 1 通过SO_REUSEADDR、SO_REUSEPORT进行端口复用
	// 2 通过iptables端口复用
	PortReuseMethod int

	Password string
}

// Args
var Args Option
