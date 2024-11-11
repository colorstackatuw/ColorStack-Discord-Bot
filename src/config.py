import os

from dotenv import load_dotenv

load_dotenv()

GITHUB_TOKEN = os.getenv("GIT_TOKEN")
DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
