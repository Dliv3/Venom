## Venom - A Multi-layer Proxy for Attackers

<p>
<a href="README.md">简体中文</a>
<a href="README-en.md">English</a>
</p>

Venom是为渗透测试人员设计开发的一个多级代理工具。

渗透测试人员可以轻松使用Venom将网络流量代理到多层内网，并轻松地管理代理节点。

> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！


## 特点

- 提供可视化网络拓扑
- 支持多级socks5代理
- 支持多级端口转发
- 支持端口复用 (apache/nginx/mysql ...)
- 支持节点间通过ssh隧道建立连接
- 支持交互式shell
- 支持文件的上传和下载
- 支持多种平台(Linux/Windows/MacOS)和多种架构(x86/x64/arm/mips)

> 由于IoT设备（arm/mips/...架构）通常资源有限，为了减小二进制文件的大小，该项目针对IoT的二进制文件不支持端口复用和ssh隧道这两个功能，并且通过减小网络并发数和缓冲区大小较少内存使用。

## 使用



## TODO

- 与regeorg联动
- 支持多个管理节点同时对网络进行管理
- 节点间通信流量加密
- socks5对udp的支持
- 与meterpreter联动 (待定)

## 致谢

- [rootkiter#Termite](https://github.com/rootkiter/Termite)
- [ring04h#s5.go](https://github.com/ring04h/s5.go)

