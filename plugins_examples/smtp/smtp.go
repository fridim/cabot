package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"regexp"
)

func main() {
	bio := bufio.NewReader(os.Stdin)

	for {
		line, err := bio.ReadString('\n')
		if err == nil {
			found, from, channel, nickname, content := parseMail(line)
			if found {
				sendMail(from, channel, nickname, content)
				fmt.Printf("PRIVMSG %s :mail sent\n", channel)
			}
		}
	}
}

type Conf struct {
	Address  string
	Host     string
	Hello    string
	User     string
	Password string
	CA       string
	From     string
	Insecure bool
	Mails    map[string]string
}

func loadConf() Conf {
	dat, err := ioutil.ReadFile("smtp.json")
	if err != nil {
		return Conf{}
	}

	conf := Conf{}

	err = json.Unmarshal(dat, &conf)
	if err != nil {
		panic(err)
	}

	return conf
}

var r1 = regexp.MustCompile(":([^!]+)[^ ]+ PRIVMSG (#[^ ]+) ::e?mail ([^ ]+) (.*)")

func parseMail(line string) (bool, string, string, string, string) {
	match := r1.FindStringSubmatch(line)

	if len(match) != 5 {
		return false, "", "", "", ""
	}

	return true, match[1], match[2], match[3], match[4]
}

func plog(err error, channel string, format string, a ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

	if len(format) == 0 {
		return
	}
	s := fmt.Sprintf(format, a...)
	fmt.Printf("PRIVMSG %s :%s\n", channel, s)
}

func sendMail(from, channel, nickname, content string) {
	conf := loadConf()

	mail := conf.Mails[nickname]
	if mail == "" {
		plog(nil, channel, "no mail set for %s", nickname)
		return
	}

	// Connect to the remote SMTP server.
	c, err := smtp.Dial(conf.Address)
	if err != nil {
		plog(err, channel, "cannot dial to %s", conf.Address)
		return
	}

	if err := c.Hello(conf.Hello); err != nil {
		plog(err, channel, "error with smtp Hello")
		return
	}

	if conf.Insecure {
		if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
			plog(err, channel, "STARTTLS with %s failed", conf.Address)
			return
		}
	} else {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(conf.CA))
		if !ok {
			panic("failed to parse root certificate")
		}

		if err := c.StartTLS(&tls.Config{RootCAs: roots}); err != nil {
			plog(err, channel, "STARTTLS with %s failed", conf.Address)
			return
		}
	}

	if err := c.Auth(smtp.PlainAuth("", conf.User, conf.Password, conf.Host)); err != nil {
		plog(err, channel, "Auth failed")
		return
	}

	// Set the sender and recipient first
	if err := c.Mail(conf.From); err != nil {
		plog(err, channel, "Set sender failed")
		return
	}
	if err := c.Rcpt(mail); err != nil {
		plog(err, channel, "Set Rcpt failed")
		return
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		plog(err, channel, "Set email body failed")
		return
	}

	fmt.Fprintf(wc, "From: %s\n", conf.From)
	fmt.Fprintf(wc, "To: %s\n", mail)
	fmt.Fprintf(wc, "Subject: t'as un message de %s sur %s\n", from, channel)
	fmt.Fprintf(wc, "Content-Type: text/plain; charset=utf-8\n\n")
	fmt.Fprintf(wc, "<%s> %s", from, content)
	err = wc.Close()
	if err != nil {
		plog(err, channel, "Set email body failed")
		return
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		plog(err, channel, "Quit() failed")
		return
	}
}
