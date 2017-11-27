package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"regexp"
)

func main() {
	s := bufio.NewScanner(os.Stdin) 
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "PRIVMSG") == true {
			skip := false;

	    		re := regexp.MustCompile("https?\\:\\/\\/[[:^space:]]*")
	    		match := re.FindString(line)

			if match != "" && match != "https://arch.cccp.io/urls.txt" {
	    			fr, err := os.Open("/var/www/html/arch/urls.txt")
	    			if err != nil {
				    fmt.Fprintln(os.Stderr, "error", err)
	    			}
				defer fr.Close()
				scanner := bufio.NewScanner(fr)
  				for scanner.Scan() {
					if strings.Contains(scanner.Text(), match) {
						skip = true
						fmt.Fprintf(os.Stdout, "PRIVMSG #esil :url déjà enregistrée\r\n")	
					}
  				}
				
				if skip == false {
	    				f, err := os.OpenFile("/var/www/html/arch/urls.txt", os.O_APPEND|os.O_WRONLY, 0600)

	    				if err != nil {
					    fmt.Fprintln(os.Stderr, "error", err)
	    				}
					defer f.Close()

	    				if _, err = f.WriteString(match + "\n"); err != nil {
					    fmt.Fprintln(os.Stderr, "error", err)
	    				}
				}
			}
		}
	}
}
