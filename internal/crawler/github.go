/*
GitHub Utilities Class

This class provides a set of utilities to interact with GitHub repositories using the PyGithub library.
It includes functionalities to establish a connection to a specified GitHub repository, update and retrieve
the last commit information, and check for new commits.
*/
package crawler

import (
	"ColorStack-Discord-Bot/internal/database"
	log "ColorStack-Discord-Bot/internal/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v59/github"
	"github.com/pkg/errors"
)

var FILEPATH = "crawlers/repository_links_commits.json"
var NOTUS [4]string = [4]string{"canada", "uk", "united kingdom", "eu"}
var readMeFile = map[JobType]string{
	NEWGRAD: "New Grad",
	SUMMER:  "Summer",
	COOP:    "Co-Op",
}

type GitHubUtilities struct {
	RepoName   string
	GitHub     *github.Client
	JobType    *JobType
	comparison *github.CommitsComparison
	savedSHA   string
}

/*
NewGitHubUtilities creates and returns a new instance of GitHubUtilities.

Parameters:
- token: A string representing the GitHub access token for authentication.
- RepoName: A string specifying the name of the GitHub repository to interact with.
Returns: A pointer to an instance of GitHubUtilities.
*/
func NewGitHubUtilities(token, repoName string, jobType JobType) *GitHubUtilities {
	client := github.NewClient(nil).WithAuthToken(token)

	return &GitHubUtilities{
		RepoName: repoName,
		GitHub:   client,
		JobType:  &jobType,
	}
}

/*
SetNewCommit saves the SHA of the latest commit to a JSON file.

Parameters:
- lastCommit: A string representing the SHA of the last commit to be saved.
- isNewGrad: True if commit is for repo

Returns: An error if saving fails, nil otherwise.
*/
func (g *GitHubUtilities) setNewCommit(lastCommit string, isNewGrad bool) error {
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

	msg := fmt.Sprintf("Writing %s to file", lastCommit)
	log.Debug(msg)

	err = os.WriteFile(FILEPATH, data, 0644)
	if err != nil {
		return errors.Wrap(err, "Couldn't write to file!")
	}

	return nil
}

func (g *GitHubUtilities) IsNotUS(location string) bool {
	lowerLocation := strings.ToLower(location)
	for _, notUS := range NOTUS {
		if strings.Contains(lowerLocation, notUS) {
			return true
		}
	}

	return false
}

