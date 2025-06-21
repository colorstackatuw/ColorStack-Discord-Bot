package crawler

type JobType string

const (
	NEWGRAD JobType = "New Grad"
	SUMMER  JobType = "Summer"
	COOP    JobType = "Co-Op"
)

type Job struct {
	company  string
	jobTitle string
	location string
	season   string
}
