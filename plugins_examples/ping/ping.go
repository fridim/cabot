package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

func main() {
	parentpid, err := os.FindProcess(os.Getppid())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// check last ping and reconnect if needed
	ping := time.Now()
	go func() {
		for {
			if time.Since(ping) > 255*time.Second {
				fmt.Fprintf(os.Stderr, "ping not received from server for 255 sec, SIGUSR1 sent to %d", parentpid.Pid)
				parentpid.Signal(syscall.SIGUSR1)
				ping = time.Now()
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
				fmt.Fprintln(os.Stderr, err)
			}
			time.Sleep(2 * time.Second)
		} else {
			if len(line) > 4 && line[:4] == "PING" {
				fmt.Printf(fmt.Sprintf("PONG%s", line[4:]))
				ping = time.Now()
			}
		}
	}
}
