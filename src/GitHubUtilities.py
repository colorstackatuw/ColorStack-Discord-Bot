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
from collections.abc import Iterable
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

        data_json["last_saved_sha"] = commit

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

    def getSavedSha(self, repo: github.Repository.Repository) -> str:
        """
        Retrieve the last commit information from the saved file

        Parameters:
            - repo: The GitHub repository
        Returns:
            - str: The last commit hexadecimal information
        """
        with self.FILEPATH.open("r") as file:
            commit_sha = json.load(file)["last_saved_sha"]

        if not commit_sha:
            # If the file is empty, get the previous commit from the repository
            recent_commit_sha = self.getLastCommit(repo)
            previous_commit = repo.get_commit(sha=recent_commit_sha)
            return previous_commit.parents[0].sha
        else:
            return commit_sha

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

    def getCommitChanges(
        self, repo: github.Repository.Repository, readme_file: str
    ) -> Iterable[str]:
        """
        Retrieve the commit changes that make additions to the .md files

        Parameters:
            - repo: The GitHub repository
            - readme_file: The name of the .md file
        Returns:
            - Iterable[str]: The lines that contain the job postings
        """
        recent_commit = self.getLastCommit(repo)
        if not recent_commit:
            return []

        previous_commit = self.getSavedSha(repo)  # Get the saved commit
        comparison = repo.compare(base=previous_commit, head=recent_commit)

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
                        yield line
