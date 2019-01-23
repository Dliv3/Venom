## Venom - A Multistage Proxy for Attackers

You can easily use venom to automatically proxy your network traffic to a multi-layer intranet, and easily manage intranet nodes.

> This tool is limited to security research and teaching, and the user bears all legal and related responsibilities caused by the use of this tool! The author does not assume any legal and related responsibilities!


## Features

- multistage socks5 proxy
- multistage port forward
- network topology
- auto routing
- port reuse (apache/nginx/mysql ...)
- ssh tunnel
- interactive shell
- upload and download file
- supports multiple platforms(Linux/Windows/MacOS) and multiple architectures(x86/x64/arm/mips)

> For IoT devices (arm/mips/...), I removed the port reuse and ssh tunnel features to reduce file size of agent, and reduce the value of some global variables to reduce memory usage of agent.

## Usage



## TODO

- combined with regeorg
- multiple administrator nodes
- network traffic encryption
- socks5 udp support
- combined with meterpreter (to be discussed)

## Acknowledgement

- [rootkiter#Termite](https://github.com/rootkiter/Termite)
- [ring04h#s5.go](https://github.com/ring04h/s5.go)

