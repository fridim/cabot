package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	s := bufio.NewScanner(os.Stdin) 
	for s.Scan() {
		if strings.Contains(s.Text(), "NOTICE * :*** Looking up your hostname") == true {
			fmt.Fprintf(os.Stdout, "USER grobot 0 * :grobot\r\n")
			fmt.Fprintf(os.Stdout, "NICK grobot\r\n")
			fmt.Fprintf(os.Stdout, "JOIN #esil\r\n")
		}
	}
}
