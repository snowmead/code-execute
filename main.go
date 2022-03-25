package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	piston "github.com/milindmadhukar/go-piston"
)

// Bot parameters
var (
	AppID    string = "955836104559460362"
	botToken string
	client   *piston.Client
)

var s *discordgo.Session

// set variables and flags
func init() {
	botToken = os.Getenv("BOT_TOKEN")
	flag.Parse()
	client = piston.CreateDefaultClient()
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
	s.AddHandler(reExecuctionHandler)

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
	start := time.Now()
	c := strings.Split(m.Content, "\n")
	// check to see if we are executing go code
	// this is based on a writing standard in discord for writing code in a paragraph message block
	// example message: ```go ... ```
	for _, b := range c {
		if matched, err := regexp.MatchString("run```.*", b); err != nil {
			fmt.Println("error matching string")
		} else if matched {
			r, _ := regexp.Compile("run```.*")
			lang := strings.TrimPrefix(string(r.Find([]byte(b))), "run```")
			codeOuput := goExec(m.ChannelID, m.Content, m.Reference(), lang)
			sendMessageComplex(m.ChannelID, codeOuput, m.Reference())
			return
		}
	}
	fmt.Println(time.Since(start))
}

// handler for re-executing go code when the "Run" button is clicked
func reExecuctionHandler(s *discordgo.Session, m *discordgo.InteractionCreate) {
	start := time.Now()

	// check if go button was clicked
	if m.MessageComponentData().CustomID == "run" {
		msg, err := s.ChannelMessage(m.ChannelID, m.Message.MessageReference.MessageID)
		if err != nil {
			log.Fatalf("Could not get message reference: %v", err)
		}
		messageReference := m.Message.Reference()

		r, _ := regexp.Compile("run```.*")
		lang := strings.TrimPrefix(string(r.Find([]byte(msg.Content))), "run```")
		codeOutput := goExec(m.ChannelID, msg.Content, messageReference, lang)
		editComplexMessage(m.Message.ID, m.ChannelID, codeOutput, messageReference)
	}

	fmt.Println(time.Since(start))
}

func goExec(channelID string, messageContent string, messageReference *discordgo.MessageReference, lang string) string {
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

	output, err := client.Execute(lang, "",
		[]piston.Code{
			{
				Name:    fmt.Sprintf("%s-code", messageReference.MessageID),
				Content: content,
			},
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		_, _ = s.ChannelMessageSendReply(channelID, err.Error(), messageReference)
		return ""
	}

	return output.GetOutput()
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
						CustomID: "run",
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
						CustomID: "run",
					},
				},
			},
		},
	})
}
