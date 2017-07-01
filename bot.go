package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	Server  string
	ssl     bool
	Conn    net.Conn
	Reader  *bufio.Reader
	plugins []*Plugin
	mutex   *sync.Mutex
}

func dispatcher(bot *Bot, line string) {
	fmt.Print(line)

	for _, plugin := range bot.plugins {
		if plugin != nil {
			fmt.Fprintf(plugin.stdin, "%s\n", line)
		}
	}
}

func (bot *Bot) connect() {
	if bot.ssl {
		conn, err := tls.Dial("tcp", bot.Server, nil)
		if err != nil {
			log.Fatal(err)
		}
		bot.Conn = conn
		bot.Reader = bufio.NewReader(conn)
	} else {
		conn, err := net.Dial("tcp", bot.Server)
		if err != nil {
			log.Fatal(err)
		}
		bot.Conn = conn
		bot.Reader = bufio.NewReader(conn)
	}
}

func (bot *Bot) reconnect() {
	bot.UnloadAllPlugins()
	bot.Conn.Close()
	bot.connect()
	bot.LoadAllPlugins()
}

func main() {
	server := flag.String("server", "chat.freenode.net:6667", "IRC server address to connect to")
	ssl := flag.Bool("ssl", false, "use SSL")
	flag.Parse()

	bot := &Bot{
		Server: *server,
		ssl:    *ssl,
		mutex:  &sync.Mutex{},
	}
	bot.connect()

	// signals
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-sigs
			switch sig {
			case os.Interrupt:
			case syscall.SIGHUP:
				bot.reloadAllPlugins()
			case syscall.SIGUSR1:
				fmt.Println("reconnecting...")
				bot.reconnect()
			}
		}
	}()

	bot.LoadAllPlugins()

	// Stdin loop: allow user to chat with server directly
	// (useful for debugging)
	go func() {
		stdin := bufio.NewReader(os.Stdin)

		for {
			line, _ := stdin.ReadString('\n')
			fmt.Fprintf(bot.Conn, line)
		}

	}()

	// Main loop
	for {
		line, err := bot.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				bot.reconnect()
			} else {
				fmt.Println(err)
				bot.reconnect()
			}
			time.Sleep(2 * time.Second)
		}
		dispatcher(bot, line)
	}
}
