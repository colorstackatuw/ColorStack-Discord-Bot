from unittest.mock import MagicMock, patch
import pytest
from src.Utilities.GitHubUtilities import GitHubUtilities

# How to test the code
# 1) Run the cmd: pytest tests/test_GitHub.py

@patch('os.getenv')
def test_create_github_connection(mock_getenv):
    # Arrange
    mock_getenv.return_value = 'mock_token'
    repo_name = "SimplifyJobs/Summer2025-Internships"
    utilities = GitHubUtilities('mock_token', repo_name)

    # Act
    mock_repo = MagicMock()
    mock_repo.name = repo_name.split("/")[-1]
    with patch.object(utilities, 'createGitHubConnection', return_value=mock_repo):
        repo = utilities.createGitHubConnection()

    # Assert
    assert repo is not None
    assert repo.name == repo_name.split("/")[-1]


@patch("github.Repository.Repository")
def test_get_last_commit(mock_repo):
    utilities = GitHubUtilities("token", "repo")
    mock_branch = MagicMock()  # Fake a respository
    mock_branch.commit.sha = "123abc"
    mock_repo.get_branches.return_value = [mock_branch]
    result = utilities.getLastCommit(mock_repo)
    assert result == "123abc"


@patch("github.Repository.Repository")
def test_is_new_commit(mock_repo):
    utilities = GitHubUtilities("token", "repo")
    mock_branch = MagicMock()  # Fake a respository
    mock_branch.commit.sha = "123abc"
    mock_repo.get_branches.return_value = [mock_branch]
    result = utilities.isNewCommit(mock_repo, "456def")
    assert result


@pytest.fixture
def setup_github_utilities():
    token = "your_token"
    repo_name = "SimplifyJobs/Summer2025-Internships"
    utilities = GitHubUtilities(token, repo_name)
    return utilities
