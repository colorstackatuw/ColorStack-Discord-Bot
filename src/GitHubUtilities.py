"""
GitHub Utilities Script

This script provides a set of utilities to interact with GitHub repositories using the PyGithub library. 
It includes functionalities to establish a connection to a specified GitHub repository, update and retrieve 
the last commit information, and check for new commits.

Prerequisites:
- PyGithub: A Python library to access the GitHub API v3.
- A GitHub personal access token with the necessary permissions.
"""
from github import Auth, Github
import github
import json
from pathlib import Path
from datetime import datetime, timedelta


class GitHubUtilities:
    FILEPATH = Path("../commits/repository_links_commits.json")

    def __init__(self, token, repo_name="SimplifyJobs/Summer2024-Internships"):
        self.repo_name = repo_name
        self.github = Github(auth=Auth.Token(token))

    def createGitHubConnection(self):
        """
        Create a connection to the specified GitHub repository
        """
        return self.github.get_repo(self.repo_name)

    def setNewCommit(self, commit: str):
        """
        Save the last commit information to prevent duplicate job postings

        Parameters:
            - commit: The last commit information
        """
        with self.FILEPATH.open("r") as file:
            data_json = json.load(file)

        data_json["last_commit"] = commit

        with self.FILEPATH.open("w") as file:
            json.dump(data_json, file)

    def getLastCommit(self, repo: github.Repository.Repository) -> str:
        """
        Retrieve the last commit information based on the repository

        Parameters:
            - repo: The GitHub repository
        Returns:
            - str: The last commit hexadecimal information on Github repository
        """
        branch = repo.get_branches()[0]  # May need to be changed in future
        return branch.commit.sha

    def getCommitLinks(self) -> str:
        """
        Retrieve the last commit information from the saved file

        Returns:
            - str: The last commit hexadecimal information
        """
        with self.FILEPATH.open("r") as file:
            return json.load(file)["last_commit"]

    def isNewCommit(self, repo: github.Repository.Repository, last_commit: str) -> bool:
        """
        Determine if there is a new commit on the GitHub repository

        Parameters:
            - repo: The GitHub repository
            - last_commit: The last commit hexadecimal information
        Returns:
            - bool: True if there is a new commit, False otherwise
        """
        return last_commit != self.getLastCommit(repo)

    def isWithinDateRange(
        self, job_posting_date: str, past_week_dates: list[datetime]
    ) -> bool:
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
            if past_week_dates[mid] == job_posting_date:
                return True
            elif past_week_dates[mid] > job_posting_date:
                high = mid - 1
            else:
                low = mid + 1
        return False

    def getPastWeekChanges(self, current_date: datetime) -> list[datetime]:
        """
        Retrieve the commits from the past week


        Returns:
            - list[str]: The list of commits from the past week
        """
        past_week_dates = []
        for i in range(6, -1, -1):
            date = current_date - timedelta(days=i)
            past_week_dates.append(date)

        return past_week_dates

    def getCommitChanges(self, repo: github.Repository.Repository, readme_file: str) -> list[str]:
        """
        Retrieve the commit changes that make additions to the .md files

        Parameters:
            - repo: The GitHub repository
        Returns:
            - list[str]: The list of commit changes in the .md files
        """
        last_commit_sha = self.getLastCommit(repo)
        if last_commit_sha == "":
            return []

        commit = repo.get_commit(sha="8e4960ba59491b379db1980d52cd49ab029afd6f")#sha=last_commit_sha)
        previous_commit = commit.parents[0].sha  # Get the last commit before the current one
        comparison = repo.compare(previous_commit, commit.sha)

        commit_changes = []
        for file in comparison.files:
            if file.filename == readme_file:
                commit_lines = file.patch.split("\n") if file.patch else []
                for line in commit_lines:
                    # Check if the line is an addition and not a file header or subtraction
                    if (
                        line.startswith("+")
                        and not line.startswith("+++")
                        and "🔒" not in line
                    ):
                        commit_changes.append(line)

        return commit_changes
