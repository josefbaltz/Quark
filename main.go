package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cloud.google.com/go/datastore"
	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Discord API Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

//User - Users have credits, attack, and defense integers
type User struct {
	Credits int
	Attack  int
	Defense int
}

func main() {
	if token == "" {
		fmt.Println("--Error--")
		fmt.Println("Your start command should look like:")
		fmt.Println("Quark -t <Discord API Token>")
		os.Exit(0)
	}

	//Build Bot
	quark, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		os.Exit(1)
	}

	//Register Callback Events
	quark.AddHandler(botConnected)
	quark.AddHandler(basicCommands)
	quark.AddHandler(gameCommands)

	//Open a Connection to Discord
	err = quark.Open()
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		os.Exit(2)
	}

	//Close if close command recieved from os
	fmt.Println("Quark is running Subatomic!")
	fmt.Println("Running until stop command ...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	quark.Close()
}

func botConnected(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateStatus(0, "Type q.help")
}

func basicCommands(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.Bot {
		return
	}

	if strings.HasPrefix(strings.ToLower(event.Content), "q.ping") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		session.ChannelMessageSend(event.ChannelID, "Ping!")
	}

	if strings.HasPrefix(strings.ToLower(event.Content), "q.help") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		session.ChannelMessageSend(event.ChannelID, "wip")
	}
}

func gameCommands(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.Bot {
		return
	}

	var failureMessage = "Failed! Message OrangeFlare#1337"

	if strings.HasPrefix(strings.ToLower(event.Content), "q.game.join") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		ctx := context.Background()
		gcp, err := datastore.NewClient(ctx, "quarkbot")
		if err != nil {
			fmt.Println("--Error--")
			fmt.Println("Failed to create GCP client")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, failureMessage)
			return
		}

		userKey := datastore.NameKey("User", event.Author.ID, nil)
		user := User{
			Attack:  0,
			Defense: 0,
			Credits: 100,
		}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			if _, err := gcp.Put(ctx, userKey, &user); err != nil {
				fmt.Println("--Warning--")
				fmt.Println("Failed to add user to GCP Datatstore")
				fmt.Println(err)
				session.ChannelMessageSend(event.ChannelID, failureMessage)
				return
			}
			session.ChannelMessageSend(event.ChannelID, "Registered!")
		} else {
			session.ChannelMessageSend(event.ChannelID, "You are already registered!")
			return
		}
	}
}
