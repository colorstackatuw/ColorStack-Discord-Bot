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
from DatabaseConnector import DatabaseConnector
from discord import ChannelType
from discord.ext import commands, tasks
from dotenv import load_dotenv
from GitHubUtilities import GitHubUtilities
from InternshipUtilities import InternshipUtilities

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
async def scheduled_task(github_utilities: GitHubUtilities, internship_utilities: InternshipUtilities):
    """
    A scheduled task that runs every 60 seconds to check for new commits in the GitHub repository.

    Parameters:
        - github_utilities: The GitHubUtilities object
        - internship_utilities: The InternshipUtilities object
    """
    async with lock:
        try:
            start_time = datetime.now()
            repo = github_utilities.createGitHubConnection()
            last_saved_commit = github_utilities.getSavedSha(repo)

            if github_utilities.isNewCommit(repo, last_saved_commit):
                logger.info("New commit has been found. Finding new jobs...")
                github_utilities.setComparison(repo)  # Get the comparison file

                # Get the channels to send the job postings
                db = DatabaseConnector()
                channel_ids = db.getChannels()

                if internship_utilities.is_coop:
                    job_postings = github_utilities.getCommitChanges("README-Off-Season.md")
                    await internship_utilities.getInternships(bot, channel_ids[:20], job_postings, start_time, False)

                if internship_utilities.is_summer:
                    job_postings = github_utilities.getCommitChanges("README.md")
                    await internship_utilities.getInternships(bot, channel_ids[:20], job_postings, start_time, True)

                github_utilities.setNewCommit(github_utilities.getLastCommit(repo))
                logger.info(f"There were {internship_utilities.total_jobs} new jobs found!")

                # Clear all the cached data
                internship_utilities.clearJobLinks()
                internship_utilities.clearJobCounter()
                github_utilities.clearComparison()

                logger.info("All jobs have been posted!")
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
    async with lock:
        logger.info("The bot joined a new server!")
        found_channel = None

        for channel in guild.channels:
            if channel.name == "opportunities-bot" and channel.type == ChannelType.text:
                found_channel = channel
                break

        if found_channel:
            db = DatabaseConnector()
            db.writeChannel(guild, found_channel)
            logger.info(f"Found 'opportunities-bot' channel in {guild.name}")
            await bot.get_channel(found_channel.id).send("Hello! I am the ColorStack Bot. I will be posting new job opportunities here.")
        else:
            logger.error(f"Could not find a channel named 'opportunities-bot' in {guild.name}.")
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

    github_utilities = GitHubUtilities(token=GITHUB_TOKEN, repo_name="SimplifyJobs/Summer2024-Internships")
    internship_utilities = InternshipUtilities(summer=True, coop=True)
    scheduled_task.start(github_utilities, internship_utilities)  # Start the loop


if __name__ == "__main__":
    try:
        bot.run(DISCORD_TOKEN)
    except Exception:
        logger.error("Fatal error in main execution:", exc_info=True)
