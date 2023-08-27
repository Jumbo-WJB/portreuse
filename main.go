package main

import (
	"context"
    "flag"
	"fmt"
	"net"
	"os"
	"syscall"
	"strings"
    "golang.org/x/sys/unix"
)

var lc = net.ListenConfig{
	Control: func(network, address string, c syscall.RawConn) error {
		var opErr error
		if err := c.Control(func(fd uintptr) {
		   opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		}); err != nil {
			return err
		}
		return opErr
	},
}

// // 定义常量
// const (
// 	localPort  = "8080"    // 本地监听端口
// 	remoteAddr = "43.128.47.230:22" // 远程转发地址
// 	sourceAddr = "112.65.12.104"    // 来源地址
// )

// 定义错误处理函数
func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 定义转发函数
func forward(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			break
		}
		n, err = dst.Write(buf[:n])
		if err != nil {
			break
		}
	}
}

func main() {
	pid := os.Getpid()
    // 定义命令行参数
    var localPort string
    var remoteAddr string
    var sourceAddr string
    var listenAddr string

    // 解析命令行参数
    flag.StringVar(&localPort, "listenport", "80", "local port to listen")
    flag.StringVar(&remoteAddr, "remoteaddr", "2.2.2.2:22", "remote address to forward")
    flag.StringVar(&sourceAddr, "sourceaddr", "1.1.1.1", "source address to filter")
    flag.StringVar(&listenAddr, "listenaddr", "172.19.0.2", "local address to listen")
    flag.Parse()
	l, err := lc.Listen(context.Background(), "tcp", listenAddr+":"+localPort)
	if err != nil {
		panic(err)
	}
	fmt.Printf("TCP Server with PID: %d is running \n", pid)

	for {
        // 接受客户端连接
        conn, err := l.Accept()
        if err != nil {
            break
        }
        fmt.Println("Accepted connection from", conn.RemoteAddr())

        // 判断来源地址是否为1.1.1.1，如果是则转发流量，否则关闭连接
        if strings.HasPrefix(conn.RemoteAddr().String(), sourceAddr) {
            fmt.Println("Forwarding traffic to", remoteAddr)
            // 创建远程连接套接字
            rconn, err := net.Dial("tcp", remoteAddr)
            handleError(err)

            // 启动两个协程，分别转发两个方向的流量
            go forward(conn, rconn)
            go forward(rconn, conn)
        } else {
            fmt.Println("Closing connection from", conn.RemoteAddr())
            conn.Close()
        }
    }
}


