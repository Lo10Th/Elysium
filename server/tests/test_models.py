"""Tests for Pydantic models."""
import pytest
from pydantic import ValidationError
from app.models import (
    EmblemCreate, Emblem, EmblemVersion,
    KeyCreate, KeyResponse
)


class TestEmblemModels:
    """Test emblem-related Pydantic models."""

    def test_emblem_create_valid(self):
        """Test valid emblem creation."""
        data = {
            "name": "test-api",
            "version": "1.0.0",
            "description": "Test API",
            "license": "MIT"
        }
        emblem = EmblemCreate(**data)
        assert emblem.name == "test-api"
        assert emblem.version == "1.0.0"
        assert emblem.description == "Test API"
        assert emblem.license == "MIT"

    def test_emblem_create_minimal(self):
        """Test emblem creation with minimal fields."""
        data = {"name": "test-api", "version": "1.0.0", "license": "MIT"}
        emblem = EmblemCreate(**data)
        assert emblem.name == "test-api"
        assert emblem.description is None

    def test_emblem_response(self):
        """Test emblem response model."""
        from app.models import Emblem
        data = {
            "id": "emblem-123",
            "name": "test-api",
            "description": "Test API",
            "license": "MIT",
            "latest_version": "1.0.0",
            "downloads_count": 100
        }
        response = Emblem(**data)
        assert response.id == "emblem-123"
        assert response.name == "test-api"

    def test_emblem_version_valid(self):
        """Test valid emblem version."""
        data = {
            "version": "1.0.0",
            "yaml_content": "apiVersion: v1\nname: test",
            "readme_content": "# Test API"
        }
        version = EmblemVersion(**data)
        assert version.version == "1.0.0"
        assert "# Test API" in version.readme_content


class TestKeyModels:
    """Test API key models."""

    def test_key_create_valid(self):
        """Test valid key creation."""
        data = {"name": "test-key"}
        key = KeyCreate(**data)
        assert key.name == "test-key"

    def test_key_create_with_expiration(self):
        """Test key creation with expiration."""
        data = {"name": "test-key", "expires_days": 30}
        key = KeyCreate(**data)
        assert key.name == "test-key"
        assert key.expires_days == 30

    def test_key_response(self):
        """Test key response model."""
        from datetime import datetime
        data = {
            "id": "key-123",
            "name": "test-key",
            "created_at": datetime(2024, 1, 1)
        }
        response = KeyResponse(**data)
        assert response.id == "key-123"
        assert response.name == "test-key"


class TestAuthModels:
    """Test authentication models from routes."""

    def test_login_request_valid(self):
        """Test valid login request."""
        from app.routes.auth import LoginRequest
        data = {
            "email": "test@example.com",
            "password": "securepassword123"
        }
        login = LoginRequest(**data)
        assert login.email == "test@example.com"

    def test_register_request_valid(self):
        """Test valid registration request."""
        from app.routes.auth import RegisterRequest
        data = {
            "email": "test@example.com",
            "password": "securepassword123",
            "username": "testuser"
        }
        register = RegisterRequest(**data)
        assert register.email == "test@example.com"
        assert register.username == "testuser"