package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
)

func main() {
	parentpid, err := os.FindProcess(os.Getppid())
	if err != nil {
		log.Fatal(err)
	}

	// check last ping and reconnect if needed
	ping := time.Now()
	go func() {
		for {
			if time.Since(ping) > 255*time.Second {
				parentpid.Signal(syscall.SIGUSR1)
			}
			time.Sleep(100 * time.Second)
		}
	}()

	bio := bufio.NewReader(os.Stdin)

	for {
		line, err := bio.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				parentpid.Signal(syscall.SIGUSR1)
			} else {
				fmt.Println(err)
			}
			time.Sleep(2 * time.Second)
		} else {
			if len(line) > 4 && line[:4] == "PING" {
				fmt.Printf(strings.Replace(line, "PING", "PONG", 1))
				ping = time.Now()
			}
		}
	}
}
