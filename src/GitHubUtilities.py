from github import Auth, Github
import github
import json

class GitHubUtilities:

    def __init__(self, token, repo_name = "SimplifyJobs/Summer2024-Internships"): 
        self.repo_name = repo_name 
        self.github = Github(auth=Auth.Token(token)) # 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

    def createGitHubConnection(self):
        return self.github.get_repo(self.repo_name)
    
    def setNewCommit(self, commit: str):
        with open("./commits/repository_links_commits.json", "r") as file:
            data_json = json.load(file)

        data_json["last_commit"] = commit

        with open("./commits/repository_links_commits.json", "w") as file:
            json.dump(data_json, file)

    def getLastCommit(self, repo: github.Repository.Repository) -> str:
        branch = repo.get_branches()[0] # May need to be changed in future
        return branch.commit.sha
    
    def getCommitLinks(self) -> str:
        with open("./commits/repository_links_commits.json") as file:
            return json.load(file)["last_commit"]

    def isNewCommit(self, repo :github.Repository.Repository, last_commit: str) -> bool:
        return last_commit != self.getLastCommit(repo)
