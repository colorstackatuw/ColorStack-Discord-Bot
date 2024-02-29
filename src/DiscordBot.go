package main

import (
	"log"
	"os"

	"ColorStack-Discord-Bot/Utilities"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/joho/godotenv"
)

var(
	discordToken string
	githubToken string
)
var logger zerolog.Logger

func init(){
	if err := godotenv.Load(); err{
		log.Fatal().Err(err).Msg("Error loading the .env files")
	}

	discordToken = os.Getenv("DISCORD_TOKEN")
	githubToken = os.Getenv("GIT_TOKEN")
}

func main() {
	githubUtilities := Utilities.NewGitHubUtilities("SimplifyJobs/Summer2024-Internships", githubToken)
	internshipUtilities := Utiliies.NewInternshipUtilities(true)

	// Configure lumberjack for log rotation
	logFile := &lumberjack.Logger{
		Filename:   "/app/logs/discord_bot.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     14,
	}
	logger := zerolog.New(logFile).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Set up Discord Bot
	colorStackBot, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Error Creating Discord Session")
	}

	// Set up the different event handlers
	colorStackBot.AddHandler(onReady)
	colorStackBot.AddHandler(onGuildJoin)
	colorStackBot.AddHandler(onGuildRemove)

	// Start the scheduled task


}

func onReady(s *discordgo.Session, event *discordgo.Ready)
{
	logger.Info().Msg("Logged in as %s", s.State.User.Username)
}

func onGuildJoin(s *discordgo.Seassion, event *discordgo.GuildCreate){

}

func onGuildRemov(s *discordgo.Seassion, event *discordgo.GuildCreate){

}

func scheduleTask(github *Utilities.GitHubUtilities, internship *Utilities.InternshipUtilities)