package main

import (
    "fmt"
    "net"
    "os"
	"io"
)


func main() {
	l, err := net.Listen("tcp", "127.0.0.1:3333")
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    defer l.Close()

    conn, err := l.Accept()
    if err != nil {
        fmt.Println("Error accepting: ", err.Error())
        os.Exit(1)
    }
	defer conn.Close()

	go copyFromClient(conn)
	go copyToClient(conn)

	for {
	}
}

func copyToClient(conn net.Conn) {
	io.Copy(conn, os.Stdin)
}

func copyFromClient(conn net.Conn) {
	io.Copy(os.Stdout, conn)
}
