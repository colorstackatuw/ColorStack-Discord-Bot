/*
GitHub Utilities Class

This class provides a set of utilities to interact with GitHub repositories using the PyGithub library.
It includes functionalities to establish a connection to a specified GitHub repository, update and retrieve
the last commit information, and check for new commits.
*/
package Utilities

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/google/go-github/v59/github"
)

const FILEPATH = "repository_links_commits.json"

type GitHubUtilities struct {
	RepoName   string
	GitHub     *github.Client
	Comparison *github.CommitsComparison
	SavedSHA   string
}

/*
NewGitHubUtilities creates and returns a new instance of GitHubUtilities.

Parameters:
- token: A string representing the GitHub access token for authentication.
- RepoName: A string specifying the name of the GitHub repository to interact with.
Returns: A pointer to an instance of GitHubUtilities.
*/
func NewGitHubUtilities(token, repoName string) *GitHubUtilities {
	client := github.NewClient(nil).WithAuthToken(token)

	return &GitHubUtilities{
		RepoName: repoName,
		GitHub:   client,
	}

}

/*
SetNewCommit saves the SHA of the latest commit to a JSON file.

Parameters:
- lastCommit: A string representing the SHA of the last commit to be saved.
Returns: An error if saving fails, nil otherwise.
*/
func (g *GitHubUtilities) SetNewCommit(lastCommit string) error {
	dataJson := make(map[string]string)

	dataJson["last_saved_sha"] = lastCommit

	data, err := json.Marshal(dataJson)
	if err != nil {
		return err
	}

	err = os.WriteFile(FILEPATH, data, 0644)
	return err
}

/*
SetSavedSha reads and returns the last saved commit SHA from a JSON file.

Parameters: None.
Returns:
- An error if reading the file or unmarshalling JSON fails, nil otherwise.
*/
func (g *GitHubUtilities) SetSavedSha() error {
	data, err := os.ReadFile(FILEPATH)

	if err != nil {
		return "", err
	}

	var dataJson map[string]string
	err = json.Unmarshal(data, &dataJson)
	if err != nil {
		return "", err
	}

	g.SavedSHA = dataJson["last_saved_sha"]
	return nil
}

/*
CreateGitHubConnection establishes a connection to the specified GitHub repository and returns the repository object.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
Returns:
- A pointer to a GitHub.Repository representing the specified repository.
- An error if the connection or retrieval fails, nil otherwise.
*/
func (g *GitHubUtilities) CreateGitHubConnection(ctx context.Context) (*github.Repository, error) {
	repo, _, err := g.GitHub.Repositories.Get(ctx, "SimplifyJobs", g.RepoName)
	return repo, err
}

/*
GetLastCommit retrieves the SHA of the latest commit from the default branch of the specified repository.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- repo: A pointer to a GitHub.Repository object representing the GitHub repository.
Returns:
- A string representing the SHA of the latest commit.
- An error if retrieving the commit fails, nil otherwise.
*/
func (g *GitHubUtilities) GetLastCommit(
	ctx context.Context,
	repo *github.Repository,
) (string, error) {
	branches, _, err := g.GitHub.Repositories.ListBranches(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		nil,
	)
	if err != nil {
		return "", err
	}

	branchName := branches[0].GetName()
	mainBranch, _, err := g.GitHub.Repositories.GetBranch(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		branchName,
		0,
	)
	if err != nil {
		return "", err
	}

	return mainBranch.Commit.GetSHA(), nil
}

/*
SetComparison sets the comparison field of the GitHubUtilities struct by comparing the most recent commit SHA with the previously saved SHA.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- repo: A pointer to a GitHub.Repository object representing the GitHub repository.
Returns: An error if the comparison fails, nil otherwise.
*/
func (g *GitHubUtilities) SetComparison(
	ctx context.Context,
	repo *github.Repository,
) error {
	recentCommitSha, err := g.GetLastCommit(ctx, repo)
	if err != nil {
		return err
	}

	comparison, _, err := g.GitHub.Repositories.CompareCommits(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		g.SavedSHA,
		recentCommitSha,
		nil,
	)
	if err != nil {
		return err
	}

	g.Comparison = comparison
	return nil

}

/* ClearComparison clears the comparison field of the GitHubUtilities struct. */
func (g *GitHubUtilities) ClearComparison() {
	g.Comparison = nil
}

/*
IsNewCommit checks if the given commit SHA is different from the last saved commit SHA.

Parameters:
- lastCommit: A string representing the SHA of the commit to check.
Returns:
- A boolean indicating whether the given commit SHA is new (true) or not (false).
- An error if retrieving the saved commit SHA fails, nil otherwise.
*/
func (g *GitHubUtilities) IsNewCommit(lastCommitSHA string) bool {
	if err := g.SetSavedSha(); err != nil {
		return err
	}
	return lastCommitSHA != g.SavedSHA
}

func (g *GitHubUtilities) GetCommitChanges(readmeFile string) <-chan string {
	channel := make(chan string)

	go func() {
		defer close(channel)

		if g.Comparison == nil {
			return
		}

		for _, file := range g.Comparison.Files {
			if file.GetFilename() == readmeFile {
				commitAdditions := file.GetPatch()
				if commitAdditions == "" {
					continue
				}
				commitLines := strings.Split(commitAdditions, "\n")
				for _, line := range commitLines {
					if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") &&
						!strings.Contains(line, "🔒") {
						channel <- line
					}
				}
				break
			}
		}
	}()

	return channel
}
