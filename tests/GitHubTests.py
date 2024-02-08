import pytest
from unittest.mock import patch, MagicMock
import unittest.mock
from src.GitHubUtilities import GitHubUtilities

# How to test the code
# 1) Run the cmd: pytest tests/GitHubTests.py

@patch('github.Github')
def test_create_github_connection(mock_github):
    token = "token"
    repo_name = "SimplifyJobs/Summer2024-Internships"
    utilities = GitHubUtilities(token, repo_name)
    repo = utilities.createGitHubConnection()
    assert repo.owner.login == "SimplifyJobs" 


@patch('github.Repository.Repository')
def test_get_last_commit(mock_repo):
    utilities = GitHubUtilities('token', 'repo')
    mock_branch = MagicMock()
    mock_branch.commit.sha = '123abc'
    mock_repo.get_branches.return_value = [mock_branch]
    result = utilities.getLastCommit(mock_repo)
    assert result == '123abc'

@patch('github.Repository.Repository')
def test_is_new_commit(mock_repo):
    utilities = GitHubUtilities('token', 'repo')
    mock_branch = MagicMock()
    mock_branch.commit.sha = '123abc'
    mock_repo.get_branches.return_value = [mock_branch]
    result = utilities.isNewCommit(mock_repo, '456def')
    assert result

@patch('builtins.open', new_callable= unittest.mock.mock_open, read_data='123abc')
def test_get_last_saved_commit(mock_open):
    utilities = GitHubUtilities('token', 'repo')
    result = utilities.getLastSavedCommit()
    assert result == '123abc'

# Run the tests
if __name__ == '__main__':
    pytest.main([__file__])