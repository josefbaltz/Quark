package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
)

//init contains all code ran when the program starts, this includes handling of arguments passed through at start (ex: -t {Discord API Token})
func init() {
	flag.StringVar(&token, "t", "", "Discord API Token")
	flag.StringVar(&gcpcred, "k", "", "GCP JSON Credentials")
	flag.Parse()
}

//Required Variables
var token string
var gcpcred string
var cmdprefix string
var buffer = make([][]byte, 0)
var gcp *datastore.Client
var ctx context.Context
var devmode bool
var gcpErr error

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

	//Build GCP Client
	if gcpcred == "" {
		ctx = context.Background()
		gcpClient, gcpErr := datastore.NewClient(ctx, "quarkbot")
		if gcpErr != nil {
			fmt.Println("--Error--")
			fmt.Println("Failed to create GCP client")
			fmt.Println(gcpErr)
			os.Exit(0)
		} else {
			gcp = gcpClient
			cmdprefix = "q"
			devmode = false
		}
	} else {
		fmt.Println("====WARNING====")
		fmt.Println("YOU ARE RUNNING QUARK IN A DEV ENVIRONMENT")
		fmt.Println("QUARK WAS NOT MEANT TO BE RAN IN THIS WAY")
		fmt.Println("ONLY RUN QUARK IN DEV MODE WHEN IN A CONTROLLED ENVIRONMENT")
		fmt.Println("====WARNING====")
		ctx = context.Background()
		gcpClient, gcpErr := datastore.NewClient(ctx, "quarkbot", option.WithCredentialsFile(gcpcred))
		if gcpErr != nil {
			fmt.Println("--Error--")
			fmt.Println("Failed to create GCP client")
			fmt.Println(gcpErr)
			os.Exit(0)
		} else {
			gcp = gcpClient
			cmdprefix = "qd"
			devmode = true
		}
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
	session.UpdateStatus(0, "Type "+cmdprefix+".help")
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
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".ping") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		session.ChannelMessageSend(event.ChannelID, "Ping!")
		return
	}

	//q.invite
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".invite") {
		if devmode == true {
			session.ChannelMessageSend(event.ChannelID, "Invite command disabled while running in Developer Mode!")
		} else {
			privateChannel, err := session.UserChannelCreate(event.Author.ID)
			if err != nil {
				session.ChannelMessageSend(event.ChannelID, "Oh no, something went wrong!")
				fmt.Println(err)
				return
			}
			session.MessageReactionAdd(event.ChannelID, event.Message.ID, "👍")
			session.ChannelMessageSend(privateChannel.ID, "A hot invite link, fresh from the ovens!")
			session.ChannelMessageSend(privateChannel.ID, "https://discordapp.com/oauth2/authorize?client_id=535127851653922816&permissions=3533888&scope=bot")
			return
		}
	}

	//q.info
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".info") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		var version string
		if devmode == true {
			version = "Dev"
		} else {
			version = "Official Release"
		}
		infoEmbed := &discordgo.MessageEmbed{
			Color:       0xfaa61a, // quarkyellow
			Title:       "Info",
			Description: "Information about Quark",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Owner",
					Value:  "OrangeFlare#1337",
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Contributors",
					Value:  "OrangeFlare, NatCreatess",
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Version",
					Value:  version,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Third-Party Packages Used",
					Value:  "cloud.google.com/go/datastore, github.com/bwmarrin/discordgo, google.golang.org/api/option",
					Inline: false,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://cdn.discordapp.com/avatars/535127851653922816/a6b1eb1cdcc29e0bd7e14228b17a28aa.png?size=512",
			},
		}
		session.ChannelMessageSendEmbed(event.ChannelID, infoEmbed)
		return
	}

	//q.help.basic
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".help.basic") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		helpEmbed := &discordgo.MessageEmbed{
			Color:       0xfaa61a, // quarkyellow
			Title:       "Help",
			Description: "Basic Command Help",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".help",
					Value:  "Shows the help index",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".ping",
					Value:  "Replies with 'Ping!'",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".invite",
					Value:  "Sends you an invite link",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".info",
					Value:  "Shows basic info about Quark",
					Inline: false,
				},
			},
		}
		session.ChannelMessageSendEmbed(event.ChannelID, helpEmbed)
		return
	}

	//q.help.game
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".help.game") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		helpEmbed := &discordgo.MessageEmbed{
			Color:       0xfaa61a, // quarkyellow
			Title:       "Help",
			Description: "Game Command Help",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".game.join",
					Value:  "Joins the game, you should only need to run this once",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".game.upgrade.attack",
					Value:  "Upgrade your attack level for 10 credits",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".game.upgrade.defense",
					Value:  "Upgrade your defense level for 10 credits",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".game.stats",
					Value:  "View your player stats",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".game.fight",
					Value:  "Fight a random enemy",
					Inline: false,
				},
			},
		}
		session.ChannelMessageSendEmbed(event.ChannelID, helpEmbed)
		return
	}

	//q.help
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".help") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		helpEmbed := &discordgo.MessageEmbed{
			Color:       0xfaa61a, // quarkyellow
			Title:       "Help",
			Description: "Welcome to Quark!",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".help.basic",
					Value:  "Display help with basic commands",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   cmdprefix + ".help.game",
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
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".game.join") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)

		//Creates a Key for the requests to follow called 'userKey'
		userKey := datastore.NameKey("User", event.Author.ID, nil)

		//Attempt to get an entity from the Datastore with the title of the User's Discord ID
		//If it finds one it just tells the user they are already registered
		//If it can not find one it creates one with the basic stats and then tells the user that they are now registered
		if err := gcp.Get(ctx, userKey, nil); err != nil {
			user := UserStructure{
				Attack:  4,
				Defense: 4,
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
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".game.upgrade.attack") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)

		userKey := datastore.NameKey("User", event.Author.ID, nil)
		user := UserStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``"+cmdprefix+".game.join``")
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
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".game.upgrade.defense") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)

		userKey := datastore.NameKey("User", event.Author.ID, nil)
		user := UserStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``"+cmdprefix+".game.join``")
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
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".game.stats") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)

		userKey := datastore.NameKey("User", event.Author.ID, nil)
		user := UserStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``"+cmdprefix+".game.join``")
			return
		}

		infoEmbed := &discordgo.MessageEmbed{
			Color: 0xfaa61a, // quarkyellow
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

	//q.game.fight
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".game.fight") {
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)

		userKey := datastore.NameKey("User", event.Author.ID, nil)
		user := UserStructure{}
		monster := MonsterStructure{}

		if err := gcp.Get(ctx, userKey, &user); err != nil {
			fmt.Println("--Warning--")
			fmt.Println("Failed to find user from GCP Datastore")
			fmt.Println(err)
			session.ChannelMessageSend(event.ChannelID, "You are not registered!")
			session.ChannelMessageSend(event.ChannelID, "Please run ``"+cmdprefix+".game.join``")
			return
		}

		rand.Seed(time.Now().UnixNano())
		monster.Attack = int(float64(user.Attack) * ((0.7 + rand.Float64()) * 0.8))
		monster.Defense = int(float64(user.Defense) * ((0.7 + rand.Float64()) * 0.8))
		monster.Reward = int(float64(rand.Intn(4)+8) * ((float64(monster.Attack) + float64(monster.Defense)) / 4) * (0.05 + rand.Float64()*(0.20)))
		fightMonster(user, monster, event, session)
	}

	//q.game.admin.addcredits
	if strings.HasPrefix(strings.ToLower(event.Content), cmdprefix+".game.admin.addcredits") {
		if event.Author.ID != "176108182056206336" { //Only OrangeFlare#1337 can run this command
			return
		}
		session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
		args := strings.Split(strings.TrimPrefix(event.Content, cmdprefix+".game.admin.addcredits "), " ")
		UserID, Credits := args[0], args[1]

		CreditsConv, err := strconv.Atoi(Credits)
		if err != nil {
			session.ChannelMessageSend(event.ChannelID, "An error has occured!")
			return
		}

		addCredits(event, session, UserID, CreditsConv)
		session.ChannelMessageSend(event.ChannelID, "The UserID "+UserID+" has been credited "+Credits+" Credits")
		return
	}
}

