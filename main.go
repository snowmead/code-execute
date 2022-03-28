package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

const (
	cblock int = iota
	cgist
	cfile
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
	ctype, lang, codeBlock := codeBlockExtractor(m.Message)
	if lang != "" || codeBlock != "" {
		go exec(m.ChannelID, codeBlock, m.Reference(), lang)
		responseContent = <-o
	} else {
		return
	}

	// only add run button for code block and gist execution
	var runButton []discordgo.MessageComponent
	if ctype != cfile {
		runButton = []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Run",
					Style:    discordgo.SuccessButton,
					CustomID: "run",
				},
			},
		},
		}
	}

	// send initial reply message containing output of code execution
	// "Run" button is injected in the message so the user may re run their code
	_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:    responseContent,
		Reference:  m.Reference(),
		Components: runButton,
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
		if _, lang, codeBlock := codeBlockExtractor(m); lang != "" || codeBlock != "" {
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
		o <- fmt.Sprintf(">>> Execution Failed [%s - %s]\n```\n%s\n```\n", output.Language, output.Version, err)
	}

	o <- fmt.Sprintf(">>> Output [%s - %s]\n```\n%s\n```\n", output.Language, output.Version, output.GetOutput())
}

func codeBlockExtractor(m *discordgo.Message) (int, string, string) {
	mc := m.Content
	// syntax for executing a code block
	// this is based on a writing standard in discord for writing code in a paragraph message block
	// example message: ```go ... ```
	rcb, _ := regexp.Compile("run```.*")
	// syntax for executing a gist
	rg, _ := regexp.Compile("run https://gist.github.com/.*/.*")
	rgist, _ := regexp.Compile("run https://gist.github.com/.*/")
	// syntax for executing an attached file
	rf, _ := regexp.Compile("run *.*")

	c := strings.Split(mc, "\n")
	for bi, bb := range c {
		// extract code block to execute
		if rcb.MatchString(bb) {
			lang := strings.TrimPrefix(string(rcb.Find([]byte(mc))), "run```")
			// find end of code block
			var codeBlock string
			endBlockRegx, _ := regexp.Compile("```")

			sa := c[bi+1:]
			for ei, eb := range sa {
				if endBlockRegx.Match([]byte(eb)) {
					// create code block to execute
					codeBlock = strings.Join(sa[:ei], "\n")
					return cblock, lang, codeBlock
				}
			}
		}
		// extract gist language and code to execute
		if rg.MatchString(bb) {
			gistID := rgist.ReplaceAllString(bb, "")
			gist, _, err := gclient.Gists.Get(ctx, gistID)
			if err != nil {
				log.Printf("Failed to obtain gist: %v\n", err)
			}
			return cgist, strings.ToLower(*gist.Files["helloworld.go"].Language), *gist.Files["helloworld.go"].Content
		}
		// extract file language and code to execute
		if rf.MatchString(bb) {
			if len(m.Attachments) > 0 {
				// handle 1 file in message attachments
				f := m.Attachments[0]
				// get language from extension
				lang := strings.TrimLeft(filepath.Ext(f.Filename), ".")
				// get code from file
				resp, err := http.Get(f.URL)
				if err != nil {
					log.Printf("Failed GET http call to file attachment URL: %v\n", err)
				}
				defer resp.Body.Close()
				codeBlock, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Failed to obtain code from response body: %v\n", err)
				}
				return cfile, lang, string(codeBlock)
			}
		}
	}

	return -1, "", ""
}
