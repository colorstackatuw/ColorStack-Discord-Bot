package crawler_test

import (
	"ColorStack-Discord-Bot/internal/crawler"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-github/v59/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitHubClient mocks github.Client behavior
type MockGitHubClient struct {
	mock.Mock
}

func TestSetNewCommit(t *testing.T) {
	util := &crawler.GitHubUtilities{}

	tmpFile := t.TempDir() + "/test.json"
	crawler.FILEPATH = tmpFile

	err := util.SetNewCommit("abc123", true)
	assert.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", result["last_saved_sha_newgrad"])
}

func TestSetSavedSha(t *testing.T) {
	util := &crawler.GitHubUtilities{}

	content := map[string]string{"last_saved_sha_internship": "test123"}
	data, _ := json.Marshal(content)

	tmpFile := t.TempDir() + "/test.json"
	_ = os.WriteFile(tmpFile, data, 0644)
	crawler.FILEPATH = tmpFile

	err := util.SetSavedSha(false)
	assert.NoError(t, err)
	assert.Equal(t, "test123", util.SavedSHA)
}

func TestCreateGitHubConnection(t *testing.T) {
	token := "dummy"
	repoName := "fake-repo"
	util := crawler.NewGitHubUtilities(token, repoName, false, false)

	ctx := context.Background()
	_, err := util.CreateGitHubConnection(ctx)

	// Will fail without network access/token, so just check if error is wrapped
	assert.Error(t, err)
}

func TestGetLastCommit_Failure(t *testing.T) {
	util := crawler.NewGitHubUtilities("", "invalid-repo", false, false)

	ctx := context.Background()
	repo := &github.Repository{
		Owner: &github.User{Login: github.String("fake")},
		Name:  github.String("fake"),
	}
	_, err := util.GetLastCommit(ctx, repo)
	assert.Error(t, err)
}

func TestIsNewCommit(t *testing.T) {
	util := crawler.NewGitHubUtilities("", "invalid-repo", false, false)

	ctx := context.Background()
	repo := &github.Repository{
		Owner: &github.User{Login: github.String("fake")},
		Name:  github.String("fake"),
	}
	result, err := util.IsNewCommit(ctx, repo, "abcdef")
	assert.Error(t, err)
	assert.False(t, result)
}

func TestGetCommitChanges(t *testing.T) {
	util := &crawler.GitHubUtilities{}

	mockPatch := "+New line\n+++ b/file\n+🔒 secret\n+Actual Line"
	files := []*github.CommitFile{
		{
			Filename: github.String("README.md"),
			Patch:    github.String(mockPatch),
		},
	}
	util.Comparison = &github.CommitsComparison{
		Files: files,
	}

	results := []string{}
	for line := range util.GetCommitChanges("README.md") {
		results = append(results, line)
	}

	assert.Equal(t, []string{"+New line", "+Actual Line"}, results)
}
