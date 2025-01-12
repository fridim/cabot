package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/eidolon/wordwrap"
	"github.com/fridim/cabot/pkg/irc"
	gogpt "github.com/sashabaranov/go-openai"
	"os"
	"regexp"
	"strings"
)

var r1 = regexp.MustCompile(":([^!]+)![^ ]+ PRIVMSG (#[^ ]+) :(.*)")
var r2 = regexp.MustCompile("^[cC]abot[,:]? (.*)")
var wrapper = wordwrap.Wrapper(400, false)

func parsePrompt(line string) (gogpt.ChatCompletionMessage, string, bool) {
	m := r1.FindStringSubmatch(line)
	if m == nil {
		return gogpt.ChatCompletionMessage{}, "", false
	}

	nick := m[1]
	channel := m[2]
	text := m[3]
	prompt := ""

	if m2 := r2.FindStringSubmatch(text); m2 != nil {
		prompt = m2[1]

		return gogpt.ChatCompletionMessage{
				Role:    "user",
				Content: fmt.Sprintf("C'est %s qui parle. %s", nick, prompt),
			},
			channel,
			true
	}

	return gogpt.ChatCompletionMessage{}, "", false
}

var messages = make(map[string][]gogpt.ChatCompletionMessage)
var messagesInit = []gogpt.ChatCompletionMessage{
	{
		Role:    "system",
		Content: "Tu réponds toujours sous la forme d'une réponse très courte, moins de 80 caractères. Prends le temps de développer point par point ton raisonnement, mais n'inclus dans ta réponse que la conclusion.",
	},
}

var maxMessages = 15

func main() {
	token := os.Getenv("OPENAI_TOKEN")
	if token == "" {
		tokenBytes, err := os.ReadFile("openai_token.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading token.txt: %v", err)
			os.Exit(1)
		}

		token = strings.Trim(string(tokenBytes), " \t\r\n")
	}

	// Detect if prompt is available locally, if then load it as the context

	// Check if file openai_context.txt exists
	if _, err := os.Stat("openai_context.txt"); err == nil {
		// Load the file
		contextBytes, err := os.ReadFile("openai_context.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading openai_context.txt: %v", err)
			os.Exit(1)
		}

		context := strings.Trim(string(contextBytes), " \t\r\n")

		// Add the context to the messagesInit
		messagesInit = append(messagesInit, gogpt.ChatCompletionMessage{
			Role:    "system",
			Content: context,
		})
	}

	c := gogpt.NewClient(token)
	bio := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	for {
		line, err := bio.ReadString('\n')
		if err == nil {
			if message, channel, ok := parsePrompt(line); ok {
				// TODO: ensure channel is autorized

				if _, ok := messages[channel]; !ok {
					messages[channel] = make([]gogpt.ChatCompletionMessage, 0)
				}
				messages[channel] = append(messages[channel], message)
				// Limit the number of messages to 10
				if len(messages[channel]) > maxMessages {
					messages[channel] = messages[channel][1:]
				}

				// Call the GPT-3 API
				req := gogpt.ChatCompletionRequest{
					Model:            gogpt.O1Mini,
					MaxTokens:        200,
					Temperature:      0.8,
					PresencePenalty:  0.8,
					FrequencyPenalty: 0.8,
					Messages:         append(messagesInit, messages[channel]...),
				}
				resp, err := c.CreateChatCompletion(ctx, req)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error calling GPT API: %v", err)
					continue
				}

				if len(resp.Choices) > 0 {
					for _, c := range resp.Choices {
						messages[channel] = append(messages[channel], c.Message)
						irc.Privmsg(channel, c.Message.Content)
					}
				}
			}
		}
	}
}
