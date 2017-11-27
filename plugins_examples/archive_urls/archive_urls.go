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

	blacklistMatch := [2]string{
		"",
		"./urls.txt",
	}

	s := bufio.NewScanner(os.Stdin) 
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "PRIVMSG") == true {
			skip := false;
			
			re := regexp.MustCompile("https?\\:\\/\\/[[:^space:]]*")
			match := re.FindString(line)
			
			for i := 1; i < len(blacklistMatch); i++ {
				if match == blacklistMatch[i] {
					skip = true
				}
			}
			
			fr, err := os.Open(destFile)
			if err != nil {
			    fmt.Fprintln(os.Stderr, "error", err)
			}
			scanner := bufio.NewScanner(fr)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), match) {
					skip = true
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
