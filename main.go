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
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	discordToken string
	githubToken  string
	mutex        sync.Mutex
	logger       zerolog.Logger
	redisClient  *redis.Client
	ctx          context.Context
	db           *Utilities.DatabaseConnector
	bot          *discordgo.Session
)

/*
init loads environment variables from .env file and initializes tokens.

Parameters: None.
Returns: None.
*/
func init() {

	// Create the logger and location
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	lj := &lumberjack.Logger{
		Filename:   "logs/discord_bot.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	// Create a new zerolog logger with the lumberjack logger as the output
	logger = zerolog.New(lj).With().Timestamp().Logger()
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
		logger.Error().Stack().Err(err).Msg("error creating Discord session")
	}

	bot.AddHandler(onGuildJoin)
	bot.AddHandler(onGuildRemove)
	bot.AddHandler(onReady)

	err = bot.Open()
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error opening connection")
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
	internshipGithub := Utilities.NewGitHubUtilities(
		githubToken,
		"Summer2025-Internships",
		true,
		true,
	)
	newgradGithub := Utilities.NewGitHubUtilities(
		githubToken,
		"New-Grad-Positions",
		false,
		false,
	)
	JobUtilities := Utilities.NewJobUtilities()
	scheduledTask(ctx, internshipGithub, newgradGithub, JobUtilities)
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

		channel, err := s.GuildChannelCreate(
			event.Guild.ID,
			"opportunities-bot",
			discordgo.ChannelTypeGuildText,
		)
		if err != nil {
			logger.Error().
				Stack().
				Err(err).
				Str("guild", event.Guild.Name).
				Msg("Could not create a channel named 'opportunities-bot'")

			if err := s.GuildLeave(event.Guild.ID); err != nil {
				logger.Error().
					Stack().
					Err(err).
					Str("guild", event.Guild.Name).
					Msg("Failed to leave guild after failing to create channel")
			}
			return
		}
		db.WriteChannel(event.Guild, channel)

		if _, err := s.ChannelMessageSend(channel.ID, "Hello! I am the ColorStack Bot. I will be posting new job opportunities here."); err != nil {
			logger.Error().Stack().Err(err).Str("channel", channel.ID).Msg("Failed to send welcome message")
		}
	} else {
		logger.Info().Msg("We have reached max capacity of 20 servers!")

		if err := s.GuildLeave(event.Guild.ID); err != nil {
			logger.Error().Stack().Err(err).Str("guild", event.Guild.Name).Msg("Failed to leave guild after reaching capacity")
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
	if err := db.DeleteServer(event.Guild); err != nil {
		logger.Error().Stack().Err(err).Msgf("Couldn't remove server from database: %s", event.Guild.Name)
	}

}

/*
scheduledTask performs a periodic task to check for new GitHub commits and post new job opportunities.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- githubUtilities: A internship pointer to a Utilities.GitHubUtilities for interacting with GitHub.
- newgradUtilities: A new grad pointer to a Utilities.GitHubUtilities for interacting with GitHub.
- JobUtilities: A pointer to a Utilities.JobUtilities for handling internship postings.
Returns: None.
*/
func scheduledTask(
	ctx context.Context,
	internshipGithub *Utilities.GitHubUtilities,
	newgradGithub *Utilities.GitHubUtilities,
	jobUtilities *Utilities.JobUtilities,
) {
	// Open Connection
	internshipRepo, err := internshipGithub.CreateGitHubConnection(ctx)
	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("Failed to create GitHub connection for internship jobs")
	}
	newgradRepo, err := newgradGithub.CreateGitHubConnection(ctx)
	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("Failed to create GitHub connection for new grad jobs")
	}

	for range time.Tick(60 * time.Second) {
		startTime := time.Now()

		// Get all the commit numbers
		internshipSHA, err := internshipGithub.GetSavedSha(ctx, internshipRepo, false)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("Failed to get internship SHA")
			continue
		}
		newgradSHA, err := newgradGithub.GetSavedSha(ctx, newgradRepo, true)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("Failed to get new grad SHA")
			continue
		}

		// Collect any new internship jobs
		newInternships, err := internshipGithub.IsNewCommit(ctx, internshipRepo, internshipSHA)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("Failed to get the new commit")
			continue
		}

		if newInternships {
			logger.Info().Msg("New commit has been found. Finding new internship jobs...")
			internshipGithub.SetComparison(ctx, internshipRepo, false)

			channelIDs, err := db.GetChannels()
			if err != nil {
				logger.Error().Stack().Err(err).Msg("Failed to get channel IDs")
			}

			if internshipGithub.IsCoop {
				jobPostings := internshipGithub.GetCommitChanges("README-Off-Season.md")
				err := jobUtilities.GetJobs(
					ctx,
					bot,
					channelIDs[:20],
					jobPostings,
					"Co-Op",
					redisClient,
				) 

				if err != nil {
					logger.Error().Stack().Err(err).Msg("Issue collecting jobs")
				}
			}

			if internshipGithub.IsSummer {
				jobPostings := internshipGithub.GetCommitChanges("README.md")
				err := jobUtilities.GetJobs(
					ctx,
					bot,
					channelIDs[:20],
					jobPostings,
					"Summer",
					redisClient,
				)
				
				if err != nil {
					logger.Error().Stack().Err(err).Msg("Issue collecting jobs")
				} 
			}

			// Update the saved commit SHA
			sha_commit, err := internshipGithub.GetLastCommit(ctx, internshipRepo)
			if err != nil {
				logger.Error().Stack().Err(err).Msg("Failed to get the latest commit!")
			}

			if err := internshipGithub.SetNewCommit(sha_commit, false); err != nil {
				logger.Error().Stack().Err(err).Msg("Failed to set the new commit")
			}

			logger.Info().Int("total_jobs", jobUtilities.TotalJobs).Msg("New jobs found!")

			jobUtilities.ClearJobLinks()
			jobUtilities.ClearJobCounter()
			internshipGithub.ClearComparison()

			logger.Info().Msg("All internship jobs have been posted!")
		} else {
			logger.Info().Msg("No new internship commits found.")
		}

		// Collect new grad jobs
		newJobs, err := newgradGithub.IsNewCommit(ctx, newgradRepo, newgradSHA)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("Failed to get the new commit")
			continue
		}

		if newJobs {
			logger.Info().Msg("New commit has been found. Finding new grad jobs...")
			newgradGithub.SetComparison(ctx, newgradRepo, true)

			channelIDs, err := db.GetChannels()
			if err != nil {
				logger.Error().Stack().Err(err).Msg("Failed to get channel IDs")
			}

			jobPostings := internshipGithub.GetCommitChanges("README.md")
			err = jobUtilities.GetJobs(
				ctx,
				bot,
				channelIDs[:20],
				jobPostings,
				"New Grad",
				redisClient,
			) 
			if err != nil {
				logger.Error().Stack().Err(err).Msg("Issue collecting jobs")
			} 

			// Update the saved commit SHA
			sha_commit, err := newgradGithub.GetLastCommit(ctx, newgradRepo)
			if err != nil {
				logger.Error().Stack().Err(err).Msg("Failed to get the latest commit!")
			}

			if err := newgradGithub.SetNewCommit(sha_commit, true); err != nil {
				logger.Error().Stack().Err(err).Msg("Failed to set the new commit")
			}

			logger.Info().Int("total_jobs", jobUtilities.TotalJobs).Msg("New jobs found!")

			jobUtilities.ClearJobLinks()
			jobUtilities.ClearJobCounter()
			internshipGithub.ClearComparison()

			logger.Info().Msg("All new grad jobs have been posted!")
		} else {
			logger.Info().Msg("No new new grad commits found.")
		}

		endTime := time.Now()
		executionTime := endTime.Sub(startTime).Seconds()
		logger.Info().Float64("Execution Time:", executionTime).Msg("End of Task")
	}
}
