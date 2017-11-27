package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"regexp"
)

func main() {

	destFile := "./urls.txt"
	blacklistFile := "./blacklist.txt"

	s := bufio.NewScanner(os.Stdin) 
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "PRIVMSG") == true {
			skip := false;
			
			re := regexp.MustCompile("https?\\:\\/\\/[[:^space:]]*")
			match := re.FindString(line)
			
			bl, err := os.Open(blacklistFile)
			if err != nil {
			    fmt.Fprintln(os.Stderr, "error", err)
			}
			scanner := bufio.NewScanner(bl)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), match) {
					skip = true;
					break;
				}
			}
			bl.Close()

			fr, err := os.Open(destFile)
			if err != nil {
			    fmt.Fprintln(os.Stderr, "error", err)
			}
			scanner = bufio.NewScanner(fr)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), match) {
					skip = true;
					break;
				}
			}
			fr.Close()
			
			if skip == false {
				
				f, err := os.OpenFile(destFile, os.O_APPEND|os.O_WRONLY, 0600)
			
				if err != nil {
				    fmt.Fprintln(os.Stderr, "error", err)
				}
			
				if _, err = f.WriteString(match + "\n"); err != nil {
				    fmt.Fprintln(os.Stderr, "error", err)
				}
				f.Close()
			
			}
		}
	}
}
