package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"cloud.google.com/go/datastore"
	"github.com/bwmarrin/discordgo"
)

//init contains all code ran when the program starts, this includes handling of arguments passed through at start (ex: -t {Discord API Token})
func init() {
	flag.StringVar(&token, "t", "", "Discord API Token")
	flag.Parse()
}

//Required Variables
var token string
var buffer = make([][]byte, 0)

//UserStructure is a structure of the GCP Datastore (NoSQL Schemaless Database) Users have credits, attack, and defense integers
type UserStructure struct {
	Attack  int
	Credits int
	Defense int
}

//MonsterStructure is a structure of the locally generated monsters Monsters have attack, defense, and reward
type MonsterStructure struct {
	Attack  int
	Defense int
	Reward  int
}

//main handles the creation of the Discord client
func main() {
	//Makes sure Token is provided to the program so we don't crash and burn
	if token == "" {
		fmt.Println("--Error--")
		fmt.Println("Your start command should look like:")
		fmt.Println("Quark -t <Discord API Token>")
		os.Exit(0)
	}

	//Build Discord Bot Client
	quark, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		os.Exit(1)
	}

	//Register Event Handler Functions
	quark.AddHandler(botConnected)
	quark.AddHandler(basicCommands)
	quark.AddHandler(gameCommands)

	//Open a Connection to Discord using the Discord Bot Client
	err = quark.Open()
	if err != nil {
		fmt.Println("--Error--")
		fmt.Println(err)
		os.Exit(2)
	}

	//Terminates the program if a terminate signal is recieved from the system (Ctrl+C, Alt+F4, Service Shutdown, etc.)
	fmt.Println("Quark is running Subatomic!")
	fmt.Println("Running until stop command ...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	quark.Close()
}

//botConnected is a collection of code that is ran when the bot connects to the Discord API
func botConnected(session *discordgo.Session, event *discordgo.Ready) {
	//Updates Bot User Status
	session.UpdateStatus(0, "Type q.help")
}

/*
// All of the Command Functions are below this section
// Please keep consistency of code when scanning for commands
// All Functions should Ignore Other Bots, Ignore Case-Sensitivity, and commands should start with q.
// Commands should be structured as follows
// A command that sends a hug reaction image to a specific user should be
//   q.hug {User/Text}
// A sub command should be structured as following
//   q.game.join
// Do not use camel case
//   q.channelInfo (Do not do this)
//   q.channel.info or q.channelinfo (Do This)
// Do not use capitals
//   q.Channel.Info (Do not do this)
//   q.channel.info (Do This)
*/

//basicCommands contains all core commands like help, invite, etcetera
func basicCommands(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.Bot {
		return
	}

	//q.ping
	if strings.HasPrefix(strings.ToLower(event.Content), "q.ping") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		session.ChannelMessageSend(event.ChannelID, "Ping!")
		return
	}

	//q.invite
	if strings.HasPrefix(strings.ToLower(event.Content), "q.invite") {
		privateChannel, err := session.UserChannelCreate(event.Author.ID)
		if err != nil {
			session.ChannelMessageSend(event.ChannelID, "Oh no, something went wrong!")
			fmt.Println(err)
			return
		}
		session.MessageReactionAdd(event.ChannelID, event.Message.ID, "494964052930592768")
		session.ChannelMessageSend(privateChannel.ID, "A hot invite link, fresh from the ovens!")
		session.ChannelMessageSend(privateChannel.ID, "https://discordapp.com/oauth2/authorize?client_id=535127851653922816&permissions=3533888&scope=bot")
		return
	}

	//q.help.basic
	if strings.HasPrefix(strings.ToLower(event.Content), "q.help.basic") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		helpEmbed := &discordgo.MessageEmbed{
			Color:       0xffff00, // yellow
			Title:       "Help",
			Description: "Basic Command Help",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "q.help",
					Value:  "Shows the help index",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.ping",
					Value:  "Replied with Ping!",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.invite",
					Value:  "Sends you an invite link",
					Inline: false,
				},
			},
		}
		session.ChannelMessageSendEmbed(event.ChannelID, helpEmbed)
		return
	}

	//q.help.game
	if strings.HasPrefix(strings.ToLower(event.Content), "q.help.game") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		helpEmbed := &discordgo.MessageEmbed{
			Color:       0xffff00, // yellow
			Title:       "Help",
			Description: "Game Command Help",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "q.game.join",
					Value:  "Joins the game, you should only need to run this once",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.game.upgrade.attack",
					Value:  "Upgrade your attack level for 10 credits",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.game.upgrade.defense",
					Value:  "Upgrade your defense level for 10 credits",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.game.stats",
					Value:  "View your player stats",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.game.fight",
					Value:  "Fight a random enemy",
					Inline: false,
				},
			},
		}
		session.ChannelMessageSendEmbed(event.ChannelID, helpEmbed)
		return
	}

	//q.help
	if strings.HasPrefix(strings.ToLower(event.Content), "q.help") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		helpEmbed := &discordgo.MessageEmbed{
			Color:       0xffff00, // yellow
			Title:       "Help",
			Description: "Welcome to Quark!",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "q.help.basic",
					Value:  "Display help with basic commands",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "q.help.game",
					Value:  "Display help with game commands",
					Inline: false,
				},
			},
		}
		session.ChannelMessageSendEmbed(event.ChannelID, helpEmbed)
		return
	}
}

