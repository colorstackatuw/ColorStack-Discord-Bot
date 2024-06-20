import random
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

# How to test the code
# 1) Run the cmd: pytest tests/test_Internship.py

with patch.dict(
    "sys.modules",
    {
        "discord": MagicMock(),
        "discord.ext": MagicMock(),
        "discord.ext.commands": MagicMock(),
    },
):
    from src.JobsUtilities import JobsUtilities


def test_is_within_date_range():
    # Arrange
    internship_util = JobsUtilities()
    random_days = random.randint(1, 6)
    job_date = datetime.now() - timedelta(days=random_days)
    current_date = datetime.now()

    # Act
    is_within_range = internship_util.isWithinDateRange(job_date, current_date)

    # Assert
    assert is_within_range


def test_save_company_name():
    # Arrange
    internship_util = JobsUtilities()
    company_name = "Test Company"

    # Act
    internship_util.saveCompanyName(company_name)

    # Assert
    assert internship_util.previous_job_title == company_name


@pytest.mark.asyncio
async def test_valid_job_posting():
    # Directly create the mock bot
    mock_bot = MagicMock()
    mock_bot.get_channel = MagicMock(return_value=AsyncMock())

    channels = [123456789, 987654321]
    job = """
            | **[Rivian](https://simplify.jobs/c/Rivian)** | UIUC Research Park Intern - Embedded Systems Software Engineer | Urbana, IL | Summer 2024, Fall 2024, Spring 2025 |
            <a href="https://careers.rivian.com/jobs/16695?lang=en-us&icims=1&utm_source=Simplify&ref=Simplify">
            <img src="https://i.imgur.com/w6lyvuC.png" width="84" alt="Apply"></a>
            <a href="https://simplify.jobs/p/707e7608-fdb1-4a20-a1b9-fc69c8c7cb9d?utm_source=GHList">
            <img src="https://i.imgur.com/aVnQdox.png" width="30" alt="Simplify"></a>
            | Feb 05 |
            """
    job_postings = [job]
    current_date = datetime(2024, 1, 8)

    # Create an instance of your class
    instance = JobsUtilities()
    instance.saveCompanyName = MagicMock()
    instance.isWithinDateRange = MagicMock(return_value=True)
    instance.previous_job_title = ""
    instance.job_cache = set()
    instance.total_jobs = 0

    # Use 'with patch' to mock any external dependencies if needed
    with patch("discord.ext.commands.Bot", new=mock_bot):
        # Execute the method
        await instance.getInternships(mock_bot, channels, job_postings, current_date, True)

    # Assertions to verify the behavior
    assert len(instance.job_cache) == 1  # Ensure the job link was added to the cache
    assert instance.total_jobs == 1  # Ensure the job count was incremented
    mock_bot.get_channel.assert_called()  # Ensure get_channel was called for each channel
