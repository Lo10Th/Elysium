"""Tests for security hardening features."""
import pytest
from unittest.mock import MagicMock, patch
from pydantic import ValidationError
from fastapi.testclient import TestClient


class TestInputValidation:
    """Test input validation security controls."""

    def test_register_password_too_short(self, client):
        """Registration should reject passwords under 8 characters."""
        response = client.post(
            "/api/auth/register",
            json={
                "email": "test@example.com",
                "password": "short",
                "username": "testuser",
            },
        )
        assert response.status_code == 422

    def test_register_invalid_email(self, client):
        """Registration should reject invalid email addresses."""
        response = client.post(
            "/api/auth/register",
            json={
                "email": "not-an-email",
                "password": "password123",
                "username": "testuser",
            },
        )
        assert response.status_code == 422

    def test_register_invalid_username(self, client):
        """Registration should reject usernames with invalid characters."""
        response = client.post(
            "/api/auth/register",
            json={
                "email": "test@example.com",
                "password": "password123",
                "username": "bad user!",
            },
        )
        assert response.status_code == 422

    def test_register_username_too_short(self, client):
        """Registration should reject usernames shorter than 3 characters."""
        response = client.post(
            "/api/auth/register",
            json={
                "email": "test@example.com",
                "password": "password123",
                "username": "ab",
            },
        )
        assert response.status_code == 422

    def test_login_invalid_email(self, client):
        """Login should reject invalid email format."""
        response = client.post(
            "/api/auth/login",
            json={"email": "not-an-email", "password": "password123"},
        )
        assert response.status_code == 422

    def test_search_query_too_long(self, client):
        """Search endpoint should reject queries longer than 200 characters."""
        response = client.get(f"/api/emblems/search?q={'a' * 201}")
        assert response.status_code == 422

    def test_search_query_max_length_allowed(self, client):
        """Search endpoint should accept queries up to 200 characters.

        The validation error (422) is what we're testing against; if there's
        a 500 from the database call, that means validation passed, which is
        the expected behavior for a 200-char query.
        """
        response = client.get(f"/api/emblems/search?q={'a' * 200}")
        # Should not be a validation error (422)
        assert response.status_code != 422

    def test_key_name_too_long(self, client):
        """API key name over 64 characters should fail model validation."""
        from app.routes.keys import KeyCreate
        with pytest.raises(ValidationError):
            KeyCreate(name="k" * 65)

    def test_key_name_empty_rejected(self, client):
        """API key creation should reject empty names."""
        from app.routes.keys import KeyCreate
        with pytest.raises(ValidationError):
            KeyCreate(name="   ")

    def test_key_expires_days_out_of_range(self, client):
        """API key creation should reject expires_days outside 1-365 range."""
        from app.routes.keys import KeyCreate
        with pytest.raises(ValidationError):
            KeyCreate(name="test-key", expires_days=366)

    def test_key_expires_days_zero_rejected(self, client):
        """API key creation should reject expires_days of 0."""
        from app.routes.keys import KeyCreate
        with pytest.raises(ValidationError):
            KeyCreate(name="test-key", expires_days=0)


class TestSecurityHeaders:
    """Test that security headers are present in responses."""

    def test_security_headers_present(self, client):
        """All responses should include security headers."""
        response = client.get("/health")
        assert response.status_code == 200
        assert response.headers.get("x-content-type-options") == "nosniff"
        assert response.headers.get("x-frame-options") == "DENY"
        assert response.headers.get("x-xss-protection") == "1; mode=block"
        assert response.headers.get("referrer-policy") == "strict-origin-when-cross-origin"

    def test_security_headers_on_error_response(self, client):
        """Security headers should be present even on error responses."""
        response = client.get("/api/emblems/nonexistent-endpoint-xyz")
        assert response.headers.get("x-content-type-options") == "nosniff"


class TestRegisterRequestModel:
    """Unit tests for RegisterRequest validation."""

    def test_valid_registration(self):
        from app.routes.auth import RegisterRequest
        req = RegisterRequest(email="test@example.com", password="password123", username="testuser")
        assert req.email == "test@example.com"

    def test_password_minimum_8_chars(self):
        from app.routes.auth import RegisterRequest
        with pytest.raises(ValidationError):
            RegisterRequest(email="test@example.com", password="short", username="testuser")

    def test_password_exactly_8_chars_allowed(self):
        from app.routes.auth import RegisterRequest
        req = RegisterRequest(email="test@example.com", password="8charspw", username="testuser")
        assert req.password == "8charspw"

    def test_username_with_special_chars_rejected(self):
        from app.routes.auth import RegisterRequest
        with pytest.raises(ValidationError):
            RegisterRequest(email="test@example.com", password="password123", username="user name!")

    def test_username_valid_with_underscore_dash(self):
        from app.routes.auth import RegisterRequest
        req = RegisterRequest(email="test@example.com", password="password123", username="test_user-1")
        assert req.username == "test_user-1"
