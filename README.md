## Venom - A Multi-layer Proxy for Attackers


<p> README: 
<a href="README.md">简体中文</a>
<a href="README-en.md">English</a>
</p>
Venom是一款使用Go开发的为渗透测试人员设计的多级代理工具。

渗透测试人员可以轻松使用Venom将网络流量代理到多层内网，并轻松地管理代理节点。

![admin-node](docs/img/admin.png)

> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！
> 


## 特点

- 提供可视化网络拓扑
- 支持多级socks5代理
- 支持多级端口转发
- 支持端口复用 (apache/mysql/...)
- 支持节点间通过ssh隧道建立连接
- 支持交互式shell
- 支持文件的上传和下载
- 支持多种平台(Linux/Windows/MacOS)和多种架构(x86/x64/arm/mips)

> 由于IoT设备（arm/mips/...架构）通常资源有限，为了减小二进制文件的大小，该项目针对IoT的二进制文件不支持端口复用和ssh隧道这两个功能，并且为了减小内存使用限制了网络并发数和缓冲区大小。

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
  # -h 之后的参数需要写本机ip，不能写0.0.0.0，否则无法进行端口复用
  ./agent_linux_x64 -h 192.168.204.139 -l 80 -reuse-port
  ```

  ```
  ./admin_macos_x64 -c 192.168.204.139 -p 80
  ```

### 2. admin节点内置命令

- help 打印帮助信息

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

- show 显示网络拓扑

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

- goto 操作某节点

  ```
  (admin node) >>> goto 1
  (node 1) >>> 
  # 在goto到某节点之后你就可以使用下面将要介绍的命令
  ```

- getdes/setdes 获取/设置节点信息描述

  ```
  (node 1) >>> setdes linux x64 blahblahblah
  (node 1) >>> getdes
  linux x64 blahblahblah
  ```

- connect/listen/sshconnect 节点间互连

  ```
  
  ```

- shell 获取节点的交互式shell

- upload/download 向节点上传/从节点下载文件

- socks 建立到某节点的socks5代理

- lforward/rforward 将本地端口转发到远程/将远程端口转发到本地

### 3. 注意事项

- 现阶段仅支持网络中存在一个admin节点对网络进行管理
- 要对新加入的节点进行操作，需要首先在admin节点运行show命令同步网络拓扑和节点编号分配

## TODO

- 与regeorg联动
- 支持多个管理节点同时对网络进行管理
- 节点间通信流量加密
- socks5对udp的支持
- 与meterpreter联动 (待定)
- RESTful API

## 致谢

- [rootkiter#Termite](https://github.com/rootkiter/Termite)
- [ring04h#s5.go](https://github.com/ring04h/s5.go)

