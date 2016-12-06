package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

type Plugin struct {
	cmd    *exec.Cmd
	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader
	bot    *Bot
}

func (p *Plugin) String() string {
	return fmt.Sprintf("%s", p.cmd.Path)
}

var wg sync.WaitGroup

func (p *Plugin) start() {
	defer wg.Done()
	err := p.cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", p, err)
		p.unload()
	}
}

func (p *Plugin) kill() {
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
}

func (p *Plugin) unload() {
	for i, plugin := range p.bot.plugins {
		if p == plugin {
			p.bot.plugins[i] = nil
		}
	}
}

func publish(p *Plugin, in io.Reader, out io.Writer) {
	reader := bufio.NewReader(in)
	defer wg.Done()
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			fmt.Println(err)
			return
		}

		// Do not output passwords
		if !strings.Contains(line, "PRIVMSG Nickserv :identify") {
			fmt.Printf("[%s] %s", p.cmd.Path, line)
		}
		p.bot.mutex.Lock()
		fmt.Fprintf(out, line)
		p.bot.mutex.Unlock()
	}
}

func (bot *Bot) newPlugin(path string) *Plugin {
	cmd := exec.Command(path)
	var p *Plugin = &Plugin{}

	var err error
	p.stdin, err = cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	p.stdout, err = cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	p.stderr, err = cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(2)
	go publish(p, p.stdout, bot.Conn)
	go publish(p, p.stderr, os.Stderr)
	p.cmd = cmd
	p.bot = bot
	wg.Add(1)
	go p.start()
	return p
}

func (bot *Bot) LoadAllPlugins() {
	bot.plugins = []*Plugin{}

	pluginDir, _ := os.Open("plugins")
	files, _ := pluginDir.Readdir(-1)
	for _, f := range files {
		if f.Mode().IsRegular() && f.Mode().Perm()&0111 != 0 {
			bot.plugins = append(bot.plugins, bot.newPlugin(path.Join("plugins", f.Name())))
		}
	}
}

func (bot *Bot) UnloadAllPlugins() {
	for _, plugin := range bot.plugins {
		if plugin == nil {
			continue
		}
		plugin.kill()
		plugin.unload()
	}
}
func (bot *Bot) reloadAllPlugins() {
	bot.UnloadAllPlugins()
	bot.LoadAllPlugins()
}
