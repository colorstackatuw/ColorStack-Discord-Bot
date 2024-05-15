package main

import (
	"context"
	"os"
	"sync"
	"time"

	"ColorStack-Discord-Bot/Utilities"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	discordToken string
	githubToken  string
)
var logger zerolog.Logger
var mutex sync.Mutex

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Fatal().Err(err).Msg("Error loading the .env files")
	}

	discordToken = os.Getenv("DISCORD_TOKEN")
	githubToken = os.Getenv("GIT_TOKEN")
}

func onReady(s *discordgo.Session, event *discordgo.Ready, logger *zerolog.Logger) {
	logger.Info().Str("Username", s.State.User.Username).Msg("Logged In")
}

func onGuildJoin(s *discordgo.Session, event *discordgo.GuildCreate) {
	mutex.Lock()
	defer mutex.Unlock()

	if len(s.State.Guilds) <= 20 {
		logger.Info().Msg("The bot joined a new server!")

		channel, err := s.GuildChannelCreate(event.Guild.ID, "opportunities-bot", discordgo.ChannelTypeGuildText)
		if err != nil {
			logger.Error().
				Err(err).
				Str("guild", event.Guild.Name).
				Msg("Could not create a channel named 'opportunities-bot'")
			if err := s.GuildLeave(event.Guild.ID); err != nil {
				logger.Error().
					Err(err).
					Str("guild", event.Guild.Name).
					Msg("Failed to leave guild after failing to create channel")
			}
			return
		}

		// Create the discord connection
		db := Utilities.NewDBConnector()

		db.writeChannel(event.Guild, channel.ID)
		if _, err := s.ChannelMessageSend(channel.ID, "Hello! I am the ColorStack Bot. I will be posting new job opportunities here."); err != nil {
			logger.Error().Err(err).Str("channel", channel.ID).Msg("Failed to send welcome message")
		}
	} else {
		logger.Info().Msg("We have reached max capacity of 20 servers!")
		if err := s.GuildLeave(event.Guild.ID); err != nil {
			logger.Error().Err(err).Str("guild", event.Guild.Name).Msg("Failed to leave guild after reaching capacity")
		}
	}

}

func onGuildRemove(s *discordgo.Session, event *discordgo.GuildDelete) {

	mutex.Lock()
	defer mutex.Unlock()

	logger.Info().Str("Guild", event.Guild.ID).Msg("The bot has been removed from a server")
	db.deleteServer(event.Guild.ID)
}

func scheduledTask(
	ctx context.Context,
	githubUtilities *Utilities.GitHubUtilities,
	internshipUtilities *Utilities.InternshipUtilities,
) {
	for range time.Tick(60 * time.Second) {
		startTime := time.Now()

		repo, err := githubUtilities.CreateGitHubConnection(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create GitHub connection")
			continue
		}

		lastCommitSHA, err := githubUtilities.GetLastCommit(ctx, repo)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get last saved commit SHA")
			continue
		}

		if githubUtilities.IsNewCommit(lastCommitSHA) {
			logger.Info().Msg("New commit has been found. Finding new jobs...")
			githubUtilities.SetComparison(ctx, repo)

			channelIDs, err := db.GetChannels()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to get channel IDs")
			}

			var isCoop, isSummer bool

			if isCoop {
				jobPostings := githubUtilities.GetCommitChanges("README-Off-Season.md")
				internshipUtilities.GetInternships(bot, channelIDs[:20], jobPostings, false)
			}

			if isSummer {
				jobPostings := githubUtilities.GetCommitChanges("README.md")
				internshipUtilities.GetInternships(bot, channelIDs[:20], jobPostings, true)
			}

			if err := githubUtilities.SetNewCommit(); err != nil {
				logger.Error().Err(err).Msg("Failed to set the new commit")
			}

			logger.Info().Int("total_jobs", internshipUtilities.TotalJobs).Msg("New jobs found!")

			internshipUtilities.ClearJobLinks()
			internshipUtilities.ClearJobCounter()
			githubUtilities.ClearComparison()

			logger.Info().Msg("All jobs have been posted!")
		} else {
			logger.Info().Msg("No new commits found.")
		}

		endTime := time.Now()
		executionTime := endTime.Sub(startTime)
		logger.Info().Dur("execution_time", executionTime).Msg("Task execution time")
	}
}

func main() {
	// Configure lumberjack for log rotation
	logFile := &lumberjack.Logger{
		Filename:   "/app/logs/discord_bot.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     14,
	}
	logger := zerolog.New(logFile).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create the instances
	githubUtilities := Utilities.NewGitHubUtilities("SimplifyJobs/Summer2024-Internships", githubToken)
	internshipUtilities := Utilities.NewInternshipUtilities(true, logFile)

	// Set up Discord Bot
	colorStackBot, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error Creating Discord Session")
	}

	// Set up the different event handlers
	colorStackBot.AddHandler(onReady)
	colorStackBot.AddHandler(onGuildJoin)
	colorStackBot.AddHandler(onGuildRemove)

	// Start the scheduled task
	if err := colorStackBot.Open(); err {
		logger.Fatal().Stack().Err(err).Msg("Error Creating a Discord Connections")
	}
	go scheduleTask(githubUtilities, internshipUtilities, logger)
}
