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
import logging
import os
from datetime import datetime

import discord
from discord.ext import commands, tasks
from dotenv import load_dotenv
from discord.ext import commands, tasks
from GitHubUtilities import GitHubUtilities
from InternshipUtilities import InternshipUtilities

load_dotenv()
DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
CHANNEL_ID = os.getenv("CHANNEL_ID")
GITHUB_TOKEN = os.getenv("GIT_TOKEN")

# Set up logging: log INFO+ levels to file, appending new entries, with detailed format.
logging.basicConfig(
    filename="/app/logs/discord_bot.log",
    filemode="a",
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)

intents = discord.Intents.default()
intents.messages = True  # Enable message tracking
intents.message_content = True  # Enable message content tracking
bot = commands.Bot(command_prefix="$", intents=intents)


@tasks.loop(seconds=60)
async def scheduled_task(
    github_utilities: GitHubUtilities, internship_utilities: InternshipUtilities
):
    """
    A scheduled task that runs every 60 seconds to check for new commits in the GitHub repository.
    """
    try:
        start_time = datetime.now()
        channel = bot.get_channel(int(CHANNEL_ID))
        repo = github_utilities.createGitHubConnection()
        last_saved_commit = github_utilities.getSavedSha(repo)

        if github_utilities.isNewCommit(repo, last_saved_commit):
            logging.info("New commit has been found. Finding new jobs...")
            github_utilities.setComparison(repo)  # Get the comparison file

            if internship_utilities.is_coop:
                job_postings = github_utilities.getCommitChanges("README-Off-Season.md")
                await internship_utilities.getInternships(
                    channel, job_postings, start_time, False
                )
            if internship_utilities.is_summer:
                job_postings = github_utilities.getCommitChanges("README.md")
                await internship_utilities.getInternships(
                    channel, job_postings, start_time, True
                )
            github_utilities.setNewCommit(github_utilities.getLastCommit(repo))
            logging.info(
                f"There were {internship_utilities.total_jobs} new jobs found!"
            )

            # Clear all the cached data
            internship_utilities.clearJobLinks()
            internship_utilities.clearJobCounter()
            github_utilities.clearComparison()

            logging.info("All jobs have been posted!")
        else:
            logging.info(
                f"No new jobs! Time: {start_time.strftime('%Y-%m-%d %H:%M:%S')}"
            )
    except Exception:
        logging.error("An error occurred in the scheduled task.", exc_info=True)
        await channel.send(
            "There is a potential issue with the bot! Please check the logs."
        )
        await bot.close()
    finally:
        end_time = datetime.now()
        execution_time = end_time - start_time
        logging.info(f"Task execution time: {execution_time}")


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
    logging.info(f"Logged in as {bot.user.name}")
    channel = bot.get_channel(int(CHANNEL_ID))
    if channel:
        logging.info("Successfully joined the Discord! Ready to provide jobs.")
    else:
        logging.error(f"Could not find channel with ID {CHANNEL_ID}")

    github_utilities = GitHubUtilities(
        token=GITHUB_TOKEN, repo_name="SimplifyJobs/Summer2024-Internships"
    )
    internship_utilities = InternshipUtilities(summer=True, coop=True)
    scheduled_task.start(github_utilities, internship_utilities)  # Start the loop


if __name__ == "__main__":
    try:
        bot.run(DISCORD_TOKEN)
    except Exception:
        logging.error("Fatal error in main execution:", exc_info=True)
