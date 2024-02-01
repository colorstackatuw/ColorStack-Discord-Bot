import pytest
from unittest.mock import MagicMock
import github
from src.InternshipUtilities import InternshipUtilities  # Adjust the import as necessary

# Mock the GitHub repository object
@pytest.fixture
def mock_repo():
    mock = MagicMock(spec=github.Repository.Repository)
    # Simulate get_contents returning mocked README.md content
    mock.get_contents.return_value = MagicMock(
        decoded_content=b"Mocked README content"
    )
    return mock

@pytest.fixture
def internship_utilities(mock_repo):
    return InternshipUtilities(mock_repo, summer=True, co_op=False)

# Test the binarySearchUS method
def test_binarySearchUS_valid_state(internship_utilities):
    assert internship_utilities.binarySearchUS("CA") == True

def test_binarySearchUS_invalid_state(internship_utilities):
    assert internship_utilities.binarySearchUS("XX") == False

# Run the tests
if __name__ == '__main__':
    pytest.main([__file__])