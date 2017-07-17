package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"
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
		p.bot.toStderr <- fmt.Sprintf("%s: %s\n", p, err)
		p.unload()
	}
}

func (p *Plugin) kill() error {
	if p.cmd == nil {
		log.Fatalf("%v: plugin.cmd is nil", p)
	}
	if p.cmd.Process == nil {
		log.Fatalf("%v: plugin.cmd.Process is nil", p)
	}

	if err := p.cmd.Process.Signal(syscall.Signal(0)); err == nil {
		if err := p.cmd.Process.Signal(syscall.Signal(15)); err != nil {
			if err := p.cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			log.Println("Killed (9): ", p)
		} else {
			log.Println("Killed (15): ", p)
		}

		p.cmd.Wait()
		return nil
	} else {
		log.Printf("command '%v' process.Signal on pid %d returned: %v\n", p.cmd, p.cmd.Process.Pid, err)
		return err
	}
}

func (p *Plugin) unload() {
	for i, plugin := range p.bot.plugins {
		if p == plugin {
			p.bot.plugins[i] = nil
		}
	}
}

func (p *Plugin) publish(in io.Reader, out chan string) {
	reader := bufio.NewReader(in)
	defer wg.Done()

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			logErr.Printf("[%s] %s", p.cmd.Path, err)
			return
		}

		out <- line
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
	go p.publish(p.stdout, bot.toConn)
	go p.publish(p.stderr, bot.toStderr)
	p.cmd = cmd
	p.bot = bot
	wg.Add(1)
	go p.start()
	return p
}

func (bot *Bot) LoadAllPlugins() {
	bot.pluginsMutex.Lock()
	bot.plugins = []*Plugin{}

	pluginDir, _ := os.Open("plugins")
	files, _ := pluginDir.Readdir(-1)
	for _, f := range files {
		if f.Mode().IsRegular() && f.Mode().Perm()&0111 != 0 {
			bot.plugins = append(bot.plugins, bot.newPlugin(path.Join("plugins", f.Name())))
		}
	}
	bot.pluginsMutex.Unlock()
	log.Printf("All plugins loaded, wg: %v\n", wg)
}

func (bot *Bot) UnloadAllPlugins() {
	bot.pluginsMutex.Lock()

	for _, plugin := range bot.plugins {
		if plugin == nil {
			continue
		}
		plugin.kill()
		plugin.unload()
	}
	bot.pluginsMutex.Unlock()
	log.Printf("All plugins unloaded, wg: %v\n", wg)
}
func (bot *Bot) reloadAllPlugins() {
	bot.UnloadAllPlugins()
	bot.LoadAllPlugins()
}
