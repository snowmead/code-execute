package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v43/github"
	piston "github.com/milindmadhukar/go-piston"
)

var s *discordgo.Session
var gclient *github.Client
var ctx context.Context

// bot parameters
var (
	botToken string
	pclient  *piston.Client
)

// code execution ouptput
var o chan string

func init() {
	botToken = os.Getenv("BOT_TOKEN")
	flag.Parse()
	pclient = piston.CreateDefaultClient()
	ctx = context.Background()
	gclient = github.NewClient(nil)
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
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) { log.Println("Bot is up!") })
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
	// avoid handling the message that the bot creates when replying to a user
	if m.Author.Bot {
		return
	}

	// extract code block from message and execute code
	var responseContent string
	if lang, codeBlock := codeBlockExtractor(m.Content); lang != "" || codeBlock != "" {
		go exec(m.ChannelID, codeBlock, m.Reference(), lang)
		responseContent = <-o
	} else {
		return
	}

	// send initial reply message containing output of code execution
	// "Run" button is injected in the message so the user may re run their code
	_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:   responseContent,
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
}

// handler for re-executing go code when the "Run" button is clicked
func reExecuctionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// check if go button was clicked
	if i.MessageComponentData().CustomID == "run" {
		// get referenced channel message
		// used to fetch the code from the message that contains it
		m, err := s.ChannelMessage(i.ChannelID, i.Message.MessageReference.MessageID)
		if err != nil {
			log.Printf("Could not get message reference: %v", err)
		}

		// extract code block from message and execute code
		var responseContent string
		if lang, codeBlock := codeBlockExtractor(m.Content); lang != "" || codeBlock != "" {
			go exec(i.ChannelID, codeBlock, i.Message.Reference(), lang)
			responseContent = <-o
		} else {
			responseContent = fmt.Sprintln("Could not find any code in message to execute")
		}

		// send interaction respond
		// update message reply with new code execution output
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: responseContent,
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

		if err != nil {
			log.Printf("Could not send respond interaction: %v", err)
		}
	}
}

// handle code execution
// sends output to chan
func exec(channelID string, code string, messageReference *discordgo.MessageReference, lang string) {
	// execute code using piston library
	output, err := pclient.Execute(lang, "",
		[]piston.Code{
			{
				Name:    fmt.Sprintf("%s-code", messageReference.MessageID),
				Content: code,
			},
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		_, _ = s.ChannelMessageSendReply(channelID, err.Error(), messageReference)
	}

	o <- fmt.Sprintf(">>> Output [%s - %s]\n```\n%s\n```\n", output.Language, output.Version, output.GetOutput())
}

func codeBlockExtractor(content string) (string, string) {
	// check to see if we are executing a code block
	// this is based on a writing standard in discord for writing code in a paragraph message block
	// example message: ```go ... ```
	rcb, _ := regexp.Compile("run```.*")
	rg, _ := regexp.Compile("run https://gist.github.com/.*/.*")
	rgist, _ := regexp.Compile("run https://gist.github.com/.*/")
	c := strings.Split(content, "\n")
	for bi, bb := range c {
		// extract gist language and code to execute
		if rg.MatchString(bb) {
			gistID := rgist.ReplaceAllString(bb, "")
			gist, _, _ := gclient.Gists.Get(ctx, gistID)
			return strings.ToLower(*gist.Files["helloworld.go"].Language), *gist.Files["helloworld.go"].Content
		}

		// extract code block to execute
		if rcb.MatchString(bb) {
			lang := strings.TrimPrefix(string(rcb.Find([]byte(content))), "run```")
			// find end of code block
			var codeBlock string
			endBlockRegx, _ := regexp.Compile("```")
			subArray := c[bi+1:]

			for ei, eb := range subArray {
				if endBlockRegx.Match([]byte(eb)) {
					// create code block to execute
					codeBlock = strings.Join(subArray[:ei], "\n")
					return lang, codeBlock
				}
			}
		}
	}

	return "", ""
}
