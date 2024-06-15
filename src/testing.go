package 

import (
	"ColorStack-Discord-Bot/Utilities"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main1() {

	// Testing for GitHub applications
	ctx := context.Background()
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal(err)
	}
	var token string = os.Getenv("GIT_TOKEN")
	var repoName string = "Summer2024-Internships"
	github := Utilities.NewGitHubUtilities(token, repoName)

	repo, err := github.CreateGitHubConnection(ctx)
	if err != nil {
		fmt.Println("Failed with CreatedGitHubConnection")
		panic(err)
	}

	lastCommit, err := github.GetLastCommit(ctx, repo)
	if err != nil {
		fmt.Println("Failed with GetLastCommit")
		panic(err)
	}
	fmt.Println(lastCommit)

	isNew, err := github.IsNewCommit(lastCommit)
	if err != nil {
		fmt.Println("Failed with IsNewCommit")
		panic(err)
	}

	fmt.Println(isNew)
	if isNew {
		//github.SetNewCommit(lastCommit)
		savedSha, err := github.GetSavedSha()
		if err != nil {
			panic(err)
		}
		github.SetComparison(ctx, repo, savedSha)
		iu := Utilities.NewInternshipUtilities(true)
		jobChannel := github.GetCommitChanges("README-Off-Season.md")

		array := []string{"1210677457397747852"}
		iu.GetInternships(array, jobChannel, false)
	}

}
