package main

import (
	"ColorStack-Discord-Bot/src/utilities"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "ColorStack-Discord-Bot/logging"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	discordToken string
	githubToken  string
	mutex        sync.Mutex
	redisClient  *redis.Client
	ctx          context.Context
	db           *utilities.DatabaseService
	bot          *discordgo.Session
)

/*
init loads environment variables from .env file and initializes tokens.

Parameters: None.
Returns: None.
*/
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	ctx = context.Background()
	discordToken = os.Getenv("DISCORD_TOKEN")
	githubToken = os.Getenv("GIT_TOKEN")

	// Set up Redis Database
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Error("Cannot connect to the redis database", err)
	} else {
		log.Info("We have connected to the Redis Database!")
	}
}

/*
main: Starts the main method of collecting jobs

Returns: None.
*/
func main() {
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
	internshipGithub := utilities.NewGitHubUtilities(
		githubToken,
		"Summer2025-Internships",
		true,
		true,
	)
	newgradGithub := utilities.NewGitHubUtilities(
		githubToken,
		"New-Grad-Positions",
		false,
		false,
	)
	jobUtilities := utilities.NewjobUtilities()
	scheduledTask(ctx, internshipGithub, newgradGithub, jobUtilities)
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
		log.Info("The bot joined a new server!")
		db := utilities.NewDatabaseConnector()

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
		db.WriteChannel(event.Guild, channel)

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
	if err := db.DeleteServer(event.Guild); err != nil {
		logMsg := fmt.Sprintf("Couldn't remove server from database: %s", event.Guild.Name)
		log.Error(logMsg, err)
	}
}

/*
scheduledTask performs a periodic task to check for new GitHub commits and post new job opportunities.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- githubUtilities: A internship pointer to a utilities.GitHubUtilities for interacting with GitHub.
- newgradUtilities: A new grad pointer to a utilities.GitHubUtilities for interacting with GitHub.
- jobUtilities: A pointer to a utilities.jobUtilities for handling internship postings.
Returns: None.
*/
func scheduledTask(
	ctx context.Context,
	internshipGithub *utilities.GitHubUtilities,
	newgradGithub *utilities.GitHubUtilities,
	jobUtilities *utilities.JobUtilities,
) {
	// Open Connection
	internshipRepo, err := internshipGithub.CreateGitHubConnection(ctx)
	if err != nil {
		log.Error("Failed to create GitHub connection for internship jobs", err)
	}
	newgradRepo, err := newgradGithub.CreateGitHubConnection(ctx)
	if err != nil {
		log.Fatal("Failed to create GitHub connection for new grad jobs", err)
	}

	for range time.Tick(60 * time.Second) {
		startTime := time.Now()

		// Get all the commit numbers
		internshipSHA, err := internshipGithub.GetSavedSha(ctx, internshipRepo, false)
		if err != nil {
			log.Error("Failed to get internship SHA", err)
			continue
		}
		newgradSHA, err := newgradGithub.GetSavedSha(ctx, newgradRepo, true)
		if err != nil {
			log.Error("Failed to get new grad SHA", err)
			continue
		}

		// Collect any new internship jobs
		newInternships, err := internshipGithub.IsNewCommit(ctx, internshipRepo, internshipSHA)
		if err != nil {
			log.Error("Failed to get the new commit", err)
			continue
		}

		if newInternships {
			log.Info("New commit has been found. Finding new internship jobs...")
			internshipGithub.SetComparison(ctx, internshipRepo, false)

			channelIDs, err := db.GetChannels()
			if err != nil {
				log.Error("Failed to get channel IDs", err)
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
					log.Error("Issue collecting jobs", err)
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
					log.Error("Issue collecting jobs", err)
				}
			}

			// Update the saved commit SHA
			sha_commit, err := internshipGithub.GetLastCommit(ctx, internshipRepo)
			if err != nil {
				log.Error("Failed to get the latest commit!", err)
			}

			if err := internshipGithub.SetNewCommit(sha_commit, false); err != nil {
				log.Error("Failed to set the new commit", err)
			}

			logMsg := fmt.Sprintf("New %d jobs found!", jobUtilities.TotalJobs)
			log.Info(logMsg)

			jobUtilities.ClearJobLinks()
			jobUtilities.ClearJobCounter()
			internshipGithub.ClearComparison()

			log.Info("All internship jobs have been posted!")
		} else {
			log.Info("No new internship commits found.")
		}

		// Collect new grad jobs
		newJobs, err := newgradGithub.IsNewCommit(ctx, newgradRepo, newgradSHA)
		if err != nil {
			log.Error("Failed to get the new commit", err)
			continue
		}

		if newJobs {
			log.Info("New commit has been found. Finding new grad jobs...")
			newgradGithub.SetComparison(ctx, newgradRepo, true)

			channelIDs, err := db.GetChannels()
			if err != nil {
				log.Error("Failed to get channel IDs", err)
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
				log.Error("Issue collecting jobs", err)
			}

			// Update the saved commit SHA
			sha_commit, err := newgradGithub.GetLastCommit(ctx, newgradRepo)
			if err != nil {
				log.Error("Failed to get the latest commit!", err)
			}

			if err := newgradGithub.SetNewCommit(sha_commit, true); err != nil {
				log.Error("Failed to set the new commit", err)
			}

			logMsg := fmt.Sprintf("New %d jobs found!", jobUtilities.TotalJobs)
			log.Info(logMsg)

			jobUtilities.ClearJobLinks()
			jobUtilities.ClearJobCounter()
			internshipGithub.ClearComparison()

			log.Info("All new grad jobs have been posted!")
		} else {
			log.Info("No new new grad commits found.")
		}

		endTime := time.Now()
		executionTime := endTime.Sub(startTime).Seconds()
		logMsg := fmt.Sprintf("Execution Time: %.2f seconds", executionTime)
		log.Info(logMsg)
	}
}
