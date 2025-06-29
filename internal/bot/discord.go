package bot

import (
	"ColorStack-Discord-Bot/internal/database"
	log "ColorStack-Discord-Bot/internal/logger"
	"ColorStack-Discord-Bot/internal/types"
	"fmt"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type DiscordBot struct {
	session *discordgo.Session
}

var mutex sync.Mutex

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

// Create a new bot
func NewDiscordBot(token string) (*DiscordBot, error) {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatal("Error loading discord token", nil)
	}

	session, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("error creating Discord session", err)
	}

	session.AddHandler(onGuildJoin)
	session.AddHandler(onGuildRemove)
	session.AddHandler(onReady)

	return &DiscordBot{session: session}, err

}

func (b *DiscordBot) Start() error {
	return b.session.Open()
}


func (b *DiscordBot) Shutdown() error {
	return b.session.Close()
}

func (b *DiscordBot) SendMessage(channelID, message string) error {
	_, err := b.session.ChannelMessageSend(channelID, message)
	return err
>>>>>>> e5eb788 (Discord Bot implementation)
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
	logMsg := fmt.Sprintf("Username: %s is signed into the channel!", s.State.User.Username)
	log.Info(logMsg)
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
	

	var guildID string = event.Guild.ID
	if err := oracleClient.DeleteServer(guildID); err != nil {
		logMsg := fmt.Sprintf("Couldn't remove server from database: %s", event.Guild.Name)
		log.Error(logMsg, err)
	}
}

var _ Bot = (*DiscordBot)(nil)
