package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"sort"
)

// Config
var confFile string = "feeds.json"
var maxFeeds = 50

type Feed struct {
	Link string
	LastRead time.Time
}
type Feeds map[string]*Feed

var channels = map[string]Feeds {}
var confFileMutex = &sync.Mutex{}

var fp = gofeed.NewParser()

func loadConf() map[string]Feeds {
	dat, err := ioutil.ReadFile(confFile)
	if err != nil {
		return map[string]Feeds{}
	}

	channels := map[string]Feeds{}

	err = json.Unmarshal(dat, &channels)
	if err != nil {
		panic(err)
	}

	return channels
}

func saveConf() {
	confFileMutex.Lock()
	defer confFileMutex.Unlock()
	bytes, err := json.Marshal(channels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	ioutil.WriteFile(confFile, bytes , 0644)
}

func check_feed(feed string) error {
	_, err := fp.ParseURL(feed)
	return err
}

func get_update(feed *Feed) (error,  string) {
	feedObject, err := fp.ParseURL(feed.Link)
	if err != nil {
		return err, ""
	}
	sort.Sort(sort.Reverse(feedObject))
	for _, item := range feedObject.Items {
		if item.PublishedParsed.After(feed.LastRead) {
			return nil, fmt.Sprintf("%v - %v %v", feedObject.Title, item.Title, item.Link)
			break
		}
	}

	return nil, ""
}

func exists(path string) bool {
    _, err := os.Stat(path)
    return !errors.Is(err, os.ErrNotExist)
}

var r1 = regexp.MustCompile(":([^!]+)[^ ]+ PRIVMSG (#[^ ]+) ::feed ([^ ]+) (https?://[^ ]+)")
var r2 = regexp.MustCompile(":([^!]+)[^ ]+ PRIVMSG (#[^ ]+) ::feeds? list")

func main() {

	if exists(confFile) {
		channels = loadConf()
	} else {
		channels = map[string]Feeds{}
		saveConf()
	}



	go func() {
		for {
			for channel, _ := range channels {
				for feed, _ := range channels[channel] {
					err, r := get_update(channels[channel][feed])
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
					} else {
						if r != "" {
							channels[channel][feed].LastRead = time.Now()
							saveConf()
							fmt.Printf("PRIVMSG %s :%s\n", channel, r)
						}
						// print item
					}
				}
			}
			time.Sleep(10*time.Minute)
		}
	}()

	bio := bufio.NewReader(os.Stdin)

	for {
		line, err := bio.ReadString('\n')
		if err == nil {
			match := r1.FindStringSubmatch(line)
			if len(match) == 5 {
				channel, command, feed := match[2], match[3], match[4]
				feed = strings.Trim(feed, "\r\n")
				switch command {
				case "add", "watch":
					if _, exists := channels[channel]; !exists {
						// Init channel
						channels[channel] = map[string]*Feed{}
					}

					if len(channels[channel]) > maxFeeds {
						fmt.Printf("PRIVMSG %s :Too many feeds to watch, sorry\n", channel)
					}

					if _, exists := channels[channel][feed]; exists {
						fmt.Printf("PRIVMSG %s :Already exists\n", channel)
					} else {

						if err := check_feed(feed); err != nil {
							fmt.Fprintf(os.Stderr, "%s\n", err)
							mess := "go check the logs"
							if len(err.Error()) < 70 {
								mess = err.Error()
							}

							fmt.Printf("PRIVMSG %s :Error: %s\n", channel, mess)
							continue
						}

						channels[channel][feed] = &Feed{
							LastRead: time.Now(),
							Link: feed,
						}
						fmt.Printf("PRIVMSG %s :OK, watching %s\n", channel, feed)
						saveConf()
						continue
					}
				case "del", "delete", "remove", "rm", "forget":
					if _, exists := channels[channel]; !exists {
						continue
					}
					if _, exists := channels[channel][feed]; exists {
						delete(channels[channel], feed)
						saveConf()
						fmt.Printf("PRIVMSG %s :%s removed\n", channel, feed)
					}
				default:
					fmt.Printf("PRIVMSG %s :wrong command '%s'\n", channel, command)
					continue
				}

				continue
			}

			match2 := r2.FindStringSubmatch(line)
			if len(match2) == 3 {
				channel := match2[2]
				if _, exists := channels[channel]; !exists {
					continue
				}
				found := []string{}
				for feed := range channels[channel] {
					found = append(found, feed)
				}
				if len(found) > 0 {
					fmt.Printf("PRIVMSG %s :%s\n", channel, found)
				}
			}

		}
	}

}
