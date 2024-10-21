"""
Discord Bot 

This bot is responsible for sending internship and co-op opportunities to a Discord channel. 
It uses the GitHub API to track changes in the repository and sends the new opportunities to the Discord channel.

Prerequisites:
- PyGithub: A Python library to access the GitHub API v3.
- Discord: A Python library to interact with the Discord API
- A Discord bot token with the necessary permissions.
- A GitHub personal access token with the necessary permissions.
"""

import asyncio
import logging
import os
from datetime import datetime
from logging.handlers import RotatingFileHandler

import discord
import redis
from DatabaseConnector import DatabaseConnector
from discord.ext import commands, tasks
from dotenv import load_dotenv
from GitHubUtilities import GitHubUtilities
from JobsUtilities import JobsUtilities

load_dotenv()
DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
GITHUB_TOKEN = os.getenv("GIT_TOKEN")

# Global Lock
lock = asyncio.Lock()

# Set up logging: log INFO+ levels to file, appending new entries, with detailed format.
logger = logging.getLogger("discord_bot_logger")
logger.setLevel(logging.INFO)

# Ensure that files are rotated every 5MB, and keep 3 backups.
handler = RotatingFileHandler(filename="/app/logs/discord_bot.log", maxBytes=5 * 1024 * 1024, backupCount=3)
handler.setFormatter(
    logging.Formatter(
        "%(asctime)s - %(levelname)s - %(name)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
)
logger.addHandler(handler)

# Set up the bot
intents = discord.Intents.default()
intents.messages = True
intents.message_content = True
bot = commands.Bot(command_prefix="$", intents=intents)


@tasks.loop(seconds=60)
async def scheduled_task(
    internship_github: GitHubUtilities, newgrad_github: GitHubUtilities, job_utilities: JobsUtilities
):
    """
    A scheduled task that runs every 60 seconds to check for new commits in the GitHub repository.

    Parameters:
        - internship_github: Internship object
        - newgrad_github: New Grad GitHubUtilities object
        - job_utilities: The JobsUtilities object
    """
    async with lock:
        try:
            start_time = datetime.now()

            internship_repo = internship_github.createGitHubConnection()
            internship_sha = internship_github.getSavedSha(internship_repo, False)
            newgrad_repo = newgrad_github.createGitHubConnection()
            newgrad_sha = newgrad_github.getSavedSha(newgrad_repo, True)
            redis_client = redis.Redis(host="localhost", port=6379, db=0)

            # Process all internship
            if internship_github.isNewCommit(internship_repo, internship_sha):
                logger.info("New internship commit has been found. Finding new jobs...")
                internship_github.setComparison(internship_repo, False)

                # Get the channels to send the job postings
                db = DatabaseConnector()
                channel_ids = db.getChannels()

                if internship_github.is_coop:
                    job_postings = internship_github.getCommitChanges("README-Off-Season.md")
                    await job_utilities.getJobs(bot, redis_client, channel_ids[:20], job_postings, "Co-Op")

                if internship_github.is_summer:
                    job_postings = internship_github.getCommitChanges("README.md")
                    await job_utilities.getJobs(bot, redis_client, channel_ids[:20], job_postings, "Summer")

                sha_commit = internship_github.getLastCommit(internship_repo)
                internship_github.setNewCommit(sha_commit, False)
                logger.info(f"There were {job_utilities.total_jobs} new jobs found!")

                # Clear all the cached data
                job_utilities.clearJobLinks()
                job_utilities.clearJobCounter()
                internship_github.clearComparison()

                logger.info("All internship jobs have been posted!")

            # Process all new gradjobs
            if newgrad_github.isNewCommit(newgrad_repo, newgrad_sha):
                logger.info("New grad commit has been found. Finding new jobs...")
                newgrad_github.setComparison(newgrad_repo, True)

                # Get the channels to send the job postings
                db = DatabaseConnector()
                channel_ids = db.getChannels()
                job_postings = newgrad_github.getCommitChanges("README.md")
                await job_utilities.getJobs(bot, redis_client, channel_ids[:20], job_postings, "New Grad")

                sha_commit = newgrad_github.getLastCommit(newgrad_repo)
                newgrad_github.setNewCommit(sha_commit, True)
                logger.info(f"There were {job_utilities.total_jobs} new jobs found!")

                # Clear all the cached data
                job_utilities.clearJobLinks()
                job_utilities.clearJobCounter()
                newgrad_github.clearComparison()

                logger.info("All new grad jobs have been posted!")

        except Exception:
            logger.error("An error occurred in the scheduled task.", exc_info=True)
            await bot.close()
        finally:
            end_time = datetime.now()
            execution_time = end_time - start_time
            logger.info(f"Task execution time: {execution_time}")


@bot.event
async def on_guild_remove(guild: discord.Guild):
    """
    Event that is triggered when the bot is removed from a server.

    Parameters:
        - guild: The guild that the bot has been removed from.
    """
    async with lock:
        logger.info(f"The bot has been removed from: {guild.name}")
        db = DatabaseConnector()
        db.deleteServer(guild)


@bot.event
async def on_guild_join(guild: discord.Guild):
    """
    Event that is triggered when the bot joins a new server.

    Parameters:
        - guild: The guild that the bot has joined.
    """
    try:
        async with lock:
            if len(bot.guilds) <= 20:
                logger.info("The bot joined a new server!")
                channel = await guild.create_text_channel("opportunities-bot")

                db = DatabaseConnector()
                db.writeChannel(guild, channel)
                await channel.send("Hello! I am the ColorStack Bot. I will be posting new job opportunities here.")
            else:
                logger.info("We have reached max capacity of 20 servers!")
                await guild.leave()
    except Exception:
        logger.error(f"Could not create a channel named 'opportunities-bot' in {guild.name}.", exc_info=True)
        await guild.leave()


@scheduled_task.before_loop
async def before_scheduled_task():
    """
    Wait until the bot is ready before starting the loop.
    """
    await bot.wait_until_ready()


@bot.event
async def on_ready():
    """
    Event that is triggered when the bot is ready to start sending messages.
    """
    logger.info(f"Logged in as {bot.user.name}")

    github_internship = GitHubUtilities(
        token=GITHUB_TOKEN, repo_name="SimplifyJobs/Summer2025-Internships", isSummer=True, isCoop=True
    )
    github_newgrad = GitHubUtilities(token=GITHUB_TOKEN, repo_name="SimplifyJobs/New-Grad-Positions")
    job_utilities = JobsUtilities()
    scheduled_task.start(github_internship, github_newgrad, job_utilities)  # Start the loop


if __name__ == "__main__":
    try:
        bot.run(DISCORD_TOKEN)
    except Exception:
        logger.error("Fatal error in main execution:", exc_info=True)
