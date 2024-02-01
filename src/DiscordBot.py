from discord.ext import tasks, commands
import discord
from GitHubUtilities import GitHubUtilities
from InternshipUtilities import InternshipUtilities
import os
from dotenv import load_dotenv
import traceback

load_dotenv() 
DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
CHANNEL_ID = os.getenv("CHANNEL_ID")
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN")

intents = discord.Intents.default()
intents.messages = True  # Track messages
intents.message_content = True  # Track message content
bot = commands.Bot(command_prefix="$",intents=intents)

# Define your scheduled task
@tasks.loop(seconds=60)
async def scheduled_task():
    try:
        channel = bot.get_channel(int(CHANNEL_ID))  # Replace with your channel ID
        github_utilitiles = GitHubUtilities(
            token = GITHUB_TOKEN,
            repo_name = "SimplifyJobs/Summer2024-Internships",
        )
        repo = github_utilitiles.createGitHubConnection()
        last_saved_commit = github_utilitiles.getCommitLinks()

        internship_utilities = InternshipUtilities(repo, summer=True, co_op=True)

        if github_utilitiles.isNewCommit(repo, last_saved_commit):
            if internship_utilities.isSummer:
                await internship_utilities.getSummerInternships(channel)
            if internship_utilities.isCoop:
                await internship_utilities.getCoopInternships(channel)
            github_utilitiles.setNewCommit(github_utilitiles.getLastCommit(repo))

    except Exception as e:
        #await channel.send("There is a potential issue with the bot! Please check the logs.")
        #await channel.send("Shutting myself down.....")
        #await bot.close()
        traceback.print_exc()
        print(e)

# Wait until the bot is ready before starting the loop
@scheduled_task.before_loop
async def before_scheduled_task():
    await bot.wait_until_ready()

@bot.event
async def on_ready():
    print(f'Logged in as {bot.user.name}')
    channel = bot.get_channel(int(CHANNEL_ID))   
    if channel: 
        await channel.send("Successfully joined the discord! Ready to provide jobs")  # Send a message to the channel
    else:
        print(f"Could not find channel with ID {CHANNEL_ID}")
    scheduled_task.start()  # Start the loop here

# Run the bot
bot.run(DISCORD_TOKEN)  # Put this in .env