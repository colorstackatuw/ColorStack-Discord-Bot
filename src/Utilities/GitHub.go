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
	"github.com/pkg/errors"
)

const FILEPATH = "src/repository_links_commits.json"

type GitHubUtilities struct {
	RepoName   string
	GitHub     *github.Client
	Comparison *github.CommitsComparison
	SavedSHA   string
	IsCoop     bool
	IsSummer   bool
}

/*
NewGitHubUtilities creates and returns a new instance of GitHubUtilities.

Parameters:
- token: A string representing the GitHub access token for authentication.
- RepoName: A string specifying the name of the GitHub repository to interact with.
Returns: A pointer to an instance of GitHubUtilities.
*/
func NewGitHubUtilities(token, repoName string, isSummer bool, isCoop bool) *GitHubUtilities {
	client := github.NewClient(nil).WithAuthToken(token)

	return &GitHubUtilities{
		RepoName: repoName,
		GitHub:   client,
		IsCoop:   isCoop,
		IsSummer: isSummer,
	}

}

/*
SetNewCommit saves the SHA of the latest commit to a JSON file.

Parameters:
- lastCommit: A string representing the SHA of the last commit to be saved.
- isNewGrad: True if commit is for repo

Returns: An error if saving fails, nil otherwise.
*/
func (g *GitHubUtilities) SetNewCommit(lastCommit string, isNewGrad bool) error {
	var key string
	if isNewGrad {
		key = "last_saved_sha_newgrad"
	} else {
		key = "last_saved_sha_internship"
	}

	dataJson := make(map[string]string)
	dataJson[key] = lastCommit

	data, err := json.Marshal(dataJson)
	if err != nil {
		return errors.Wrap(err, "Couldn't parse data")
	}

	err = os.WriteFile(FILEPATH, data, 0644)
	if err != nil {
		return errors.Wrap(err, "Couldn't write to file!")
	}

	return nil
}

/*
SetSavedSha reads and returns the last saved commit SHA from a JSON file.

Parameters: None.

Returns:
- An error if reading the file or unmarshalling JSON fails, nil otherwise.
*/
func (g *GitHubUtilities) SetSavedSha(isNewGrad bool) error {
	var key string
	if isNewGrad {
		key = "last_saved_sha_newgrad"
	} else {
		key = "last_saved_sha_internship"
	}

	data, err := os.ReadFile(FILEPATH)
	if err != nil {
		return errors.Wrap(err, "Can't Read file")
	}

	var dataJson map[string]string
	err = json.Unmarshal(data, &dataJson)
	if err != nil {
		return errors.Wrap(err, "Can't unwrap data")
	}

	g.SavedSHA = dataJson[key]
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
	return repo, errors.Wrap(err, "Failed to connect")
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
	var branchName string = "dev"
	mainBranch, _, err := g.GitHub.Repositories.GetBranch(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		branchName,
		0,
	)
	if err != nil {
		return "", errors.Wrap(err, "Could not retrieve the main branch")
	}

	return mainBranch.Commit.GetSHA(), nil
}

/*
Retrieve the last commit information from the saved file

Parameters:
  - repo: The GitHub repository
  - isNewGrad: True if getting new grad sha

Returns:
  - str: The last commit hexadecimal information
*/
func (g *GitHubUtilities) GetSavedSha(ctx context.Context, repo *github.Repository, isNewGrad bool) (string, error) {
	// Determine the key based on isNewGrad
	var key string = "last_saved_sha_newgrad"
	if !isNewGrad {
		key = "last_saved_sha_internship"
	}

	data, err := os.ReadFile(FILEPATH)
	if err != nil {
		return "", errors.Wrap(err, "Can't read file")
	}
	var dataJSON map[string]string
	err = json.Unmarshal(data, &dataJSON)
	if err != nil {
		return "", errors.Wrap(err, "Can't unwrap data")
	}

	// If the file is empty, get the previous commit from the repository
	if dataJSON[key] != "" {
		return dataJSON[key], nil
	} else {
		recentCommitSHA, err := g.GetLastCommit(ctx, repo)
		if err != nil {
			return "", errors.Wrap(err, "Can't get the last commit") 
		}
		previousCommit, _, err := g.GitHub.Repositories.GetCommit(ctx, repo.GetOwner().GetLogin(), repo.GetName(), recentCommitSHA, nil)
		if err != nil {
			return "", errors.Wrap(err, "Can't access the previous commit") 
		}

		return *previousCommit.Parents[0].SHA, nil
	}
}

/*
SetComparison sets the comparison field of the GitHubUtilities struct by comparing the most recent commit SHA with the previously saved SHA.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- repo: A pointer to a GitHub.Repository object representing the GitHub repository.
- isNewGrad: True if repo is for new grad

Returns: 
	An error if the comparison fails, nil otherwise.
*/
func (g *GitHubUtilities) SetComparison(
	ctx context.Context,
	repo *github.Repository,
	isNewGrad bool,
) error {
	recentCommitSha, err := g.GetLastCommit(ctx, repo)
	if err != nil {
		g.Comparison = nil
		return errors.Wrap(err, "Can't find last commit")
	}

	prevCommit, err := g.GetSavedSha(ctx, repo, isNewGrad)
	if err != nil{
		return errors.Wrap(err, "Couldn't get the saved SHA")
	}

	comparison, _, err := g.GitHub.Repositories.CompareCommits(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		prevCommit,
		recentCommitSha,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Can't make the commit comparisons")
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
func (g *GitHubUtilities) IsNewCommit(ctx context.Context, repo *github.Repository, savedSHA string) (bool, error) {
	lastCommit, err:= g.GetLastCommit(ctx, repo)	
	if err != nil{
		return false, errors.Wrap(err, "Couldn't get the last commit")
	}

	return savedSHA != lastCommit, nil 
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
