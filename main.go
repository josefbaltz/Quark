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
	flag.StringVar(&projectID, "g", "", "Google Cloud Project ID")
	flag.Parse()
}

var projectID string
var token string
var buffer = make([][]byte, 0)

type Task struct {
	Credits int
	Level   int
	Attack  int
	Defense int
}

func main() {

	if token == "" || projectID == "" {
		fmt.Println("--Error--")
		fmt.Println("Your start command should look like:")
		fmt.Println("Quark -t <Discord API Token> -g <GCP ProjectID>")
		os.Exit(0)
	}

	//Build Bot
	quark, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := context.Background()
	gcp, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println("Failed to create GCP client")
		fmt.Println(err)
		os.Exit(1)
	}

	//Register Callback Events
	quark.AddHandler(botConnected)
	quark.AddHandler(messageRecieved)

	//Open a Connection to Discord
	err = quark.Open()
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Println("Quark is running Subatomic!")
	fmt.Println("Running until stop command ...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	quark.Close()
}

func populateUserGCP(ID string, session *discordgo.Session, event *discordgo.MessageCreate) {
	ctx := context.Background()
	gcp, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println("Failed to create GCP client")
		fmt.Println(err)
	}

	taskKey := datastore.NameKey("User", ID, nil)

	task := Task{
		Attack:  0,
		Defense: 0,
		Credits: 10,
		Level:   0,
	}

	if _, err := gcp.Put(ctx, taskKey, &task); err != nil {
		fmt.Println("--Warning--")
		fmt.Println("Couldn't Add user to GCP Datatstore")
		fmt.Println(err)

	}
}

func botConnected(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateStatus(0, "at the subatomic level")
}

func messageRecieved(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.Bot {
		return
	}

	if strings.HasPrefix(strings.ToLower(event.Content), "q.ping") {
		session.ChannelMessageSend(event.ChannelID, "Ping!")
	}
}