//fightMonster contains all code for battles against randomly generated monsters
func fightMonster(user UserStructure, monster MonsterStructure, event *discordgo.MessageCreate, session *discordgo.Session) {
	//HP is based off of Defense
	rand.Seed(time.Now().UnixNano())

	monsterStats := &discordgo.MessageEmbed{
		Color: 0xff0000, // red
		Title: "Random Monster's Statistics",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Attack",
				Value:  strconv.Itoa(monster.Attack),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Defense",
				Value:  strconv.Itoa(monster.Defense),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Level",
				Value:  strconv.Itoa(monster.Attack + monster.Defense),
				Inline: true,
			},
		},
	}

	playerStats := &discordgo.MessageEmbed{
		Color: 0x00ff00, // green
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
		},
	}

	session.ChannelMessageSend(event.ChannelID, event.Author.Username+" vs Random Monster!")
	session.ChannelMessageSendEmbed(event.ChannelID, playerStats)
	session.ChannelMessageSendEmbed(event.ChannelID, monsterStats)

	if rand.Intn(1) == 1 {
		for user.Defense > 0 && monster.Defense > 0 {
			bulkDeleteSlice := []string{}

			var userDamage = int(float64(user.Attack)*(0.8+rand.Float64()*0.4)) / 4
			monster.Defense = monster.Defense - userDamage
			if monster.Defense > 0 {
				sentMessageContent, err := session.ChannelMessageSend(event.ChannelID, "``"+event.Author.Username+" dealt "+strconv.Itoa(userDamage)+" damage to the enemy!``\n``The enemy has "+strconv.Itoa(monster.Defense)+"hp left!``")
				bulkDeleteSlice = append(bulkDeleteSlice, sentMessageContent.ID)
				if err != nil {
					return
				}
			} else {
				session.ChannelMessageSend(event.ChannelID, "``"+event.Author.Username+" dealt "+strconv.Itoa(userDamage)+" damage to the enemy!``\n``The enemy has been slain!``")
				addCredits(event, session, event.Author.ID, monster.Reward)
				session.ChannelMessageSend(event.ChannelID, "Your account has been credited "+strconv.Itoa(monster.Reward)+" Credits!")
				return
			}
			time.Sleep(2 * time.Second)
			session.ChannelMessagesBulkDelete(event.ChannelID, bulkDeleteSlice)

			var monsterDamage = int(float64(monster.Attack)*(0.8+rand.Float64()*0.4)) / 4
			user.Defense = user.Defense - monsterDamage
			if user.Defense > 0 {
				sentMessageContent, err := session.ChannelMessageSend(event.ChannelID, "``The monster dealt "+strconv.Itoa(monsterDamage)+" damage to "+event.Author.Username+"!``\n``"+event.Author.Username+" has "+strconv.Itoa(user.Defense)+"hp left!``")
				bulkDeleteSlice = append(bulkDeleteSlice, sentMessageContent.ID)
				if err != nil {
					return
				}
			} else {
				session.ChannelMessageSend(event.ChannelID, "``The monster dealt "+strconv.Itoa(monsterDamage)+" damage to "+event.Author.Username+"!``\n``"+event.Author.Username+" has been slain!``")
				removeCredits(event, session, event.Author.ID, monster.Reward/5)
				session.ChannelMessageSend(event.ChannelID, "The monster has taken "+strconv.Itoa(monster.Reward/5)+" Credits from you!")
				return
			}
			time.Sleep(2 * time.Second)
			session.ChannelMessagesBulkDelete(event.ChannelID, bulkDeleteSlice)
		}
	} else {
		for user.Defense > 0 && monster.Defense > 0 {
			bulkDeleteSlice := []string{}

			var monsterDamage = int(float64(monster.Attack)*(0.8+rand.Float64()*0.4)) / 4
			user.Defense = user.Defense - monsterDamage
			if user.Defense > 0 {
				sentMessageContent, err := session.ChannelMessageSend(event.ChannelID, "``The monster dealt "+strconv.Itoa(monsterDamage)+" damage to "+event.Author.Username+"!``\n``"+event.Author.Username+" has "+strconv.Itoa(user.Defense)+"hp left!``")
				bulkDeleteSlice = append(bulkDeleteSlice, sentMessageContent.ID)
				if err != nil {
					return
				}
			} else {
				session.ChannelMessageSend(event.ChannelID, "``The monster dealt "+strconv.Itoa(monsterDamage)+" damage to "+event.Author.Username+"!``\n``"+event.Author.Username+" has been slain!``")
				removeCredits(event, session, event.Author.ID, monster.Reward/5)
				session.ChannelMessageSend(event.ChannelID, "The monster has taken "+strconv.Itoa(monster.Reward/5)+" Credits from you!")
				return
			}
			time.Sleep(2 * time.Second)
			session.ChannelMessagesBulkDelete(event.ChannelID, bulkDeleteSlice)

			var userDamage = int(float64(user.Attack)*(0.8+rand.Float64()*0.4)) / 4
			monster.Defense = monster.Defense - userDamage
			if monster.Defense > 0 {
				sentMessageContent, err := session.ChannelMessageSend(event.ChannelID, "``"+event.Author.Username+" dealt "+strconv.Itoa(userDamage)+" damage to the enemy!``\n``The enemy has "+strconv.Itoa(monster.Defense)+"hp left!``")
				bulkDeleteSlice = append(bulkDeleteSlice, sentMessageContent.ID)
				if err != nil {
					return
				}
			} else {
				session.ChannelMessageSend(event.ChannelID, "``"+event.Author.Username+" dealt "+strconv.Itoa(userDamage)+" damage to the enemy!``\n``The enemy has been slain!``")
				addCredits(event, session, event.Author.ID, monster.Reward)
				session.ChannelMessageSend(event.ChannelID, "Your account has been credited "+strconv.Itoa(monster.Reward)+" Credits!")
				return
			}
			time.Sleep(2 * time.Second)
			session.ChannelMessagesBulkDelete(event.ChannelID, bulkDeleteSlice)
		}
	}

}

