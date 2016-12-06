package main

import (
	"bufio"
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
	conn, err := net.Dial("tcp", bot.Server)
	if err != nil {
		log.Fatal(err)
	}
	bot.Conn = conn
	bot.Reader = bufio.NewReader(conn)
}

func (bot *Bot) reconnect() {
	bot.UnloadAllPlugins()
	bot.Conn.Close()
	bot.connect()
	bot.LoadAllPlugins()
}

func usage(args []string) {
	fmt.Printf("%s SERVER:PORT\n", args[0])
	fmt.Println()
	fmt.Printf("Ex: %s irc.freenode.org:6667\n", args[0])
	os.Exit(2)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		usage(args)
		os.Exit(2)
	}
	if args[1] == "-h" || args[1] == "--help" {
		usage(args)
		os.Exit(0)
	}
	server := args[1]
	bot := &Bot{
		Server: server,
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

	// Main loop
	for {
		line, err := bot.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				bot.reconnect()
			} else {
				fmt.Println(err)
			}
			time.Sleep(2 * time.Second)
		}
		dispatcher(bot, line)
	}
}
