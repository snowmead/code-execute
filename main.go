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
	GuildID  = flag.String("guild", "953488559724183602", "Test guild ID")
	BotToken = flag.String("token", "OTU1ODM2MTA0NTU5NDYwMzYy.YjndvQ.Ywgrne6NkSVUXvX23y6giAHLt2c", "Bot access token")
	AppID    = flag.String("app", "955836104559460362", "Application ID")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	var err error
	s.AddHandler(goExecutionHandler)

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func goExecutionHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// handle running code sent from user
	if c := strings.Split(m.Content, "\n"); c[0] == "run" {

		// check to see if we are executing go code
		// this is based on a writing standard in discord for writing code in a message
		// example message: ```go ... ```
		if strings.Contains(c[1], "go") {
			// add regex string replacements for content
			var r []regexp.Regexp
			runre := regexp.MustCompile("run")
			blockre := regexp.MustCompile("```.*")
			whitespacere := regexp.MustCompile("\n\n")
			r = append(r, *runre, *blockre, *whitespacere)

			// remove strings based on regex for proper code execution
			content := m.Content
			for _, regex := range r {
				content = regex.ReplaceAllString(content, "")
			}

			// create go execution file
			ioutil.WriteFile("code/code.go", []byte(content), 0644)

			// run command
			cmd := exec.Command("go", "run", "code/code.go")
			o, err := cmd.Output()

			// output error in discord if code did not successfully execute
			if err != nil {
				fmt.Println(err.Error())
				_, _ = s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
				return
			}

			fmt.Println(string(o))
			// send execute output
			_, _ = s.ChannelMessageSendReply(m.ChannelID, string(o), m.Reference())
		}
	}
}
