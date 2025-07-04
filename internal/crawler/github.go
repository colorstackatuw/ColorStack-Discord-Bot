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
	"ColorStack-Discord-Bot/internal/types"
	jobTypes "ColorStack-Discord-Bot/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/google/go-github/v59/github"
	"github.com/pkg/errors"
)

type ReadMeInfo struct {
	ReadMeName string
	RepoName   string
}

type GitHubUtilities struct {
	RepoName   string
	GitHub     *github.Client
	JobType    *types.JobType
	Comparison *github.CommitsComparison
	SavedSHA   string
}

var FILEPATH = "crawlers/repository_links_commits.json"
var NOTUS [4]string = [4]string{"canada", "uk", "united kingdom", "eu"}
var readMeFile = map[types.JobType]ReadMeInfo{
	jobTypes.NEWGRAD:    {ReadMeName: "README.md", RepoName: "New-Grad-Positions"},
	jobTypes.INTERNSHIP: {ReadMeName: "README.md", RepoName: "Summer2025-Internships"},
	jobTypes.COOP:       {ReadMeName: "README-Off-Season.md", RepoName: "Summer2025-Internships"},
}

func init() {
	// Loads the .env fies
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

/*
NewGitHubUtilities creates and returns a new instance of GitHubUtilities.

Parameters:
- token: A string representing the GitHub access token for authentication.
- RepoName: A string specifying the name of the GitHub repository to interact with.
Returns: A pointer to an instance of GitHubUtilities.
*/
func NewGitHubUtilities(jobType types.JobType) *GitHubUtilities {
	gitHubToken := os.Getenv("GIT_TOKEN")
	if gitHubToken == "" {
		log.Fatal("No Git Token passed", errors.New("Empty Git Token"))
	}

	client := github.NewClient(nil).WithAuthToken(gitHubToken)
	return &GitHubUtilities{
		RepoName: readMeFile[jobType].RepoName,
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
SetComparison sets the Comparison field of the GitHubUtilities struct by comparing the most recent commit SHA with the previously saved SHA.

Parameters:
- ctx: A context.Context object for managing cancellations and timeouts.
- repo: A pointer to a GitHub.Repository object representing the GitHub repository.
- isNewGrad: True if repo is for new grad

Returns:

	An error if the Comparison fails, nil otherwise.
*/
func (g *GitHubUtilities) setComparison(
	ctx context.Context,
	repo *github.Repository,
	isNewGrad bool,
) error {
	log.Info("Retrieveing the last commit...")
	recentCommitSha, err := g.GetLastCommit(ctx, repo)
	if err != nil {
		g.Comparison = nil
		return errors.Wrap(err, "Can't find last commit")
	}

	prevCommit, err := g.getSavedSha(ctx, repo, isNewGrad)
	if err != nil {
		return errors.Wrap(err, "Couldn't get the saved SHA")
	}

	Comparison, _, err := g.GitHub.Repositories.CompareCommits(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		prevCommit,
		recentCommitSha,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Can't make the commit Comparisons")
	}

	g.Comparison = Comparison
	return nil
}

/* ClearComparison clears the Comparison field of the GitHubUtilities struct. */
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
func (g *GitHubUtilities) IsNewCommit(
	ctx context.Context,
	repo *github.Repository,
	SavedSHA string,
) (bool, error) {
	lastCommit, err := g.GetLastCommit(ctx, repo)
	if err != nil {
		return false, errors.Wrap(err, "Couldn't get the last commit")
	}

	return SavedSHA != lastCommit, nil
}

func (g *GitHubUtilities) GetJobs(jobType types.JobType, jobsChannel chan<- string) {
	var readmeFile string = readMeFile[jobType].ReadMeName
	initial := ""
	var prevJobTitle *string = &initial

	go func() {

		if g.Comparison == nil {
			return
		}

		for _, file := range g.Comparison.Files {
			if file.GetFilename() == readmeFile {
				var commitAdditions string = file.GetPatch()
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
						jobPost := g.parseJobPosting(line, prevJobTitle, jobType)
						jobsChannel <- jobPost
					}
				}
				break
			}
		}
	}()
}

func (g *GitHubUtilities) parseJobPosting(
	jobLine string,
	prevJobTitle *string,
	season types.JobType,
) string {
	var companyName, jobTitle, jobLink, terms, location string
	var listLocations []string
	var jobLinkIndex int = 4
	if season == jobTypes.COOP {
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
		companyName = *prevJobTitle
	}
	*prevJobTitle = companyName

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
	if season == jobTypes.INTERNSHIP {
		terms = "Summer " + strconv.Itoa(currentYear)
	} else if season == jobTypes.COOP {
		terms = strings.Join(strings.Split(nonEmptyElements[4], ","), " |")
	}

	jobTitle = nonEmptyElements[2]
	var post strings.Builder
	post.WriteString(fmt.Sprintf(">>> ## [%s @ %s](<%s>)\n", jobTitle, companyName, jobLink))
	post.WriteString(fmt.Sprintf("### Locations: \n%s\n", location))
	if season != jobTypes.NEWGRAD {
		post.WriteString(fmt.Sprintf("When?: %s\n", terms))
	}

	// Update the Redis Database
	if err := redisClient.WriteURL(jobLink); err != nil {
		log.Error("Failed to update Redis DB", err)
	}

	return post.String()
}
