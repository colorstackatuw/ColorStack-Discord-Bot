package bot

type Bot interface {
	Start() error
	Shutdown() error
	SendMessage(channelID, message string) error
}
