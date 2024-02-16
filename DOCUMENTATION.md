# ColorStack-Discord-Bot Documentation

## Table of Contents

- [Installation](#installation)
- [Daily Usage](#daily-usage)
- [Classes](#classes)
  - [DiscordBot](#discordbot)
  - [GitHubUtilities](#githubutilities)
  - [InternshipUtilities](#internshiputilities)

## Installation

Please refer to the [installation guide](https://github.com/colorstackatuw/ColorStack-Discord-Bot/blob/main/INSTALLATION.md)

## Daily Usage

While the bot is running, it will review the GitHub repositories and post any new opportunities in the Discord server the minute they are released. Here is the daily workflow:

1. The bot runs every mintue within `DiscordBot.py` and checks for new opportunities.
1. If a new opportunity is found, the bot will process the opportunity string using `getInternships()` found in `InternshipUtilites.py` and verify:
   1. It's in the United States or Remote
   1. The job posting is from the past 7 days
   1. The job posting is not a duplicate of a co-op internship
1. Once the post is validated, it will be posted in the Discord server.
1. After all the processing is done, the bot will save the commit SHA in `commits/repository_links_commits.json`, sleep for 60 seconds, and repeat the process.

## Classes

There are three main Python classes that allow the bot to function properly.

## DiscordBot

This class allows the bot to scrape the GitHub repositories and post the opportunities in the Discord server every 60 seconds.

### scheduled_task

A scheduled task that runs every 60 seconds to check for new commits in the GitHub repository.

| Parameter              | Description                                                                                                                                                  |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `github_utilities`     | An instance of the `GitHubUtilities` class, enabling the bot to connect to the GitHub API and scrape GitHub repositories                                     |
| `internship_utilities` | An instance of the `InternshipUtilities` class, allowing the bot to scrape GitHub repositories and post opportunities to the Discord server every 60 seconds |

### before_scheduled_task

Wait until the bot is ready before starting the loop.

### on_ready

Event that is triggered when the bot is ready to start sending messages.

## GitHubUtilities

This class allows the bot to scrape the GitHub repositories and post the opportunities in the Discord server every 60 seconds. This script provides a set of utilities to interact with GitHub repositories using the PyGithub library. It includes functionalities to establish a connection to a specified GitHub repository, update and retrieve the last commit information, and check for new commits.

### GitHubUtilities Constructor

This method initializes the GitHubUtilities class with the specified GitHub repository and the GitHub API token.

| Parameter   | Description                                           |
| ----------- | ----------------------------------------------------- |
| `token`     | A GitHub API Token to pass into the GitHub library    |
| `repo_name` | Name of the repository to collect the internship jobs |

### createGitHubConnection

Create a connection to the specified GitHub repository.

### setNewCommit

Save the last commit information to prevent duplicate job postings.

| Parameter     | Description                                                            |
| ------------- | ---------------------------------------------------------------------- |
| `last_commit` | The last saved commit SHA from `commits/repository_links_commits.json` |

### getLastCommit

Retrieve the last commit information based on the repository.

| Parameter | Description           |
| --------- | --------------------- |
| `repo`    | The GitHub repository |

### getSavedSha

Retrieve the last commit information from the saved file.

| Parameter | Description           |
| --------- | --------------------- |
| `repo`    | The GitHub repository |

### setComparison

Set the comparison between the previous commit and the recent commit.

| Parameter | Description           |
| --------- | --------------------- |
| `repo`    | The GitHub repository |

### clearComparison

Clear up the comparison between the previous commit and the recent commit.

### isNewCommit

Determine if there is a new commit on the GitHub repository.

| Parameter     | Description                                                            |
| ------------- | ---------------------------------------------------------------------- |
| `repo`        | The GitHub repository                                                  |
| `last_commit` | The last saved commit SHA from `commits/repository_links_commits.json` |

### getCommitChanges

Retrieve the commit changes that make additions to the Markdown files.

## InternshipUtilities

This class scrapes the GitHub repositories, processes the opportunities, and posts the opportunities in the Discord server every 60 seconds.

### InternshipUtilities Constructor

This method initializes the InternshipUtilities class with the specified GitHub repository and the Discord bot.

| Parameter | Description                             |
| --------- | --------------------------------------- |
| `summer`  | True, if looking for summer internships |
| `coop`    | True, if looking for coop internships   |

### clearJobLinks

Clear the Co-Op dictionary links.

### clearJobCounter

Clear the job counter.

### isWithinDateRange

Determine if the job posting is within the past week.

| Parameter      | Description                 |
| -------------- | --------------------------- |
| `job_date`     | The date of the job posting |
| `current_date` | The current date            |

### saveCompanyName

Save the previous job title into the class variable.

| Parameter      | Description      |
| -------------- | ---------------- |
| `company_name` | The company name |

### getInternships

Retrieve the Summer or Co-op internships from the GitHub repository.

| Parameter      | Description                                                   |
| -------------- | ------------------------------------------------------------- |
| `channel`      | The Discord channel to send the job postings                  |
| `job_postings` | The list of job postings                                      |
| `current_date` | The current date                                              |
| `is_summer`    | A boolean to record a job if it's summer or co-op internships |