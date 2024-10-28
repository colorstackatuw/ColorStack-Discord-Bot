package utilities

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type JobUtilities struct {
	PreviousJobTitle string
	JobCache         map[string]struct{}
	TotalJobs        int
	NotUS            [4]string
}

/*
NewJobUtilities creates and returns a new instance of JobUtilities.
It initializes the JobCache and sets the IsSummer flag based on the provided summer parameter.

Parameters:
- summer: Determines if the utility is for summer internships.
Returns: A pointer to the newly created JobUtilities instance.
*/
func NewJobUtilities() *JobUtilities {
	return &JobUtilities{
		JobCache: make(map[string]struct{}),
		NotUS:    [4]string{"canada", "uk", "united kingdom", "eu"},
	}
}

/*
ClearJobLinks resets the JobCache map to an empty state.

This method does not take parameters or return any value.
*/
func (iu *JobUtilities) ClearJobLinks() {
	iu.JobCache = make(map[string]struct{})
}

/*
ClearJobCounter sets the TotalJobs counter to 0.

This method does not take parameters or return any value.
*/
func (iu *JobUtilities) ClearJobCounter() {
	iu.TotalJobs = 0
}

/*
SaveCompanyName sets the PreviousJobTitle field to the provided companyName.

Parameters:
- companyName: The name of the company to save as the previous job title.
*/
func (iu *JobUtilities) SaveCompanyName(companyName string) {
	iu.PreviousJobTitle = companyName
}

/*
IsNotUS checks if the given location is not within the US based on predefined locations in NotUS.

Parameters:
- location: The location to check.
Returns: True if the location is not within the US, false otherwise.
*/
func (iu *JobUtilities) IsNotUS(location string) bool {
	lowerLocation := strings.ToLower(location)
	for _, notUS := range iu.NotUS {
		if strings.Contains(lowerLocation, notUS) {
			return true
		}
	}

	return false
}

/*
GetJobs processes job postings, filters them based on certain criteria, and sends them to the specified channels.
It operates asynchronously and sends the formatted job postings through a jobPostingsChan channel.

Parameters:
- channels: A slice of strings representing the channels to send job postings to.
- jobPostings: A slice of strings representing the job postings to process.
- isSummer: A flag indicating whether to process summer internships.
Returns: A channel through which processed and formatted job postings are sent.
*/
func (iu *JobUtilities) GetJobs(
	ctx context.Context,
	discordBot *discordgo.Session,
	channels []string,
	jobPostingChannel <-chan string,
	term string,
	redisClient *redis.Client,
) error {
	// Check the term is accurate
	var isValid bool = false
	for _, types := range [3]string{"Sumemr", "Co-Op", "New Grad"} {
		if types == term {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("Term must be one of these: Summer, Coop, NewGrad")
	}

	// Determine the index of the job link
	currentYear := time.Now().Year()
	var jobLinkIndex int = 4
	var hasPrinted bool = false
	if term == "Co-Op" {
		jobLinkIndex = 5
	}

	for job := range jobPostingChannel {
		var companyName, jobTitle, jobLink, terms, location string
		var listLocations []string

		// Grab data and remove the empty elements
		elements := strings.Split(job, "|")
		nonEmptyElements := make([]string, 0)
		for _, element := range elements {
			if strings.TrimSpace(element) != "" {
				nonEmptyElements = append(nonEmptyElements, strings.TrimSpace(element))
			}
		}

		// If job link is already in cache or redis db,  we skip the job
		re, err := regexp.Compile(`href="([^"]+)"`)
		if err != nil {
			log.Print("There is no job link within the job posting")
			continue
		}
		matches := re.FindStringSubmatch(nonEmptyElements[jobLinkIndex])
		if len(matches) < 2 {
			continue
		}
		jobLink = matches[1]
		if _, exists := iu.JobCache[jobLink]; exists {
			continue
		}

		if _, err := redisClient.Get(ctx, jobLink).Result(); err != nil {
			log.Printf("It already exists within the database %v", jobLink)
			continue
		}
		iu.JobCache[jobLink] = struct{}{}

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
			companyName = iu.PreviousJobTitle
		}
		iu.SaveCompanyName(companyName)

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
					if location != "" && !iu.IsNotUS(location) {
						listLocations = append(listLocations, location)
					}
				}
			}
		} else if strings.Contains(locationHTML, "</br>") {
			locations := strings.Split(locationHTML, "</br>")
			for _, location := range locations {
				location = strings.TrimSpace(location)
				if location != "" && !iu.IsNotUS(location) {
					listLocations = append(listLocations, location)
				}
			}
		} else if locationHTML != "" {
			var location string = "Remote"
			if !strings.Contains(strings.ToLower(locationHTML), "remote") && !iu.IsNotUS(locationHTML) {
				location = locationHTML
			}
			listLocations = append(listLocations, location)
		}

		if len(listLocations) >= 1 {
			location = strings.Join(listLocations, " | ")
		} else {
			continue
		}

		if term == "Summer" {
			terms = "Summer " + strconv.Itoa(currentYear)
		} else if term == "Co-Op" {
			terms = strings.Join(strings.Split(nonEmptyElements[4], ","), " |")
		}

		jobTitle = nonEmptyElements[2]
		var post strings.Builder
		datePosted := nonEmptyElements[len(nonEmptyElements)-1]
		if !hasPrinted {
			post.WriteString(fmt.Sprintf("# %s Postings!\n\n", term))
			hasPrinted = true
		}
		post.WriteString(fmt.Sprintf("**📅 Date Posted:** %s\n", datePosted))
		post.WriteString(fmt.Sprintf("**ℹ️ Company:** __%s__\n", companyName))
		post.WriteString(fmt.Sprintf("**👨‍💻 Job Title:** %s\n", jobTitle))
		post.WriteString(fmt.Sprintf("**📍 Location:** %s\n", location))
		if term != "New Grad" {
			post.WriteString(fmt.Sprintf("**➡️  When?:**  %s\n", terms))
		}
		post.WriteString(
			fmt.Sprintf("**👉 Job Link:** <%s>\n%s\n", jobLink, strings.Repeat("-", 153)),
		)
		iu.TotalJobs++

		// Update the Redis Database
		if err := redisClient.Set(ctx, jobLink, "", 0).Err(); err != nil {
			return errors.Wrap(err, "Cannot update the Redis DB")
		}

		// Work on concurrent posts
		wg := sync.WaitGroup{}
		for _, channel := range channels {
			wg.Add(1)
			go func(ch string) {
				defer wg.Done()
				if _, err := discordBot.ChannelMessageSend(ch, post.String()); err != nil {
					log.Panic(err)
				}
			}(channel)
		}
		wg.Wait()
	}

	return nil
}
