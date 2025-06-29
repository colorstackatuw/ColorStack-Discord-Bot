package bot

import (
	"ColorStack-Discord-Bot/internal/database"
	log "ColorStack-Discord-Bot/internal/logger"
	"ColorStack-Discord-Bot/internal/types"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var mutex sync.Mutex

type SlackBot struct {
	client    *slack.Client
	socket    *socketmode.Client
	botUserID string
}

// NewSlackBot initializes a new SlackBot instance
func NewSlackBot(botToken, appToken string, enableDebug bool) (*SlackBot, error) {
	client := slack.New(
		botToken,
		slack.OptionDebug(enableDebug),
		slack.OptionAppLevelToken(appToken),
	)
	socket := socketmode.New(
		client,
		socketmode.OptionDebug(enableDebug),
	)

	return &SlackBot{
		client: client,
		socket: socket,
	}, nil
}

// Start begins the Slack bot event loop
func (b *SlackBot) Start() error {
	go func() {
		for evt := range b.socket.Events {
			switch ev := evt.Data.(type) {

			case *slack.ConnectedEvent:
				log.Info("Slack bot connected.")

			case *slack.ChannelJoinedEvent:
				b.onChannelJoin(ev)

			case *slack.MemberLeftChannelEvent:
				b.onChannelLeave(ev)

			case *slack.InteractionCallback:
				log.Info("Slack interaction callback received.")
			}
		}
	}()

	return b.socket.Run()
}

// Slack socketmode doesn't expose a close method
func (b *SlackBot) Shutdown() error {
	return nil
}

// SendMessage sends a message to a Slack channel
func (b *SlackBot) SendMessage(channelID, message string) error {
	_, _, err := b.client.PostMessage(channelID, slack.MsgOptionText(message, false))
	return err
}

// onChannelJoin handles logic when the bot joins a new channel
func (b *SlackBot) onChannelJoin(ev *slack.ChannelJoinedEvent) {
	mutex.Lock()
	defer mutex.Unlock()

	log.Info("Slack bot joined a new channel.")

	oracleClient := database.GetDatabaseInstance()
	defer oracleClient.Close()

	channel := ev.Channel
	channelID := channel.ID
	channelName := channel.Name

	err := oracleClient.WriteChannel(channelID, channelName, channelID, types.SLACK)
	if err != nil {
		log.Error("Failed to write Slack channel to DB", err)
		return
	}

	if err := b.SendMessage(channelID, "Hello from the ColorStack Slack Bot! I’ll post new opportunities here."); err != nil {
		log.Error("Failed to send welcome message to Slack channel", err)
	}
}

// onChannelLeave handles logic when the bot is removed from a channel
func (b *SlackBot) onChannelLeave(ev *slack.MemberLeftChannelEvent) {
	mutex.Lock()
	defer mutex.Unlock()

	log.Info("Slack bot removed from a channel.")

	oracleClient := database.GetDatabaseInstance()
	defer oracleClient.Close()

	err := oracleClient.DeleteServer(ev.Channel)
	if err != nil {
		log.Error("Failed to delete Slack channel from DB", err)
	}
}

// Ensure SlackBot implements the Bot interface
var _ Bot = (*SlackBot)(nil)
