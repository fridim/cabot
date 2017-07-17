package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	Server         string
	ssl            bool
	Conn           net.Conn
	Reader         *bufio.Reader
	plugins        []*Plugin
	mutex          *sync.Mutex
	reconnectMutex *sync.Mutex
	pluginsMutex   *sync.Mutex
}

func dispatcher(bot *Bot, line string) {
	fmt.Print(line)

	bot.pluginsMutex.Lock()
	for _, plugin := range bot.plugins {
		if plugin != nil {
			fmt.Fprintln(plugin.stdin, line)
		}
	}
	bot.pluginsMutex.Unlock()
}

func (bot *Bot) connect() error {
	if bot.ssl {
		conn, err := tls.Dial("tcp", bot.Server, nil)
		if err != nil {
			return err
		}
		bot.Conn = conn
		bot.Reader = bufio.NewReader(conn)
	} else {
		conn, err := net.Dial("tcp", bot.Server)
		if err != nil {
			return err
		}
		bot.Conn = conn
		bot.Reader = bufio.NewReader(conn)
	}
	return nil
}
func (bot *Bot) disconnect() error {
	bot.Reader = nil
	bot.Conn.Close()
	return nil
}

var reconnecting bool = false

func (bot *Bot) reconnect() {
	if reconnecting {
		// already reconnecting, just wait
		bot.reconnectMutex.Lock()
		bot.reconnectMutex.Unlock()
		return
	}

	bot.reconnectMutex.Lock()
	reconnecting = true
	log.Println("reconnecting...")
	bot.UnloadAllPlugins()
	bot.disconnect()
	delay := 2 * time.Second
	for {
		err := bot.connect()
		if err == nil {
			break
		}
		log.Println(err)
		log.Printf("Reconnecting in %s seconds...\n", delay)
		time.Sleep(delay)
		delay = delay * 2
	}
	bot.LoadAllPlugins()
	reconnecting = false
	bot.reconnectMutex.Unlock()
}

func main() {
	server := flag.String("server", "chat.freenode.net:6667", "IRC server address to connect to")
	ssl := flag.Bool("ssl", false, "use SSL")
	flag.Parse()

	bot := &Bot{
		Server:         *server,
		ssl:            *ssl,
		mutex:          &sync.Mutex{},
		reconnectMutex: &sync.Mutex{},
		pluginsMutex:   &sync.Mutex{},
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
				fmt.Println("Signal HUP received")
				bot.reloadAllPlugins()
			case syscall.SIGUSR1:
				fmt.Println("Signal USR1 received")
				bot.reconnect()
			default:
				fmt.Println("unhandled signal")
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
			bot.reconnect()
			fmt.Println(err)
			time.Sleep(2 * time.Second)
		} else {
			dispatcher(bot, line)
		}
	}
}
