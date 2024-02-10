"""
Internship Utilities Script

This script provides a set of utilities to interact with the GitHub repository containing the job postings

Prerequisites:
- PyGithub: A Python library to access the GitHub API v3.
- Discord: A Python library to interact with the Discord API
- A GitHub personal access token with the necessary permissions.
"""
import re
from datetime import datetime, timedelta
import discord
import traceback
from collections.abc import Iterable
from pathlib import Path


class InternshipUtilities:
    FILEPATH = Path("../commits/repository_links_commits.json") 
    NOT_US = ["Canada", "UK", "United Kingdom", "EU"]

    def __init__(self, summer: bool, coop: bool):
        self.is_summer = summer
        self.is_coop = coop
        self.previous_job_title = ""
        self.coop_job_links = dict()
        self.total_jobs = 0

    def clearJobLinks(self) -> None:
        """
        Clear the Co-Op dictionary links
        """
        self.coop_job_links = dict()
    
    def clearJobCounter(self) -> None:
        """
        Clear the job counter
        """
        self.total_jobs = 0

    def isWithinDateRange(self, job_date: datetime, current_date: datetime) -> bool:
        """
        Determine if the job posting is within the past week

        Parameters:
            - job_date: The date of the job posting
            - current_date: The current date
        Returns:
            - bool: True if the job posting is within the past week, False otherwise
        """
        return timedelta(days=0) <= current_date - job_date <= timedelta(days=7)

    def saveCompanyName(self, company_name: str) -> None:
        """
        Save the previous job title into the class variable

        Parameters:
            - company_name: The company name
        """
        self.previous_job_title = company_name

    async def getInternships(
        self,
        channel: discord.TextChannel,
        job_postings: Iterable[str],
        current_date: datetime,
        is_summer: bool,
    ):
        """
        Retrieve the Summer or Co-op internships from the GitHub repository

        Parameters:
            - channel: The discord channel to send the job postings
            - job_postings: The list of job postings
            - current_date: The current date
            - is_summer: A boolean to record a job if it's summer or co-op internships
        """
        try:
            for job in job_postings:
                # Determine the index of the job link
                job_link_index = 4 if is_summer else 5

                # Grab the data and remove the empty elements
                non_empty_elements = [
                    element.strip() for element in job.split("|") if element.strip()
                ]

                # If the company name is not present, we need to use the previous company name
                if "↳" not in non_empty_elements[1]:
                    match = re.search(r"\[([^\]]+)\]", non_empty_elements[1])

                    # If the company doesn't have link embedded, we just use the company name
                    company_name = match.group(1) if match else non_empty_elements[1]
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
                    locations_content = re.search(
                        r"(?<=</summary>)(.*?)(?=</details>)",
                        location_html,
                        flags=re.DOTALL,
                    ).group(1)
                    for location in locations_content.split("</br>"):
                        location = location.strip()
                        if location and not any(
                            not_us_country in location for not_us_country in self.NOT_US
                        ):
                            list_locations.append(location)

                elif "</br>" in location_html:
                    split_locations = location_html.split("</br>")
                    for location in split_locations:
                        if not any(
                            not_us_country in location for not_us_country in self.NOT_US
                        ):
                            list_locations.append(location)
                elif location_html:
                    location = (
                        "Remote"
                        if re.search(r"(?i)\bremote\b", location_html)
                        else location_html
                    )
                    is_outside_us = any(
                        not_us_country in location for not_us_country in self.NOT_US
                    )

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

                job_link = re.search(
                    r'href="([^"]+)"', non_empty_elements[job_link_index]
                ).group(1)

                job_title = non_empty_elements[2]

                if is_summer:
                    terms = "Summer" + " " + str(current_year)

                    # If the job link is already in the dictionary, we skip the job
                    if job_link in self.coop_job_links:
                        continue
                else:
                    terms = " |".join(non_empty_elements[4].split(","))
                    self.coop_job_links[job_link] = None  # Save the job link

                post = (
                    f"**📅 Date Posted:** {date_posted}\n"
                    f"**ℹ️ Company:** {company_name}\n"
                    f"**👨‍💻 Job Title:** {job_title}\n"
                    f"**📍 Location:** {location}\n"
                    f"**➡️  When?:**  {terms}\n"
                    f"\n"
                    f"**👉 Job Link:** {job_link}\n"
                    f"\n\n\n"
                )
                self.total_jobs += 1
                await channel.send(post)

        except Exception as e:
            traceback.print_exc()
            raise e
