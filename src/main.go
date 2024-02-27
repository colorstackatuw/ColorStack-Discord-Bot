package main

import (
	"ColorStack-Discord-Bot/Utilities"
	"context"
	"os"
	"fmt"
)

func main() {

	// Testing for GitHub applications
	ctx := context.Background()
	var token string = os.Getenv("GIT_TOKEN") 
	var repoName string = "Summer2024-Internships"
	github := Utilities.NewGitHubUtilities(token, repoName)

	repo, err := github.CreateGitHubConnection(ctx)
	if err != nil{
		fmt.Println("Failed with CreatedGitHubConnection")
		panic(err)
	}

	lastCommit, err := github.GetLastCommit(ctx, repo)
	if err != nil{
		fmt.Println("Failed with GetLastCommit")
		panic(err)
	}
	fmt.Println(lastCommit)

	isNew, err := github.IsNewCommit(lastCommit)
	if err != nil{
		fmt.Println("Failed with IsNewCommit")
		panic(err)
	}

	fmt.Println(isNew)
	if isNew{
		github.SetNewCommit(lastCommit)
		
	}


	// Testing for Internship Utilities 

}