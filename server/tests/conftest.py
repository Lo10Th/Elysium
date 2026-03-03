"""Pytest fixtures and configuration for Elysium server tests."""
import os
import pytest
from unittest.mock import Mock, MagicMock, patch
from fastapi.testclient import TestClient

# Set test environment variables BEFORE importing app
os.environ["SUPABASE_URL"] = "https://test.supabase.co"
os.environ["SUPABASE_ANON_KEY"] = "test-anon-key"
os.environ["SUPABASE_SERVICE_KEY"] = "test-service-key"
os.environ["SECRET_KEY"] = "test-secret-key-for-testing-only"

from app.main import app


def make_chainable_query():
    """Create a MagicMock query that returns itself for all chaining methods."""
    query = MagicMock()
    for method in (
        'select', 'eq', 'neq', 'gt', 'gte', 'lt', 'lte',
        'like', 'ilike', 'or_', 'order', 'range', 'single',
        'limit', 'insert', 'update', 'delete', 'upsert',
    ):
        getattr(query, method).return_value = query
    return query


@pytest.fixture
def client():
    """Create a test client for the FastAPI app."""
    return TestClient(app)


@pytest.fixture
def mock_supabase():
    """Mock Supabase client with chainable query support."""
    mock = MagicMock()

    # Set up default auth mock with valid string attributes so auth checks pass
    auth_user = MagicMock()
    auth_user.id = "user-123"
    auth_user.email = "test@example.com"
    mock.auth.get_user.return_value.user = auth_user

    # Set up chainable query mock for table operations
    query = make_chainable_query()
    mock.table.return_value = query

    with patch('app.routes.auth.get_supabase', return_value=mock), \
         patch('app.routes.emblems.get_supabase', return_value=mock), \
         patch('app.routes.keys.get_supabase', return_value=mock):
        yield mock


@pytest.fixture
def mock_auth_user(mock_supabase):
    """Mock authenticated user (uses the auth user configured in mock_supabase)."""
    return mock_supabase.auth.get_user.return_value.user


@pytest.fixture
def mock_emblem():
    """Mock emblem data."""
    return {
        "id": "emblem-123",
        "name": "test-api",
        "description": "Test API description",
        "author_id": "user-123",
        "author_name": "testuser",
        "category": "general",
        "tags": ["test", "api"],
        "license": "MIT",
        "repository_url": "https://github.com/test/test-api",
        "homepage_url": "https://test-api.example.com",
        "latest_version": "1.0.0",
        "downloads_count": 100,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-02T00:00:00Z"
    }


@pytest.fixture
def mock_key():
    """Mock API key."""
    return {
        "id": "key-123",
        "name": "test-key",
        "key": "sk_test_abc123",
        "created_at": "2024-01-01T00:00:00Z",
        "expires_at": None
    }


@pytest.fixture
def auth_headers():
    """Mock authorization headers."""
    return {"Authorization": "Bearer test-token"}


@pytest.fixture
def mock_supabase_response():
    """Mock Supabase query response."""
    def _make_response(data=None, error=None):
        response = MagicMock()
        response.data = data or []
        response.error = error
        response.status_code = 200 if not error else 400
        return response
    return _make_response
