package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Discord API Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

func main() {

	if token == "" {
		fmt.Println("Error 404 Token not found!")
		fmt.Println("Your start command should look like:")
		fmt.Println("Quark -t <API Token>")
		return
	}

	//Build Bot
	quark, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		return
	}

	//Register Callback Events
	quark.AddHandler(botConnected)

	//Open a Connection to Discord
	err = quark.Open()
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
	}

	fmt.Println("Quark is running Subatomic!")
	fmt.Println("Running until stop command ...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	quark.Close()
}

func botConnected(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateStatus(0, "at the subatomic level")
}
