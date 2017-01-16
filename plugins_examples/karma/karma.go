package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// allow every nickname for now
func filter(nickname, channel string) bool {
	return true
}

// from RFC 2812 :
// nickname   =  ( letter / special ) *8( letter / digit / special / "-" )
// letter     =  %x41-5A / %x61-7A       ; A-Z / a-z
//  digit      =  %x30-39                 ; 0-9
//  special    =  %x5B-60 / %x7B-7D
//                   ; "[", "]", "\", "`", "_", "^", "{", "|", "}"
func letterOrSpecial(c byte) bool {
	switch {
	case c >= 0x41 && c <= 0x5A:
		return true
	case c >= 0x61 && c <= 0x7A:
		return true
	case c >= 0x5B && c <= 0x60:
		return true
	case c >= 0x7B && c <= 0x7D:
		return true
	default:
		return false
	}
}
func nicknameChar(c byte) bool {
	switch {
	case c >= 0x30 && c <= 0x39:
		return true
	case c == '-':
		return true
	default:
		return letterOrSpecial(c)
	}
}

func trimNickname(word string) string {
	if len(word) == 0 {
		return word
	}

	if !letterOrSpecial(word[0]) {
		if len(word) > 1 {
			return trimNickname(word[1:])
		} else {
			return ""
		}
	}

	if !nicknameChar(word[len(word)-1]) {
		if len(word) > 1 {
			return trimNickname(word[:len(word)-1])
		} else {
			return ""
		}
	}

	return word
}

func validNickname(nickname string) bool {
	if len(nickname) < 2 {
		return false
	}
	if !letterOrSpecial(nickname[0]) {
		return false
	}
	for i := 0; i < len(nickname); i++ {
		if !nicknameChar(nickname[i]) {
			return false
		}
	}
	return true
}

func parseKarma(line string) ([]string, string, bool) {
	var (
		channel   string
		command   string
		who       string
		nicknames []string
		found     bool
	)

	fmt.Sscanf(line, "%s %s %s", &who, &command, &channel)
	if command != "PRIVMSG" {
		return []string{}, "", false
	}

	// position of text in PRIVMSG lines, line[1:] for ignoring first char usually ':'
	index := strings.Index(line[1:], ":")
	if index == -1 {
		return []string{}, "", false
	}

	// prevent out of slice
	if len(line) < index+3 {
		return []string{}, "", false
	}

	text := line[index+2:]

	if !strings.Contains(text, "++") {
		return []string{}, "", false
	}

	words := strings.Split(text, " ")
	for _, word := range words {
		word := strings.Trim(word, "\n \r")

		if !strings.Contains(word, "++") {
			continue
		}

		if len(word) > 2 {
			l := len(word)
			// word must end with "++" + optional char (like ')')
			if word[l-2] == '+' && (word[l-1] == '+' || word[l-3] == '+') {
				trimmed := trimNickname(word)
				if validNickname(trimmed) {
					nicknames = append(nicknames, trimmed)
				}
			}
		}
	}

	if len(nicknames) > 0 {
		found = true
	}

	return nicknames, channel, found
}

// check if user asked for karma of someone (:karma nickname)
// if yes, then reply to the channel with current karma for that nickname
func replyKarma(line string) {
	var (
		channel string
		command string
		who     string
	)

	fmt.Sscanf(line, "%s %s %s :", &who, &command, &channel)
	if command != "PRIVMSG" {
		return
	}

	// position of text in PRIVMSG lines, line[1:] for ignoring first char usually ':'
	index := strings.Index(line[1:], ":")
	if index == -1 {
		return
	}

	// prevent out of slice
	if len(line) < index+3 {
		return
	}
	text := line[index+2:]

	if len(text) > 6 && text[:6] == ":karma" {
		nickname := strings.Trim(text[7:len(text)-1], "\n \r")
		fmt.Printf("PRIVMSG %s :%s has a karma of %d\n", channel, nickname, db[nickname])
	}

	return
}

var db map[string]uint

func inc(nickname string) {
	db[nickname] += 1
}

func dump() {
	f, _ := os.Create("db.json")
	defer f.Close()
	data, _ := json.Marshal(db)
	f.Write(data)
	f.Sync()
}

func load() {
	_, err := os.Stat("db.json")
	if err != nil {
		db = make(map[string]uint)
		return
	}

	data, _ := ioutil.ReadFile("db.json")
	err = json.Unmarshal(data, &db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error", err)
	}
}

func main() {
	bio := bufio.NewReader(os.Stdin)

	load()

	for {
		line, err := bio.ReadString('\n')
		if err == nil {
			replyKarma(line)

			nicknames, channel, found := parseKarma(line)

			if found {
				for _, nickname := range nicknames {
					if filter(nickname, channel) {
						inc(nickname)
						dump()
					}
				}
			}
		}
	}
}
