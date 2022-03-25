package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	piston "github.com/milindmadhukar/go-piston"
)

// bot parameters
var (
	AppID    string = "955836104559460362"
	botToken string
	client   *piston.Client
)

var s *discordgo.Session

// code execution ouptput
var o chan string

const of string = "Output:\n```\n%s\n```\n"

// regex for parsing message to execute code
const r string = "run```.*"

// used to trim to obtain language form message
const t string = "run```"

func init() {
	botToken = os.Getenv("BOT_TOKEN")
	flag.Parse()
	client = piston.CreateDefaultClient()
	o = make(chan string)
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
	// check to see if we are executing go code
	// this is based on a writing standard in discord for writing code in a paragraph message block
	// example message: ```go ... ```
	regx, _ := regexp.Compile(r)
	c := strings.Split(m.Content, "\n")
	for _, b := range c {
		if regx.MatchString(b) {
			// execute code
			lang := getLanguage(b)
			go exec(m.ChannelID, m.Content, m.Reference(), lang)

			// send initial reply message containing output of code execution
			// "Run" button is injected in the message so the user may re run their code
			_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content:   fmt.Sprintf(of, <-o),
				Reference: m.Reference(),
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
			return
		}
	}
}

// handler for re-executing go code when the "Run" button is clicked
func reExecuctionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// check if go button was clicked
	if i.MessageComponentData().CustomID == "run" {
		// get referenced channel message
		// used to fetch the code from the message that contains it
		msg, err := s.ChannelMessage(i.ChannelID, i.Message.MessageReference.MessageID)
		if err != nil {
			log.Fatalf("Could not get message reference: %v", err)
		}

		// execute code
		lang := getLanguage(msg.Content)
		go exec(i.ChannelID, msg.Content, i.Message.Reference(), lang)

		// send interaction respond
		// update message reply with new code execution output
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf(of, <-o),
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
			},
		})
	}

}

func exec(channelID string, messageContent string, messageReference *discordgo.MessageReference, lang string) {
	// remove strings based on regex for proper code execution
	rs := []string{".*```.*", "\n\n"}
	for _, r := range rs {
		regex := regexp.MustCompile(r)
		messageContent = regex.ReplaceAllString(messageContent, "")
	}

	// execute code using piston library
	output, err := client.Execute(lang, "",
		[]piston.Code{
			{
				Name:    fmt.Sprintf("%s-code", messageReference.MessageID),
				Content: messageContent,
			},
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		_, _ = s.ChannelMessageSendReply(channelID, err.Error(), messageReference)
	}

	o <- output.GetOutput()
}

// get coding language in message block
func getLanguage(content string) string {
	r, _ := regexp.Compile("run```.*")
	return strings.TrimPrefix(string(r.Find([]byte(content))), t)
}
