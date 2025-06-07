package bot

import (
	"ColorStack-Discord-Bot/internal/crawler"
	"ColorStack-Discord-Bot/internal/database"
	log "ColorStack-Discord-Bot/internal/logger"
	"ColorStack-Discord-Bot/internal/types"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	mutex sync.Mutex
	bot   *discordgo.Session
)

/*
init loads environment variables from .env file.

Parameters: None.
Returns: None.
*/
func init() {
	// Loads the .env fies
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

/*
main: Starts the main method of collecting jobs

Returns: None.
*/
func main() {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatal("Error loading discord token", nil)
	}

	bot, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Error("error creating Discord session", err)
	}

	bot.AddHandler(onGuildJoin)
	bot.AddHandler(onGuildRemove)
	bot.AddHandler(onReady)

	err = bot.Open()
	if err != nil {
		log.Error("error opening connection", err)
	}

	// Shut down bot when there is CTRL-C or OS interruption
	defer bot.Close()

	// Wait here until CTRL-C or other term signal is received
	log.Info("Bot is now running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	log.Info("Shutting down...")
}

/*
onReady logs a message when the bot is ready.

Parameters:
- s: A pointer to a discordgo.Session representing the current session.
- event: A pointer to a discordgo.Ready event.
- log: A pointer to a zerolog.log for logging.
Returns: None.
*/
func onReady(s *discordgo.Session, event *discordgo.Ready) {
	logMsg := fmt.Sprintf("Username: %s logged in", s.State.User.Username)
	log.Info(logMsg)

	githubToken := os.Getenv("GIT_TOKEN")
	if githubToken == "" {
		log.Fatal("Error loading discord token", nil)
	}

	internshipGithub := crawler.NewGitHubUtilities(
		githubToken,
		"Summer2025-Internships",
		true,
		true,
	)
	newgradGithub := crawler.NewGitHubUtilities(
		githubToken,
		"New-Grad-Positions",
		false,
		false,
	)
	jobUtilities := crawler.NewJobUtilities()

	ctx := context.Background()
	// Schedule process for all the jobs
	for range time.Tick(120 * time.Second) {
		processJobs(ctx, internshipGithub, newgradGithub, jobUtilities)
	}
}

/*
onGuildJoin handles actions to be taken when the bot joins a new guild.

Parameters:
- s: A pointer to a discordgo.Session representing the current session.
- event: A pointer to a discordgo.GuildCreate event.
Returns: None.
*/
func onGuildJoin(s *discordgo.Session, event *discordgo.GuildCreate) {
	mutex.Lock()
	defer mutex.Unlock()
	log.Info("Request coming in to join server")

	if len(s.State.Guilds) <= 20 {
		log.Info("Adding bot to new server")
		oracleClient := database.GetDatabaseInstance()
		defer oracleClient.Close()

		log.Info("Creating new channel within the server")
		channel, err := s.GuildChannelCreate(
			event.Guild.ID,
			"opportunities-bot",
			discordgo.ChannelTypeGuildText,
		)

		if err != nil {
			logMsg := fmt.Sprintf(
				"Guild: %s  Couldn't create a channel named 'opporutnities-bot'",
				event.Guild.Name,
			)
			log.Error(logMsg, err)

			if err := s.GuildLeave(event.Guild.ID); err != nil {
				logMsg := fmt.Sprintf(
					"Guild: %s Failed to leave guild after failing to create channel",
					event.Guild.Name,
				)
				log.Error(logMsg, err)
			}
			return
		}

		log.Info("Writting server info to database...")
		var guildName string = event.Guild.Name
		var guildID string = event.Guild.ID
		var channelName string = channel.ID
		oracleClient.WriteChannel(guildID, guildName, channelName, types.DISCORD)

		if _, err := s.ChannelMessageSend(channel.ID, "Hello! I am the ColorStack Bot. I will be posting new job opportunities here."); err != nil {
			logMsg := fmt.Sprintf("Channel: %s failed to send welcome message", channel.ID)
			log.Error(logMsg, err)
		}
	} else {
		log.Info("We have reached max capacity of 20 servers!")

		if err := s.GuildLeave(event.Guild.ID); err != nil {
			logMsg := fmt.Sprintf("Guild: %s failed to leave guild after reaching capacity", event.Guild.Name)
			log.Error(logMsg, err)
		}
	}
}

/*
onGuildRemove handles actions to be taken when the bot is removed from a guild.

Parameters:
- s: A pointer to a discordgo.Session representing the current session.
- event: A pointer to a discordgo.GuildDelete event.
Returns: None.
*/
func onGuildRemove(s *discordgo.Session, event *discordgo.GuildDelete) {
	mutex.Lock()
	defer mutex.Unlock()
	logMsg := fmt.Sprintf("Guild: %s The bot has been removed from a server", event.Guild.ID)
	log.Info(logMsg)

	// Connecting to oracle database
	oracleClient := database.GetDatabaseInstance()
	defer oracleClient.Close()

	var guildID string = event.Guild.ID
	if err := oracleClient.DeleteServer(guildID); err != nil {
		logMsg := fmt.Sprintf("Couldn't remove server from database: %s", event.Guild.Name)
		log.Error(logMsg, err)
	}
}
