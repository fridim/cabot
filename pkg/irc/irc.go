package irc

import (
	"github.com/eidolon/wordwrap"
	"strings"
	"time"
	"fmt"
)

const (
	Delay = 1 * time.Second
)

func Privmsg(target, message string) {
	// Define the maximum length, including the PRIVMSG command and the target
	// RFC tells us that the maximum length is 512
	maxLen := 512 - len(target) - 15
	wrapper := wordwrap.Wrapper(maxLen, false)
	mess := wrapper(message)


	messages := strings.Split(mess, "\n")
	for _, line := range messages {
		fmt.Printf(
			"PRIVMSG %s :%s\r\n",
			target,
			line)
		time.Sleep(Delay)
	}
}
