package main

import (
	"testing"
)

func HelperBenchmarkFetchKarma(b *testing.B, f func(string) ([]string, string, bool)) {
	var tests = []string{
		":foo!~user@addr PRIVMSG #channel :fridim++",
		":foo!~user@addr PRIVMSG #channel :thanks fridim",
		":foo!~user@addr NOTICE #channel :thanks fridim++\n",
		":foo!~user@addr PRIVMSG #channel :thanks fridim++ oon++",
		":foo!~user@addr PRIVMSG #channel :o/\n",
		"PING :bla\n",
	}

	for i := 0; i < b.N; i++ {
		for _, line := range tests {
			nicknames, channel, found := f(line)
			b.Log(nicknames, channel, found)
		}
	}
}

func BenchmarkParseKarmaRegex(b *testing.B) {
	HelperBenchmarkFetchKarma(b, parseKarmaRegex)
}

func BenchmarkParseKarma(b *testing.B) {
	HelperBenchmarkFetchKarma(b, parseKarma)
}

func TestParseKarma(t *testing.T) {
	HelperTestParseKarma(t, parseKarma)
}
func TestParseKarmaRegex(t *testing.T) {
	HelperTestParseKarma(t, parseKarmaRegex)
}

func HelperTestParseKarma(t *testing.T, f func(string) ([]string, string, bool)) {
	var line string
	var nicknames []string
	var channel string
	var found bool

	line = ":foo!~user@addr PRIVMSG #channel :fridim++"
	nicknames, channel, found = f(line)
	if !found || channel != "#channel" || nicknames[0] != "fridim" {
		t.Error(line)
		t.Log(nicknames, channel, found)
	}

	line = ":foo!~user@addr PRIVMSG #channel :thanks fridim++ !"
	nicknames, channel, found = f(line)
	if !found || channel != "#channel" || nicknames[0] != "fridim" {
		t.Error(line)
		t.Log(nicknames, channel, found)
	}
	line = ":foo!~user@addr PRIVMSG #channel :bla (fridim++)"
	nicknames, channel, found = f(line)
	if !found || channel != "#channel" || nicknames[0] != "fridim" {
		t.Error(line)
		t.Log(nicknames, channel, found)
	}

	line = ":fridim!~fridim@addr PRIVMSG #channel :thanks sim++ oon++"
	nicknames, channel, found = f(line)
	if !found || channel != "#channel" || nicknames[0] != "sim" || nicknames[1] != "oon" {
		t.Error(line)
		t.Log(nicknames, channel, found)
	}

	line = ":fridim!~fridim@addr PRIVMSG #channel :thanks oon"
	nicknames, channel, found = f(line)
	if found {
		t.Error(line)
		t.Log(nicknames, channel, found)
	}

	line = ":fridim!~fridim@addr NOTICE #channel :thanks oon++"
	nicknames, channel, found = f(line)
	if found {
		t.Error(line)
		t.Log(nicknames, channel, found)
	}
}

func TestTrimNickname(t *testing.T) {
	var (
		trim     string
		expected string
		input    string
	)

	input = "(foo"
	trim = trimNickname(input)
	expected = "foo"
	if trim != expected {
		t.Error(input, trim, expected)
	}

	input = "::::(8foo"
	trim = trimNickname(input)
	expected = "foo"
	if trim != expected {
		t.Error(input, trim, expected)
	}

	input = "-`foo{}"
	trim = trimNickname(input)
	expected = "`foo{}"
	if trim != expected {
		t.Error(input, trim, expected)
	}

	input = "(foo++)"
	trim = trimNickname(input)
	expected = "foo"
	if trim != expected {
		t.Error(input, trim, expected)
	}
}

func TestInc(t *testing.T) {
	db = make(map[string]uint)
	inc("foo")
	inc("foo")
	inc("foo")
	inc("foo")
	inc("bar")

	if db["foo"] != 4 {
		t.Error(db)
	}
	dump()
}

func TestLoad(t *testing.T) {
	db = nil
	load()

	if db["foo"] != 4 {
		t.Error(db)
	}
}