//gameCommands contains all of the commands relating to the game
func gameCommands(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.Bot {
		return
	}

	var failureMessage = "Failed! Message OrangeFlare#1337"

	//q.game.join
	if strings.HasPrefix(strings.ToLower(event.Content), "q.game.join") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)

		//Build GCP Datastore Client with GCP Project ID quarkbot
		ctx := context.Background()
		gcp, err := datastore.NewClient(ctx, "quarkbot")
		if err != nil {
			fmt.Println("--Error--")
			fmt.Println("Failed to create GCP client")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, failureMessage)
			return
		}

		//Creates a Key for the requests to follow called 'userKey'
		userKey := datastore.NameKey("User", event.Author.ID, nil)

		//Attempt to get an entity from the Datastore with the title of the User's Discord ID
		//If it finds one it just tells the user they are already registered
		//If it can not find one it creates one with the basic stats and then tells the user that they are now registered
		if err := gcp.Get(ctx, userKey, nil); err != nil {
			user := UserStructure{
				Attack:  0,
				Defense: 0,
				Credits: 100,
			}
			if _, err := gcp.Put(ctx, userKey, &user); err != nil {
				fmt.Println("--Warning--")
				fmt.Println("Failed to add user to GCP Datatstore")
				fmt.Println(err)
				session.ChannelMessageSend(event.ChannelID, failureMessage)
				return
			}
			session.ChannelMessageSend(event.ChannelID, "Registered!")
			return
		}
		session.ChannelMessageSend(event.ChannelID, "You are already registered!")
		return
	}

	//q.game.upgrade.attack
	if strings.HasPrefix(strings.ToLower(event.Content), "q.game.upgrade.attack") {
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
		user := UserStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``q.game.join``")
			return
		}

		if user.Credits >= 10 {
			user.Credits = user.Credits - 10
			user.Attack++
		} else {
			session.ChannelMessageSend(event.ChannelID, "You don't have enough credits!")
			return
		}

		session.ChannelMessageSend(event.ChannelID, "Success! Your attack is now level "+strconv.Itoa(user.Attack))
		session.ChannelMessageSend(event.ChannelID, "You now have "+strconv.Itoa(user.Credits)+" credits left!")

		if _, err := gcp.Put(ctx, userKey, &user); err != nil {
			fmt.Println("--Error--")
			fmt.Println("Failed to create GCP client")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, failureMessage)
		}
		return
	}

	//q.game.upgrade.defense
	if strings.HasPrefix(strings.ToLower(event.Content), "q.game.upgrade.defense") {
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
		user := UserStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``q.game.join``")
			return
		}

		if user.Credits >= 10 {
			user.Credits = user.Credits - 10
			user.Defense++
		} else {
			session.ChannelMessageSend(event.ChannelID, "You don't have enough credits!")
			return
		}

		session.ChannelMessageSend(event.ChannelID, "Success! Your defense is now level "+strconv.Itoa(user.Defense))
		session.ChannelMessageSend(event.ChannelID, "You now have "+strconv.Itoa(user.Credits)+" credits left!")

		if _, err := gcp.Put(ctx, userKey, &user); err != nil {
			fmt.Println("--Error--")
			fmt.Println("Failed to create GCP client")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, failureMessage)
		}
		return
	}

	//q.game.stats
	if strings.HasPrefix(strings.ToLower(event.Content), "q.game.stats") {
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
		user := UserStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``q.game.join``")
			return
		}

		infoEmbed := &discordgo.MessageEmbed{
			Color: 0xffff00, // yellow
			Title: event.Author.Username + "'s Statistics",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Attack",
					Value:  strconv.Itoa(user.Attack),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Defense",
					Value:  strconv.Itoa(user.Defense),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Level",
					Value:  strconv.Itoa(user.Attack + user.Defense),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Credits",
					Value:  strconv.Itoa(user.Credits),
					Inline: false,
				},
			},
		}

		session.ChannelMessageSendEmbed(event.Message.ChannelID, infoEmbed)
		return
	}
}
