"""
Internship Utilities Class

This class provides a set of utilities to interact with the GitHub repository containing the job postings

Prerequisites:
- PyGithub: A Python library to access the GitHub API v3.
- Discord: A Python library to interact with the Discord API
- A GitHub personal access token with the necessary permissions.
"""

import asyncio
import logging
import re
import redis
from collections.abc import Iterable
from datetime import datetime, timedelta

import discord


class JobsUtilities:
    NOT_US = ["canada", "uk", "united kingdom", "eu"]

    def __init__(self):
        self.previous_job_title = ""
        self.job_cache = set()
        self.total_jobs = 0

    def clearJobLinks(self) -> None:
        """
        Clear the Co-Op dictionary links.
        """
        self.job_cache = set()

    def clearJobCounter(self) -> None:
        """
        Clear the job counter.
        """
        self.total_jobs = 0

    def isWithinDateRange(self, job_date: datetime, current_date: datetime) -> bool:
        """
        Determine if the job posting is within the past week.

        Parameters:
            - job_date: The date of the job posting.
            - current_date: The current date.
        Returns:
            - bool: True if the job posting is within the past week, False otherwise.
        """
        return timedelta(days=0) <= current_date - job_date <= timedelta(days=7)

    def saveCompanyName(self, company_name: str) -> None:
        """
        Save the previous job title into the class variable.

        Parameters:
            - company_name: The company name.
        """
        self.previous_job_title = company_name

    async def getJobs(
        self,
        bot: discord.ext.commands.Bot,
        redis_client: redis.client.Redis,
        channels: list[int],
        job_postings: Iterable[str],
        term: str,
    ) -> None:
        """
        Retrieve the job postings from the GitHub repository.

        Parameters:
            - bot: The Discord bot.
            - channels: All the channels to send the job postings to
            - job_postings: The list of job postings.
            - term: Timeline of the job posting
        """
        if term not in ["Summer", "Co-Op", "New Grad"]:
            raise ValueError("Term must be one of these: Summer, Coop, NewGrad")

        current_date = datetime.now()
        has_printed = False
        for job in job_postings:
            try:
                # Determine the index of the job link
                job_link_index = 5 if term == "Co-Op" else 4

                # Grab the data and remove the empty elements
                non_empty_elements = [element.strip() for element in job.split("|") if element.strip()]

                # If the job link is already in the cache, we skip the job posting
                job_link = re.search(r'href="([^"]+)"', non_empty_elements[job_link_index]).group(1)
                if job_link in self.job_cache:
                    continue

                # Verify it hasn't been posted
                if redis_client.exists(job_link):
                    continue

                self.job_cache.add(job_link)  # Save the job link

                # If the company name is not present, we need to use the previous company name
                if "↳" not in non_empty_elements[1]:
                    job_header = non_empty_elements[1]
                    start_pos = job_header.find("[") + 1
                    end_pos = job_header.find("]", start_pos)

                    # If the company doesn't have link embedded, we just use the company name
                    if start_pos >= 0 and end_pos >= 0:
                        company_name = job_header[start_pos:end_pos]
                    else:
                        company_name = non_empty_elements[1]
                    self.saveCompanyName(company_name)
                else:
                    company_name = self.previous_job_title

                # Verify that job posting date was within past week
                date_posted = non_empty_elements[-1]
                current_year = datetime.now().year
                search_date = f"{date_posted} {current_year}"
                job_date = datetime.strptime(search_date, "%b %d %Y")
                if not self.isWithinDateRange(job_date, current_date):
                    # Save the previous_job_title in case a "↳" is in US while root is not
                    self.saveCompanyName(company_name)
                    continue

                # We need to check that the position is within the US or remote
                list_locations = []
                location_html = non_empty_elements[3]
                if "<details>" in location_html:
                    start = location_html.find("</summary>") + len("</summary>")
                    end = location_html.find("</details>", start)
                    locations_content = location_html[start:end]
                    for location in locations_content.split("</br>"):
                        location = location.strip()
                        lower_location = location.lower()
                        if location and not any(not_us_country in lower_location for not_us_country in self.NOT_US):
                            list_locations.append(location)

                elif "</br>" in location_html:
                    split_locations = location_html.split("</br>")
                    for location in split_locations:
                        lower_location = location.lower()
                        if not any(not_us_country in lower_location for not_us_country in self.NOT_US):
                            list_locations.append(location)
                elif location_html:
                    location = "Remote" if "remote" in location_html.lower() else location_html
                    lower_location = location.lower()
                    is_outside_us = any(not_us_country in lower_location for not_us_country in self.NOT_US)

                    if location == "Remote" or not is_outside_us:
                        list_locations.append(location)
                    else:
                        self.saveCompanyName(company_name)
                        continue

                if len(list_locations) >= 1:
                    location = " | ".join(list_locations)
                else:
                    self.saveCompanyName(company_name)
                    continue

                job_title = non_empty_elements[2]
                if term == "Co-Op":
                    terms = " |".join(non_empty_elements[4].split(","))
                elif term == "Summer":
                    terms = "Summer 2025"

                post = ""
                if not has_printed:
                    post += f"# {term} Postings!\n\n"
                    has_printed = True

                post += (
                    f"**📅 Date Posted:** {date_posted}\n"
                    f"**ℹ️ Company:** __{company_name}__\n"
                    f"**👨‍💻 Job Title:** {job_title}\n"
                    f"**📍 Location:** {location}\n"
                )
                if term != "New Grad":
                    post += f"**➡️  When?:**  {terms}\n"
                post += f"**👉 Job Link:** <{job_link}>\n" f"{'-' * 153}"
                self.total_jobs += 1

                # Add the job link to redis database
                redis_client.set(job_link, datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
                logging.info("Added the job link to redis!")

                # Send the job posting to the Discord channel
                coroutines = (bot.get_channel(channel).send(post) for channel in channels if bot.get_channel(channel))
                await asyncio.gather(*coroutines)
            except Exception as e:
                logging.exception("Failed to process job posting: %s\nJob: %s", e, job)
                continue
