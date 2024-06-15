package main

import (
	"ColorStack-Discord-Bot/src/Utilities"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var (
	discordToken string
	githubToken  string
	mutex        sync.Mutex
	logger       *zerolog.Logger
	redisClient  *redis.Client
	ctx          context.Context
	db 			 *Utilities.DatabaseConnector
	bot 		 *discordgo.Session
)

/*
init loads environment variables from .env file and initializes tokens.

Parameters: None.
Returns: None.
*/
func init() {
	if err := godotenv.Load(); err != nil {
		logger.Fatal().Err(err).Msg("Error loading the .env files")
	}

	ctx = context.Background()
	db = Utilities.NewDatabaseConnector()
	discordToken = os.Getenv("DISCORD_TOKEN")
	githubToken = os.Getenv("GIT_TOKEN")

	//Set up Redis Database
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logger.Error().Stack().Err(err).Msg("Cannot connect to the redis database")
	} else {
		logger.Info().Msg("We have connected to the Redis Database!")
	}
}

/*
main: Starts the main method of collecting jobs

Returns: None.
*/
func main() {
	bot, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		logger.Error().Stack().Msg("error creating Discord session")
	}

	bot.AddHandler(onGuildJoin)
	bot.AddHandler(onGuildRemove)
	bot.AddHandler(onReady)

	err = bot.Open()
	if err != nil {
		logger.Error().Stack().Msg("error opening connection")
	}

	// Shut down bot when there is CTRL-C or OS interruption
	defer bot.Close()

    // Wait here until CTRL-C or other term signal is received
    logger.Info().Msg("Bot is now running. Press CTRL+C to exit.")
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-stop

    logger.Info().Msg("Shutting down...")
}

/*
onReady logs a message when the bot is ready.

Parameters:
- s: A pointer to a discordgo.Session representing the current session.
- event: A pointer to a discordgo.Ready event.
- logger: A pointer to a zerolog.Logger for logging.
Returns: None.
*/
func onReady(s *discordgo.Session, event *discordgo.Ready) {
	logger.Info().Str("Username", s.State.User.Username).Msg("Logged In")

	githubUtilities := Utilities.NewGitHubUtilities(githubToken, "SimplifyJobs/Summer2024-Internships")
	internshipUtilities := Utilities.NewInternshipUtilities(true)

	// Start the tasks
	scheduledTask(ctx, githubUtilities, internshipUtilities)
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

	if len(s.State.Guilds) <= 20 {
		logger.Info().Msg("The bot joined a new server!")
		db := Utilities.NewDatabaseConnector()

		channel, err := s.GuildChannelCreate(event.Guild.ID, "opportunities-bot", discordgo.ChannelTypeGuildText)
		if err != nil {
			logger.Error().
				Stack().
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
		db.WriteChannel(event.Guild, channel)
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

	logger.Info().Str("Guild", event.Guild.ID).Msg("The bot has been removed from a server")
	db.DeleteServer(event.Guild)
}

/*
scheduledTask performs a periodic task to check for new GitHub commits and post new job opportunities.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- githubUtilities: A pointer to a Utilities.GitHubUtilities for interacting with GitHub.
- internshipUtilities: A pointer to a Utilities.InternshipUtilities for handling internship postings.
Returns: None.
*/
func scheduledTask(
	ctx context.Context,
	githubUtilities *Utilities.GitHubUtilities,
	internshipUtilities *Utilities.InternshipUtilities,
) {
	// Open Connection
	repo, err := githubUtilities.CreateGitHubConnection(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create GitHub connection")
	}

	for range time.Tick(60 * time.Second) {
		startTime := time.Now()

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
				internshipUtilities.GetInternships(bot, channelIDs[:20], jobPostings, false, redisClient)
			}

			if isSummer {
				jobPostings := githubUtilities.GetCommitChanges("README.md")
				internshipUtilities.GetInternships(bot, channelIDs[:20], jobPostings, true, redisClient)
			}

			if err := githubUtilities.SetNewCommit(lastCommitSHA); err != nil {
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
