package Utilities

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type InternshipUtilities struct {
	IsSummer         bool
	PreviousJobTitle string
	JobCache         map[string]struct{}
	TotalJobs        int
	NotUS            [4]string
}

/*
NewInternshipUtilities creates and returns a new instance of InternshipUtilities.
It initializes the JobCache and sets the IsSummer flag based on the provided summer parameter.

Parameters:
- summer: Determines if the utility is for summer internships.
Returns: A pointer to the newly created InternshipUtilities instance.
*/
func NewInternshipUtilities(summer bool) *InternshipUtilities {
	return &InternshipUtilities{
		IsSummer: summer,
		JobCache: make(map[string]struct{}),
		NotUS:    [4]string{"canada", "uk", "united kingdom", "eu"},
	}
}

/*
ClearJobLinks resets the JobCache map to an empty state.

This method does not take parameters or return any value.
*/
func (iu *InternshipUtilities) ClearJobLinks() {
	iu.JobCache = make(map[string]struct{})
}

/*
ClearJobCounter sets the TotalJobs counter to 0.

This method does not take parameters or return any value.
*/
func (iu *InternshipUtilities) ClearJobCounter() {
	iu.TotalJobs = 0
}

/*
IsWithinDateRange checks if a given jobDate is within 3 days before the currentDate.

Parameters:
- jobDate: The date of the job posting.
- currentDate: The current date for comparison.
Returns: True if jobDate is within 3 days before currentDate, false otherwise.
*/
func (iu *InternshipUtilities) IsWithinDateRange(jobDate time.Time) bool {
	currentDate := time.Now().UTC()
	normalizedJobDate := time.Date(
		currentDate.Year(),
		jobDate.Month(),
		jobDate.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	normalizedCurrentDate := time.Date(
		currentDate.Year(),
		currentDate.Month(),
		currentDate.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	diff := normalizedCurrentDate.Sub(normalizedJobDate)
	days := int(diff.Hours() / 24)

	return 0 <= days && days <= 3
}

/*
SaveCompanyName sets the PreviousJobTitle field to the provided companyName.

Parameters:
- companyName: The name of the company to save as the previous job title.
*/
func (iu *InternshipUtilities) SaveCompanyName(companyName string) {
	iu.PreviousJobTitle = companyName
}

/*
IsNotUS checks if the given location is not within the US based on predefined locations in NotUS.

Parameters:
- location: The location to check.
Returns: True if the location is not within the US, false otherwise.
*/
func (iu *InternshipUtilities) IsNotUS(location string) bool {
	lowerLocation := strings.ToLower(location)
	for _, notUS := range iu.NotUS {
		if strings.Contains(lowerLocation, notUS) {
			return true
		}
	}

	return false
}

/*
GetInternships processes job postings, filters them based on certain criteria, and sends them to the specified channels.
It operates asynchronously and sends the formatted job postings through a jobPostingsChan channel.

Parameters:
- channels: A slice of strings representing the channels to send job postings to.
- jobPostings: A slice of strings representing the job postings to process.
- isSummer: A flag indicating whether to process summer internships.
Returns: A channel through which processed and formatted job postings are sent.
*/
func (iu *InternshipUtilities) GetInternships(
	discordBot *discordgo.Session,
	channels []string,
	jobPostingChannel <-chan string,
	isSummer bool,
) {

	// Determine the index of the job link
	currentYear := time.Now().Year()
	var jobLinkIndex int = 4
	if !isSummer {
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

		// If job link is already in cache, we skip the job
		re, err := regexp.Compile(`href="([^"]+)"`)
		if err != nil {
			fmt.Println("Regex compile error:", err) //! Add proper stacktrack error
			return
		}

		matches := re.FindStringSubmatch(nonEmptyElements[jobLinkIndex])
		if len(matches) < 2 {
			continue
		}

		jobLink = matches[1]
		if _, exists := iu.JobCache[jobLink]; exists {
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

		datePosted := nonEmptyElements[len(nonEmptyElements)-1]
		formatedDate := fmt.Sprintf("%s %d", datePosted, currentYear)
		layout := "Jan 02 2006"
		jobDate, err := time.Parse(layout, formatedDate)
		if err != nil {
			panic(err) // ! Add proper stack trace
		}

		if !iu.IsWithinDateRange(jobDate) {
			continue
		}

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

		if isSummer {
			terms = "Summer " + strconv.Itoa(currentYear)
		} else {
			terms = strings.Join(strings.Split(nonEmptyElements[4], ","), " |")
		}
		jobTitle = nonEmptyElements[2]

		var post string = fmt.Sprintf(
			"**📅 Date Posted:** %s\n"+
				"**ℹ️ Company:** %s\n"+
				"**👨‍💻 Job Title:** %s\n"+
				"**📍 Location:** %s\n"+
				"**➡️ When?:** %s\n\n"+
				"**👉 Job Link:** %s\n\n\n",
			datePosted, companyName, jobTitle, location, terms, jobLink,
		)
		iu.TotalJobs++

		//Work on concurrent posts
		wg := sync.WaitGroup{}
		for _, channel := range channels {
			wg.Add(1)
			go func(ch string) {
				defer wg.Done()
				discordBot.ChannelMessageSend(ch, post)
			}(channel)
		}
		wg.Wait()
	}

}
