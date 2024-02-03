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