//addCredits contains all code for adding credits to a user
func addCredits(event *discordgo.MessageCreate, session *discordgo.Session, UserID string, Credits int) {
	var failureMessage = "Failed! Message OrangeFlare#1337"
	userKey := datastore.NameKey("User", UserID, nil)
	user := UserStructure{}

	if err := gcp.Get(ctx, userKey, &user); err != nil {
		fmt.Println("--Warning--")
		fmt.Println("Failed to find user from GCP Datastore")
		fmt.Println(err)
		session.ChannelMessageSend(event.ChannelID, "The User ``"+UserID+"`` is not registered!")
		session.ChannelMessageSend(event.ChannelID, "Please have them run ``"+cmdprefix+".game.join``")
		session.ChannelMessageSend(event.ChannelID, "***__If you believe this to be a mistake, contact OrangeFlare#1337 immediately!__***")
		return
	}

	user.Credits = user.Credits + Credits

	if _, err := gcp.Put(ctx, userKey, &user); err != nil {
		fmt.Println("--Error--")
		fmt.Println("Failed to create GCP client")
		fmt.Println(err)
		session.ChannelMessageSend(event.ChannelID, failureMessage)
	}
	return
}

//takeCredits contains all code for removing credits from a user
func removeCredits(event *discordgo.MessageCreate, session *discordgo.Session, UserID string, Credits int) {
	var failureMessage = "Failed! Message OrangeFlare#1337"
	userKey := datastore.NameKey("User", UserID, nil)
	user := UserStructure{}

	if err := gcp.Get(ctx, userKey, &user); err != nil {
		fmt.Println("--Warning--")
		fmt.Println("Failed to find user from GCP Datastore")
		fmt.Println(err)
		session.ChannelMessageSend(event.ChannelID, "The User ``"+UserID+"`` is not registered!")
		session.ChannelMessageSend(event.ChannelID, "Please have them run ``"+cmdprefix+".game.join``")
		session.ChannelMessageSend(event.ChannelID, "***__If you believe this to be a mistake, contact OrangeFlare#1337 immediately!__***")
		return
	}

	if user.Credits < Credits {
		user.Credits = 0
	} else {
		user.Credits = user.Credits - Credits
	}

	if _, err := gcp.Put(ctx, userKey, &user); err != nil {
		fmt.Println("--Error--")
		fmt.Println("Failed to create GCP client")
		fmt.Println(err)
		session.ChannelMessageSend(event.ChannelID, failureMessage)
	}
	return
}
