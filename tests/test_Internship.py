import random
from datetime import datetime, timedelta
from unittest.mock import AsyncMock
import discord
import discord.ext as dpytest
import discord.ext.commands as commands
import pytest
import pytest_asyncio
from discord.ext.commands import Cog, command
from src.InternshipUtilities import InternshipUtilities

# How to test the code
# 1) Run the cmd: pytest tests/test_InternshipTests.py

def test_is_within_date_range():
    # Arrange
    internship_util = InternshipUtilities(True, False)
    random_days = random.randint(1, 6)
    job_date = datetime.now() - timedelta(days=random_days)
    current_date = datetime.now()

    # Act
    is_within_range = internship_util.isWithinDateRange(job_date, current_date)

    # Assert
    assert is_within_range


def test_save_company_name():
    # Arrange
    internship_util = InternshipUtilities(True, False)
    company_name = "Test Company"

    # Act
    internship_util.saveCompanyName(company_name)

    # Assert
    assert internship_util.previous_job_title == company_name


class Misc(Cog):
    @command()
    async def send(self, ctx):
        await ctx.send("Testing !")


@pytest_asyncio.fixture
async def bot():
    # Setup
    intents = discord.Intents.default()
    intents.members = True
    intents.message_content = True
    bot = commands.Bot(command_prefix="!", intents=intents)
    # setup the commands
    await bot._async_setup_hook()
    await bot.add_cog(Misc())

    dpytest.configure(bot)

    yield bot

    # Teardown
    await dpytest.empty_queue()


@pytest.mark.asyncio
async def test_get_internships():
    # Arrange
    internship_util = InternshipUtilities(True, False)
    channel = AsyncMock()
    channel.send = AsyncMock()  # Directly assign AsyncMock to channel's send method
    job_postings = [
        (
            "| **[Western Digital](https://simplify.jobs/c/Western-Digital)** | Advanced Manufacturing Apprentice | "
            'Fremont, CA | Spring 2024 | <a href="https://jobs.smartrecruiters.com/WesternDigital/743999965654172?utm_source=Simplify&ref=Simplify">'
            '<img src="https://i.imgur.com/w6lyvuC.png" width="84" alt="Apply"></a> <a href="https://simplify.jobs/p/b84301fb-e173-45b1-a080-82f5e444af44?utm_source=GHList">'
            '<img src="https://i.imgur.com/aVnQdox.png" width="30" alt="Simplify"></a> | Feb 05 |'
        )
    ]
    current_date = datetime.now()
    isSummer = True

    # Act
    await internship_util.getInternships(channel, job_postings, current_date, isSummer)

    # Diagnostic send call (to verify mock setup)
    await channel.send("Diagnostic message")

    # Assert
    channel.send.assert_called()
