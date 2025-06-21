package jobs

import (
	"strings"
)

type JobUtilities struct {
	PreviousJobTitle string
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
		NotUS: [4]string{"canada", "uk", "united kingdom", "eu"},
	}
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