/*
SetSavedSha reads and returns the last saved commit SHA from a JSON file.

Parameters: None.

Returns:
- An error if reading the file or unmarshalling JSON fails, nil otherwise.
*/
func (g *GitHubUtilities) SetSavedSha(isNewGrad bool) error {
	log.Info("Retrieving saved SHA commit from files...")
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

	msg := fmt.Sprintf("Collected SHA: %s", dataJson[key])
	log.Debug(msg)
	g.savedSHA = dataJson[key]

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
	log.Debug("Making connection to github...")
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
	msg := fmt.Sprintf("dev branch commit sha: %s", mainBranch.Commit.GetSHA())
	log.Debug(msg)

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
func (g *GitHubUtilities) getSavedSha(
	ctx context.Context,
	repo *github.Repository,
	isNewGrad bool,
) (string, error) {
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
		log.Info("File is empty! Getting previous commit from repo")
		return dataJSON[key], nil
	} else {
		recentCommitSHA, err := g.GetLastCommit(ctx, repo)
		if err != nil {
			return "", errors.Wrap(err, "Can't get the last commit")
		}
		msg := fmt.Sprintf("Recent Commit SHA: %s", recentCommitSHA)
		log.Info(msg)

		previousCommit, _, err := g.GitHub.Repositories.GetCommit(ctx, repo.GetOwner().GetLogin(), repo.GetName(), recentCommitSHA, nil)
		if err != nil {
			return "", errors.Wrap(err, "Can't access the previous commit")
		}

		msg = fmt.Sprintf("Previous Commit SHA: %s", previousCommit)
		log.Debug(msg)

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
func (g *GitHubUtilities) setComparison(
	ctx context.Context,
	repo *github.Repository,
	isNewGrad bool,
) error {
	log.Info("Retrieveing the last commit...")
	recentCommitSha, err := g.GetLastCommit(ctx, repo)
	if err != nil {
		g.comparison = nil
		return errors.Wrap(err, "Can't find last commit")
	}

	prevCommit, err := g.getSavedSha(ctx, repo, isNewGrad)
	if err != nil {
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

	g.comparison = comparison
	return nil
}

/* ClearComparison clears the comparison field of the GitHubUtilities struct. */
func (g *GitHubUtilities) clearComparison() {
	g.comparison = nil
}

/*
IsNewCommit checks if the given commit SHA is different from the last saved commit SHA.

Parameters:
- lastCommit: A string representing the SHA of the commit to check.
Returns:
- A boolean indicating whether the given commit SHA is new (true) or not (false).
- An error if retrieving the saved commit SHA fails, nil otherwise.
*/
func (g *GitHubUtilities) isNewCommit(
	ctx context.Context,
	repo *github.Repository,
	savedSHA string,
) (bool, error) {
	lastCommit, err := g.GetLastCommit(ctx, repo)
	if err != nil {
		return false, errors.Wrap(err, "Couldn't get the last commit")
	}

	return savedSHA != lastCommit, nil
}

func (g *GitHubUtilities) GetJobs(jobType JobType, jobsChannel <-chan JobType) {
	var readmeFile string = readMeFile[jobType]

	go func() {

		if g.comparison == nil {
			return
		}

		for _, file := range g.comparison.Files {
			if file.GetFilename() == readmeFile {
				commitAdditions := file.GetPatch()
				if commitAdditions == "" {
					continue
				}
				commitLines := strings.Split(commitAdditions, "\n")
				for _, line := range commitLines {
					if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") &&
						!strings.Contains(line, "🔒") {
						msg := fmt.Sprintf("Collected Github line: %s", line)
						log.Debug(msg)

						// Formatting Job
						g.parseJobPosting(line)
						channel <- line
					}
				}
				break
			}
		}
	}()

	return channel
}

func (g *GitHubUtilities) parseJobPosting(
	jobLine string,
	prevJobTitle string,
	season JobType,
) string {
	var companyName, jobTitle, jobLink, terms, location string
	var listLocations []string
	var jobLinkIndex int = 4
	if season == COOP {
		jobLinkIndex = 5
	}

	// Grab data and remove the empty elements
	elements := strings.Split(jobLine, "|")
	nonEmptyElements := make([]string, 0)
	for _, element := range elements {
		if strings.TrimSpace(element) != "" {
			nonEmptyElements = append(nonEmptyElements, strings.TrimSpace(element))
		}
	}

	// If job link is already in cache or redis db,  we skip the job
	re, err := regexp.Compile(`href="([^"]+)"`)
	if err != nil {
		log.Info("There is no job link within the job posting")
		return ""
	}
	matches := re.FindStringSubmatch(nonEmptyElements[jobLinkIndex])
	if len(matches) < 2 {
		log.Info("Could not find job link within github line")
	}
	jobLink = matches[1]

	redisClient := database.GetRedisInstance()
	if _, err := redisClient.Ping(); err != nil {
		log.Error("Cannot connect to the redis database", err)
		return ""
	} else {
		log.Info("We have connected to the Redis Database!")
	}
	defer redisClient.Close()

	urlExists, err := redisClient.CheckURL(jobLink)
	if urlExists {
		logMsg := fmt.Sprintf("It already exists within the database %v", urlExists)
		log.Info(logMsg)
		return ""
	}

	// If the company name is not present, we need to use the previous company name
	if !strings.Contains(nonEmptyElements[1], "↳") {
		jobHeader := nonEmptyElements[1]
		startPos := strings.Index(jobHeader, "[") + 1
		endPos := strings.Index(jobHeader[startPos:], "]") + startPos

		if startPos > 0 && endPos > 0 {
			companyName = jobHeader[startPos:endPos]
		} else {
			companyName = jobHeader
		}
	} else {
		companyName = prevJobTitle
	}
	prevJobTitle = companyName

	// We need to check that the position is within the US or remote
	locationHTML := nonEmptyElements[3]
	if strings.Contains(locationHTML, "<details>") {
		start := strings.Index(locationHTML, "</summary>") + len("</summary>")
		end := strings.Index(locationHTML, "</details>")
		if start >= 0 && end >= 0 {
			locationsContent := locationHTML[start:end]
			locations := strings.Split(locationsContent, "</br>")
			for _, location := range locations {
				location = strings.TrimSpace(location)
				if location != "" && !g.IsNotUS(location) {
					listLocations = append(listLocations, location)
				}
			}
		}
	} else if strings.Contains(locationHTML, "</br>") {
		locations := strings.Split(locationHTML, "</br>")
		for _, location := range locations {
			location = strings.TrimSpace(location)
			if location != "" && !g.IsNotUS(location) {
				listLocations = append(listLocations, location)
			}
		}
	} else if locationHTML != "" {
		var location string = "Remote"
		if !strings.Contains(strings.ToLower(locationHTML), "remote") && !g.IsNotUS(locationHTML) {
			location = locationHTML
		}
		listLocations = append(listLocations, location)
	}

	log.Debug("List of locations")
	if len(listLocations) >= 1 {
		location = strings.Join(listLocations, " | ")
	} else {
		log.Info("No locations found!")
		return ""
	}

	currentYear := time.Now().Year()
	if season == SUMMER {
		terms = "Summer " + strconv.Itoa(currentYear)
	} else if season == COOP {
		terms = strings.Join(strings.Split(nonEmptyElements[4], ","), " |")
	}

	jobTitle = nonEmptyElements[2]
	var post strings.Builder
	datePosted := nonEmptyElements[len(nonEmptyElements)-1]
	post.WriteString(fmt.Sprintf("**📅 Date Posted:** %s\n", datePosted))
	post.WriteString(fmt.Sprintf("**ℹ️ Company:** __%s__\n", companyName))
	post.WriteString(fmt.Sprintf("**👨‍💻 Job Title:** %s\n", jobTitle))
	post.WriteString(fmt.Sprintf("**📍 Location:** %s\n", location))
	if season != NEWGRAD {
		post.WriteString(fmt.Sprintf("**➡️  When?:**  %s\n", terms))
	}
	post.WriteString(
		fmt.Sprintf("**👉 Job Link:** <%s>\n%s\n", jobLink, strings.Repeat("-", 153)),
	)

	// Update the Redis Database
	if err := redisClient.WriteURL(jobLink); err != nil {
		log.Error("Failed to update Redis DB", err)
		return errors.Wrap(err, "Failed to update the Redis DB")
	}

	return post
}

/*
processJobs performs a periodic task to check for new GitHub commits and post new job opportunities.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
Returns: None.
*/
func (g *GitHubUtilities) GetJobPostings(ctx context.Context, jobType JobType) {
	// Open Connection
	log.Info("Connecting to github repos...")
	var isNewGrad bool = false
	repo, err := g.CreateGitHubConnection(ctx)
	if err != nil {
		log.Error("Failed to create GitHub connection for internship jobs", err)
		return
	}

	if jobType == NEWGRAD {
		isNewGrad = true
	}

	// Get commit SHA
	savedSHA, err := g.getSavedSha(ctx, repo, isNewGrad)
	if err != nil {
		log.Error("Failed to get internship SHA", err)
		return
	}

	// Collect any new internship or newgrad jobs
	isNewJobs, err := g.isNewCommit(ctx, repo, savedSHA)
	if err != nil {
		log.Error("Failed to get the new internship commit", err)
		return
	}

	if !isNewJobs {
		log.Info("No new jobs found!")
		return
	}

	// Set up Redis Database
	redisClient := database.GetRedisInstance()
	if _, err := redisClient.Ping(); err != nil {
		log.Error("Cannot connect to the redis database", err)
		return
	} else {
		log.Info("We have connected to the Redis Database!")
	}
	defer redisClient.Close()

	// Set up Oracle db
	oracleClient := database.GetDatabaseInstance()
	defer oracleClient.Close()

	log.Info("New commit has been found. Finding new jobs...")
	g.setComparison(ctx, repo, isNewGrad)

	channelIDs, err := oracleClient.GetChannels()
	if err != nil {
		log.Error("Failed to get channel IDs", err)
	}

	jobPostings := g.getCommitChanges(jobType)

	if err != nil {
		log.Error("Issue collecting jobs", err)
		return
	}

	// Save latest commit on repo
	sha_commit, err := g.getLastCommit(ctx, repo)
	if err != nil {
		log.Error("Failed to get the latest commit!", err)
	}

	if err := g.setNewCommit(sha_commit, false); err != nil {
		log.Error("Failed to set the new commit", err)
	}

	logMsg := fmt.Sprintf("New %d jobs found!")
	log.Info(logMsg)

}
