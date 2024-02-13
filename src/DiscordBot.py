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
from discord.ext import tasks, commands
import discord
from GitHubUtilities import GitHubUtilities
from InternshipUtilities import InternshipUtilities
import os
from dotenv import load_dotenv
from datetime import datetime
import traceback

load_dotenv()
DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
CHANNEL_ID = os.getenv("CHANNEL_ID")
GITHUB_TOKEN = os.getenv("GIT_TOKEN")

intents = discord.Intents.default()
intents.messages = True  # Track messages
intents.message_content = True  # Track message content
bot = commands.Bot(command_prefix="$", intents=intents)


@tasks.loop(seconds=60)
async def scheduled_task(
    github_utilities: GitHubUtilities, internship_utilities: InternshipUtilities
):
    """
    A scheduled task that runs every 60 seconds to check for new commits in the GitHub repository

    Parameters:
        - github_utilities: An instance of the GitHubUtilities class
        - internship_utilities: An instance of the InternshipUtilities class
    """
    try:
        start_time = current_date = datetime.now()
        channel = bot.get_channel(int(CHANNEL_ID))
        repo = github_utilities.createGitHubConnection()
        last_saved_commit = github_utilities.getSavedSha(repo)

        if github_utilities.isNewCommit(repo, last_saved_commit):
            print("New commit has been found. Finding new jobs...")
            github_utilities.setComparison(repo)  # Get the comparison file

            if internship_utilities.is_coop:
                job_postings = github_utilities.getCommitChanges("README-Off-Season.md")
                await internship_utilities.getInternships(
                    channel, job_postings, current_date, False
                )
            if internship_utilities.is_summer:
                job_postings = github_utilities.getCommitChanges("README.md")
                await internship_utilities.getInternships(
                    channel, job_postings, current_date, True
                )
            github_utilities.setNewCommit(github_utilities.getLastCommit(repo))
            print("There were", internship_utilities.total_jobs, " new jobs found!")

            # Clear all the cached data
            internship_utilities.clearJobLinks()
            internship_utilities.clearJobCounter()
            github_utilities.clearComparison()

            print("All jobs have been posted!")
        else:
            print("No new jobs! Time: ", datetime.now().strftime("%Y-%m-%d %H:%M:%S"))

    except Exception as e:
        traceback.print_exc()
        await channel.send(
            "There is a potential issue with the bot! Please check the logs. Shutting down the bot..."
        )
        await bot.close()
        print(e)
    finally:
        end_time = datetime.now()
        execution_time = end_time - start_time  # Calculate execution time
        print(f"Task execution time: {execution_time}")


@scheduled_task.before_loop
async def before_scheduled_task():
    """
    Wait until the bot is ready before starting the loop
    """
    await bot.wait_until_ready()


@bot.event
async def on_ready():
    """
    Event that is triggered when the bot is ready to start sending messages
    """
    print(f"Logged in as {bot.user.name}")
    channel = bot.get_channel(int(CHANNEL_ID))
    if channel:
        print("Successfully joined the discord! Ready to provide jobs")
    else:
        print(f"Could not find channel with ID {CHANNEL_ID}")
    github_utilities = GitHubUtilities(
        token=GITHUB_TOKEN,
        repo_name="SimplifyJobs/Summer2024-Internships",
    )
    internship_utilities = InternshipUtilities(summer=True, coop=True)
    scheduled_task.start(github_utilities, internship_utilities)  # Start the loop here


# Run the bot
bot.run(DISCORD_TOKEN)
