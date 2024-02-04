"""
Internship Utilities Script

This script provides a set of utilities to interact with the GitHub repository containing the job postings

Prerequisites:
- PyGithub: A Python library to access the GitHub API v3.
- Discord: A Python library to interact with the Discord API
- A GitHub personal access token with the necessary permissions.
"""
import re
from datetime import datetime
import discord
import json
import traceback
from pathlib import Path


class InternshipUtilities:
    FILEPATH = Path("../commits/repository_links_commits.json")
    US_STATES = [
        "AK", "AL", "AR", "AZ", "CA", "CO", "CT", "DC", "DE",
        "FL", "GA", "HI", "IA", "ID", "IL", "IN", "KS", "KY", "LA",
        "MA", "MD", "ME", "MI", "MN", "MO", "MS", "MT", "NC", "ND",
        "NE", "NH", "NJ", "NM", "NV", "NY", "OH", "OK", "OR", "PA",
        "RI", "SC", "SD", "TN", "TX", "UT", "VA", "VT", "WA", "WI",
        "WV", "WY",
    ]
    NOT_US = ["Canada", "UK", "United Kingdom"]

    def __init__(self, summer: bool, co_op: bool):
        self.isSummer = summer
        self.isCoop = co_op

    def binarySearchUS(self, state: str):
        """
        Determine if the state is within the US

        Parameters:
            - state: The acronym of the state
        Returns:
            - bool: True if the state is within the US, False otherwise
        """
        low = 0
        high = len(self.US_STATES) - 1
        while low <= high:
            mid = low + (high - low) // 2
            if self.US_STATES[mid] == state:
                return True
            elif self.US_STATES[mid] > state:
                high = mid - 1
            else:
                low = mid + 1
        return False

    def isWithinDateRange(self, job_posting_date: str, past_week_dates: list[str]) -> bool:
        """
        Filter the commits based on the past week dates

        Parameters:
            - commits: The list of commits
            - past_week_dates: The list of dates from the past week
        Returns:
            - bool: True if the job posting is within the past week, False otherwise
        """
        low = 0
        high = len(past_week_dates) - 1
        while low <= high:
            mid = low + (high - low) // 2
            if past_week_dates[mid].date() == job_posting_date.date():
                return True
            elif past_week_dates[mid].date() > job_posting_date.date():
                high = mid - 1
            else:
                low = mid + 1
        return False

    def getLinks(self) -> dict:
        """
        Retrieve all the saved information for the bot

        Returns:
            - dict: The saved information
        """
        with self.FILEPATH.open("r") as file:
            return json.load(file)

    async def getSummerInternships(
        self,
        channel: discord.TextChannel,
        job_postings: str,
        past_week_dates: list[datetime],
    ):
        """
        Retrieve the summer internships from the GitHub repository

        Parameters:
            - channel: The discord channel to send the job postings
            - job_postings: The list of job postings
            - past_week_dates: The list of dates from the past week compared to current date
        """
        try:
            if len(job_postings) >= 1:
                for job in job_postings:
                    # Grab the data and remove the empty elements
                    non_empty_elements = [
                        element.strip() for element in job.split("|") if element.strip()
                    ]

                    # Verify that job posting date was within past week
                    date_posted = non_empty_elements[-1]
                    current_year = datetime.now().year
                    search_date = f"{date_posted} {current_year}"
                    datetime_format = datetime.strptime(search_date, "%b %d %Y")
                    if not self.isWithinDateRange(datetime_format, past_week_dates):
                        continue

                    # Make sure that the position is still open
                    if "🔒" in non_empty_elements[3]:
                        continue
                    else:
                        job_link = re.search(
                            r'href="([^"]+)"', non_empty_elements[4]
                        ).group(1)

                    # We need to check that the position is within the US and not remote
                    list_locations = []
                    if "<details>" in non_empty_elements[3]:
                        matches = re.findall(
                            r"([A-Za-z\s]+),\s([A-Z]{2})|\bRemote\b",
                            non_empty_elements[2],
                        )
                        for match in matches:
                            if match[0]:
                                city_state = ", ".join(match[:2])
                                list_locations.append(city_state)
                            else:
                                list_locations.append("Remote")
                    elif "</br>" in non_empty_elements[3]:
                        list_locations = non_empty_elements[3].split("</br>")

                    # If there are multiple locations, we need to populate the string correctly
                    if len(list_locations) > 1:
                        location = " | ".join(list_locations)
                    else:
                        is_remote = bool(
                            re.search(r"(?i)\bremote\b", non_empty_elements[3])
                        )
                        location = "Remote" if is_remote else non_empty_elements[3]

                        if location != "Remote":
                            match = re.search(r",\s*(.+)", location)
                            us_state = match.group(1) if match else None

                            # If the location has a US state, we need to check if it's valid
                            # If its not in the US_STATES list, we need to verify if it's in US
                            if match:
                                if not us_state or not self.binarySearchUS(us_state):
                                    continue
                            else:
                                if location in self.NOT_US:
                                    continue

                    if "↳" not in non_empty_elements[1]:
                        match = re.search(r"\[([^\]]+)\]", non_empty_elements[1])
                        company_name = match.group(1) if match else "None"
                        previous_job_title = company_name
                    else:
                        company_name = previous_job_title

                    job_title = non_empty_elements[2]

                    post = (
                        f"**📅 Date Posted:** {date_posted}\n"
                        f"**ℹ️ Company Name:** {company_name}\n"
                        f"**👨‍💻 Job Title:** {job_title}\n"
                        f"**📍 Location:** {location}\n"
                        f"**➡️  When?:** Summer {current_year}\n"
                        f"\n"
                        f"**👉 Job Link:** {job_link}\n"
                        f"\n"
                    )
                    await channel.send(post)
        except Exception as e:
            traceback.print_exc()
            raise e

    async def getCoopInternships(
        self,
        channel: discord.TextChannel,
        job_postings: str,
        past_week_dates: list[str],
    ):
        """
        Retrieve the Co-op internships from the GitHub repository

        Parameters:
            - channel: The discord channel to send the job postings
            - job_postings: The list of job postings
            - past_week_dates: The list of dates from the past week compared to current date
        """
        try:
            if len(job_postings) >= 1:
                for job in job_postings:
                    # Grab the data and remove the empty elements
                    non_empty_elements = [
                        element.strip() for element in job.split("|") if element.strip()
                    ]

                    # Verify that job posting date was within past week
                    date_posted = non_empty_elements[-1]
                    current_year = datetime.now().year
                    search_date = f"{date_posted} {current_year}"
                    datetime_format = datetime.strptime(search_date, "%b %d %Y")
                    if not self.isWithinDateRange(datetime_format, past_week_dates):
                        continue

                    # Make sure that the position is still open
                    if "🔒" in non_empty_elements[5]:
                        continue
                    else:
                        job_link = re.search(
                            r'href="([^"]+)"', non_empty_elements[5]
                        ).group(1)

                    # We need to check that the position is within the US and not remote
                    list_locations = []
                    if "<details>" in non_empty_elements[3]:
                        matches = re.findall(
                            r"([A-Za-z\s]+),\s([A-Z]{2})|\bRemote\b",
                            non_empty_elements[2],
                        )
                        for match in matches:
                            if match[0]:
                                city_state = ", ".join(match[:2])
                                list_locations.append(city_state)
                            else:
                                list_locations.append("Remote")
                    elif "</br>" in non_empty_elements[3]:
                        list_locations = non_empty_elements[3].split("</br>")

                    # If there are multiple locations, we need to populate the string correctly
                    if len(list_locations) > 1:
                        location = " | ".join(list_locations)
                    else:
                        is_remote = bool(
                            re.search(r"(?i)\bremote\b", non_empty_elements[3])
                        )
                        location = "Remote" if is_remote else non_empty_elements[3]

                        if location != "Remote":
                            match = re.search(r",\s*(.+)", location)
                            us_state = match.group(1) if match else None

                            # If the location has a US state, we need to check if it's valid
                            # If its not in the US_STATES list, we need to verify if it's in US
                            if match:
                                if not us_state or not self.binarySearchUS(us_state):
                                    continue
                            else:
                                if location in self.NOT_US:
                                    continue

                    if "↳" not in non_empty_elements[1]:
                        match = re.search(r"\[([^\]]+)\]", non_empty_elements[1])
                        company_name = match.group(1) if match else "None"
                        previous_job_title = company_name
                    else:
                        company_name = previous_job_title

                    job_title = non_empty_elements[2]
                    terms = " |".join(non_empty_elements[4].split(","))

                    post = (
                        f"**📅 Date Posted:** {date_posted}\n"
                        f"**ℹ️ Company Name:** {company_name}\n"
                        f"**👨‍💻 Job Title:** {job_title}\n"
                        f"**📍 Location:** {location}\n"
                        f"**➡️  When?:**  {terms}\n"
                        f"\n"
                        f"**👉 Job Link:** {job_link}\n"
                        f"\n"
                    )
                    await channel.send(post)

        except Exception as e:
            traceback.print_exc()
            raise e
