## Venom - A Multi-layer Proxy for Attackers

<p>
<a href="README.md">简体中文</a>
<a href="README-en.md">English</a>
</p>

Venom是一款使用Go开发的为渗透测试人员设计的多级代理工具。

Venom可将多个节点进行连接，然后以节点为跳板，构建多级代理。

渗透测试人员可以轻松使用Venom将网络流量代理到多层内网，并轻松地管理代理节点。


> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！


## 特点

- 可视化网络拓扑
- 多级socks5代理
- 多级端口转发
- 端口复用 (apache/mysql/...)
- ssh隧道
- 交互式shell
- 文件的上传和下载
- 支持多种平台(Linux/Windows/MacOS)和多种架构(x86/x64/arm/mips)

> 由于IoT设备（arm/mips/...架构）通常资源有限，为了减小二进制文件的大小，该项目针对IoT环境编译的二进制文件不支持端口复用和ssh隧道这两个功能，并且为了减小内存使用限制了网络并发数和缓冲区大小。

## 使用

### 1. admin/agent命令行参数

- admin节点和agent节点均可建立连接也可发起连接

  admin监听端口，agent发起连接:

  ```bash
  ./admin_macos_x64 -l 9999
  ```

  ```
  ./agent_linux_x64 -c 192.168.0.103 -p 9999
  ```

  agent监听端口，admin发起连接:

  ```
  ./agent_linux_x64 -l 8888
  ```

  ```
  ./admin_macos_x64 -c 192.168.204.139 -p 8888
  ```

- agent节点支持端口复用，可复用如apache、mysql等支持端口复用的服务的端口

  ```
  # 复用apache 80端口，不影响apache提供正常的http服务
  # -h 的值为本机ip，不能写0.0.0.0，否则无法进行端口复用
  ./agent_linux_x64 -h 192.168.204.139 -l 80 -reuse-port
  ```

  ```
  ./admin_macos_x64 -c 192.168.204.139 -p 80
  ```

### 2. admin节点内置命令

- **help** 打印帮助信息

  ```
  (admin node) >>> help
  
    help                                     Help information.
    exit                                     Exit.
    show                                     Display network topology.
    setdes     [id] [info]                   Add a description to the target node.
    getdes     [id]                          View description of the target node.
    goto       [id]                          Select id as the target node.
    listen     [port]                        Listen on a port on the target node.
    connect    [ip] [port]                   Connect to a new node through current node.
    sshconnect [user@ip:port] [dport]        Connect to a new node through ssh tunnel.
    shell                                    Start an interactive shell on the target node.
    upload     [local_file]  [remote_file]   Upload file to the target node.
    download   [remote_file]  [local_file]   Download file from the target node.
    socks      [lport]                       Start a socks server.
    lforward   [lhost] [sport] [dport]       Forward a local sport to a remote dport.
    rforward   [rhost] [sport] [dport]       Forward a remote sport to a local dport.
  
  ```

- **show** 显示网络拓扑

  A表示admin节点，数字表示agent节点

  下面的拓扑图表示，admin节点下连接了1节点，1节点下连接了2、4节点，2节点下连接了3节点

  ```
  (admin node) >>> show
  A
  + -- 1
       + -- 2
            + -- 3
       + -- 4
  ```

- **goto** 操作某节点

  ```
  (admin node) >>> goto 1
  (node 1) >>> 
  ```
  在goto到某节点之后你就可以使用下面将要介绍的命令

- **getdes/setdes** 获取/设置节点信息描述

  ```
  (node 1) >>> setdes linux x64 blahblahblah
  (node 1) >>> getdes
  linux x64 blahblahblah
  ```

- **connect/listen/sshconnect** 节点间互连

  node 1节点连接192.168.0.103的9999端口

  ```
  (node 1) >>> connect 192.168.0.103 9999
  ip port 192.168.0.103 9999
  connect to remote port success!
  (node 1) >>> show
  A
  + -- 1
       + -- 2
  ```
  在node1节点监听9997端口, 然后在另一台机器上运行 ./agent_linux_x64 -c 192.168.204.139 -p 9997 连接node1
  ```
  (node 1) >>> listen 9997
  port 9997
  listen local port success!
  (node 1) >>> show
  A
  + -- 1
       + -- 2
       + -- 3
  
  ```
  node3通过ssh隧道连接192.168.0.104的9999端口。你可以使用密码或者是ssh私钥进行认证。
  ```
  (node 1) >>> goto 3
  (node 3) >>> sshconnect root@192.168.0.104:22 9999
  use password (1) / ssh key (2)?2
  file path of ssh key:/Users/dlive/.ssh/id_rsa
  connect to target host's 9999 through ssh tunnel (root@192.168.0.104:22).
  ssh connect to remote node success!
  (node 3) >>> show
  A
  + -- 1
       + -- 2
       + -- 3
            + -- 4
  ```

- **shell** 获取节点的交互式shell

  ```
  (node 1) >>> shell
  You can execute dispather in this shell :D, 'exit' to exit.
  bash: no job control in this shell
  bash-3.2$ whoami
  whoami
  dlive
  bash-3.2$ exit
  exit
  exit
  ```

- **upload/download** 向节点上传/从节点下载文件

  将本地/tmp/test.pdf上传到node1的/tmp/test2.pdf

  ```
  (node 1) >>> upload /tmp/test.pdf /tmp/test2.pdf
  path /tmp/test.pdf /tmp/test2.pdf
  this file is too large(>100M), still uploading? (y/n)y
   154.29 MiB / 154.23 MiB [========================================] 100.04% 1s
  upload file success!
  ```
  将node1的文件/tmp/test.pdf下载到本地的/tmp/test2.pdf
  ```
  (node 1) >>> download /tmp/test2.pdf /tmp/test3.pdf
  path /tmp/test2.pdf /tmp/test3.pdf
  this file is too large(>100M), still downloading? (y/n)y
   154.23 MiB / 154.23 MiB [========================================] 100.00% 1s
  download file success!
  ```

- **socks** 建立到某节点的socks5代理

  ```
  (node 1) >>> socks 7777
  port 7777
  a socks5 proxy of the target node has started up on local port 7777
  ```

  执行成功socks命令之后，会在admin节点本地开启一个端口，如上述的7777，使用7777即可进行socks5代理

- **lforward/rforward** 将本地端口转发到远程/将远程端口转发到本地

  lforward将admin节点本地的8888端口转发到node1的8888端口

  ```
  (node 1) >>> lforward 127.0.0.1 8888 8888
  forward 127.0.0.1 port 8888 to remote port 8888
  ```

  rforward 将node1网段的192.168.204.103端口8889转发到admin节点本地的8889端口 
  ```
  (node 1) >>> rforward 192.168.204.103 8889 8889
  forward 192.168.204.103 port 8889 to local port 8889
  ```

### 3. 注意事项

- 现阶段仅支持单个admin节点对网络进行管理
- 要对新加入的节点进行操作，需要首先在admin节点运行show命令同步网络拓扑和节点编号

## TODO

- 与regeorg联动
- 多个admin节点同时对网络进行管理
- 节点间通信流量加密
- socks5对udp的支持
- 与meterpreter联动 (待定)
- RESTful API

## 致谢

- [rootkiter#Termite](https://github.com/rootkiter/Termite)
- [ring04h#s5.go](https://github.com/ring04h/s5.go)

