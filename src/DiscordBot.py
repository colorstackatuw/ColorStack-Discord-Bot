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
async def scheduled_task():
    """
    A scheduled task that runs every 60 seconds to check for new commits in the GitHub repository
    """
    try:
        channel = bot.get_channel(int(CHANNEL_ID))  # Replace with your channel ID
        github_utilitiles = GitHubUtilities(
            token=GITHUB_TOKEN,
            repo_name="SimplifyJobs/Summer2024-Internships",
        )
        repo = github_utilitiles.createGitHubConnection()
        last_saved_commit = github_utilitiles.getCommitLinks()

        internship_utilities = InternshipUtilities(repo, summer=True, co_op=True)

        if github_utilitiles.isNewCommit(repo, last_saved_commit):
            print("New commit has been found. Finding new jobs...")
            if internship_utilities.isSummer:
                await internship_utilities.getSummerInternships(channel)
            if internship_utilities.isCoop:
                await internship_utilities.getCoopInternships(channel)
            github_utilitiles.setNewCommit(github_utilitiles.getLastCommit(repo))
        else:
            print("No new jobs! Time: ", datetime.now().strftime("%Y-%m-%d %H:%M:%S"))

    except Exception as e:
        traceback.print_exc()
        await channel.send(
            "There is a potential issue with the bot! Please check the logs."
        )
        await channel.send("Shutting myself down.....")
        await bot.close()
        print(e)


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
        await channel.send("Successfully joined the discord! Ready to provide jobs")
    else:
        print(f"Could not find channel with ID {CHANNEL_ID}")
    scheduled_task.start()  # Start the loop here


# Run the bot
bot.run(DISCORD_TOKEN)