package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	AppID    string = "955836104559460362"
	botToken string
)

var s *discordgo.Session

// set variables and flags
func init() {
	botToken = os.Getenv("BOT_TOKEN")
	flag.Parse()
}

// create discord session
func init() {
	var err error
	s, err = discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	// add function handlers for code execution
	s.AddHandler(executionHandler)
	s.AddHandler(reExec)

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func executionHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	c := strings.Split(m.Content, "\n")
	// check to see if we are executing go code
	// this is based on a writing standard in discord for writing code in a paragraph message block
	// example message: ```go ... ```
	for _, b := range c {
		if matched, err := regexp.MatchString("run```go", b); err != nil {
			fmt.Println("did not find any string matching")
		} else if matched {
			codeOuput := goExec(m.ChannelID, m.Content, m.Reference())
			sendMessageComplex(m.ChannelID, string(codeOuput), m.Reference())
			return
		}
	}
}

// handler for re-executing go code when the "Run" button is clicked
func reExec(s *discordgo.Session, m *discordgo.InteractionCreate) {
	// check if go button was clicked
	if m.MessageComponentData().CustomID == "go_run" {
		msg, err := s.ChannelMessage(m.ChannelID, m.Message.MessageReference.MessageID)
		if err != nil {
			log.Fatalf("Could not get message reference: %v", err)
		}

		messageReference := m.Message.Reference()
		codeOutput := goExec(m.ChannelID, msg.Content, messageReference)
		editComplexMessage(m.Message.ID, m.ChannelID, string(codeOutput), messageReference)
	}
}

func goExec(channelID string, messageContent string, messageReference *discordgo.MessageReference) []byte {
	// add regex string replacements for content
	var r []regexp.Regexp
	blockre := regexp.MustCompile(".*```.*")
	whitespacere := regexp.MustCompile("\n\n")
	r = append(r, *blockre, *whitespacere)

	// remove strings based on regex for proper code execution
	content := messageContent
	for _, regex := range r {
		content = regex.ReplaceAllString(content, "")
	}

	// create go execution file
	if err := os.MkdirAll("code", os.ModePerm); err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("code/code.go", []byte(content), 0644)

	// run command
	cmd := exec.Command("go", "run", "code/code.go")
	o, err := cmd.Output()

	// output error in discord if code did not successfully execute
	if err != nil {
		fmt.Println(err.Error())
		_, _ = s.ChannelMessageSendReply(channelID, err.Error(), messageReference)
		return nil
	}

	return o
}

// send initial reply message containing output of code execution
// "Run" button is injected in the message so the user may re run their code
func sendMessageComplex(channelID string, codeOutput string, messageReference *discordgo.MessageReference) {
	_, _ = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:   fmt.Sprintf("Output:\n```\n%s\n```\n", string(codeOutput)),
		Reference: messageReference,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Run",
						Style:    discordgo.SuccessButton,
						CustomID: "go_run",
					},
				},
			},
		},
	})
}

func editComplexMessage(messageID string, channelID string, codeOutput string, messageReference *discordgo.MessageReference) {
	content := fmt.Sprintf("Output:\n```\n%s\n```\n", string(codeOutput))
	s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      messageID,
		Channel: channelID,
		Content: &content,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Run",
						Style:    discordgo.SuccessButton,
						CustomID: "go_run",
					},
				},
			},
		},
	})
}
