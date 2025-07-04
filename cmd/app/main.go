package app

import (
	"os"
	"os/signal"
	"syscall"

	"ColorStack-Discord-Bot/internal/bot"
	"ColorStack-Discord-Bot/internal/crawler"
	"ColorStack-Discord-Bot/internal/database"
	log "ColorStack-Discord-Bot/internal/logger"
	jobTypes "ColorStack-Discord-Bot/internal/types"

	"github.com/robfig/cron/v3"
)

var discordBot *bot.DiscordBot
var slackBot *bot.SlackBot

/*
main: Starts the main method of collecting jobs

Returns: None.
*/
func main() {
	// Wait here until CTRL-C or other term signal is received
	log.Info("Bot is now running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	c := cron.New()
	discordBot = bot.NewDiscordBot()
	// slackBot = bot.NewSlackBot()

	discordBot.Start()
	// slackBot.Start()

	defer discordBot.Shutdown()
	// defer slackBot.Shutdown()

	c.AddFunc("@every 2m", processJobs)
	c.Start()

	select {}
}

/*
processJobs performs a periodic task to check for new GitHub commits and post new job opportunities.

Parameters:
Returns: None.
*/
func processJobs() {
	jobChannel := make(chan string)

	githubCrawler := crawler.NewGitHubUtilities(token)
	db := database.GetDatabaseInstance()
	channels, err := db.GetChannels()
	if err != nil {
		log.Error("Failed to retrieve channels", err)
	}

	// Github crawler
	go githubCrawler.GetJobs(jobTypes.COOP, jobChannel)
	go githubCrawler.GetJobs(jobTypes.INTERNSHIP, jobChannel)
	go githubCrawler.GetJobs(jobTypes.NEWGRAD, jobChannel)

	for jobPost := range jobChannel {

		for c := range channels {
			discordBot.SendMessage(c, job)
		}
	}
}
