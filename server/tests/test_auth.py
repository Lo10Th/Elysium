"""Tests for authentication routes."""
import pytest
from unittest.mock import Mock, MagicMock, patch
from fastapi.testclient import TestClient


class TestAuthRoutes:
    """Test authentication endpoints."""

    def test_register_success(self, client, mock_supabase, mock_auth_user):
        """Test successful user registration."""
        # Mock Supabase signup
        mock_response = MagicMock()
        mock_response.user = mock_auth_user
        mock_response.session = MagicMock()
        mock_response.session.access_token = "test-token"
        mock_response.session.refresh_token = "test-refresh-token"
        mock_supabase.auth.sign_up.return_value = mock_response

        # Make request
        response = client.post(
            "/api/auth/register",
            json={
                "email": "test@example.com",
                "password": "securepassword123",
                "username": "testuser"
            }
        )

        # Assert
        assert response.status_code in [200, 201]
        # Status code depends on implementation

    def test_register_duplicate_email(self, client, mock_supabase):
        """Test registration with duplicate email."""
        # Mock Supabase error
        mock_supabase.auth.sign_up.side_effect = Exception("User already registered")

        response = client.post(
            "/api/auth/register",
            json={
                "email": "existing@example.com",
                "password": "password123",
                "username": "existinguser"
            }
        )

        assert response.status_code in [400, 409]

    def test_login_success(self, client, mock_supabase, mock_auth_user):
        """Test successful login."""
        # Mock Supabase signin
        mock_response = MagicMock()
        mock_response.user = mock_auth_user
        mock_response.session = MagicMock()
        mock_response.session.access_token = "test-token"
        mock_response.session.refresh_token = "refresh-token"
        mock_supabase.auth.sign_in_with_password.return_value = mock_response

        response = client.post(
            "/api/auth/login",
            json={
                "email": "test@example.com",
                "password": "securepassword123"
            }
        )

        assert response.status_code in [200, 201]

    def test_login_invalid_credentials(self, client, mock_supabase):
        """Test login with invalid credentials."""
        # Mock Supabase error
        mock_supabase.auth.sign_in_with_password.side_effect = Exception("Invalid credentials")

        response = client.post(
            "/api/auth/login",
            json={
                "email": "test@example.com",
                "password": "wrongpassword"
            }
        )

        assert response.status_code in [401, 400]

    def test_logout_success(self, client, mock_supabase):
        """Test successful logout."""
        mock_supabase.auth.sign_out.return_value = None

        response = client.post(
            "/api/auth/logout",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 204]

    def test_get_current_user_success(self, client, mock_supabase, mock_auth_user):
        """Test getting current authenticated user."""
        # Mock Supabase get_user
        mock_response = MagicMock()
        mock_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_response

        response = client.get(
            "/api/auth/me",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code == 200
        # Response should contain user data

    def test_get_current_user_unauthorized(self, client, mock_supabase):
        """Test getting current user without token."""
        # Mock Supabase error
        mock_supabase.auth.get_user.side_effect = Exception("Invalid token")

        response = client.get("/api/auth/me")

        assert response.status_code in [401, 403]

    def test_refresh_token_success(self, client, mock_supabase, mock_auth_user):
        """Test token refresh."""
        # Mock Supabase refresh
        mock_response = MagicMock()
        mock_response.user = mock_auth_user
        mock_response.session = MagicMock()
        mock_response.session.access_token = "new-test-token"
        mock_response.session.refresh_token = "new-refresh-token"
        mock_supabase.auth.refresh_session.return_value = mock_response

        response = client.post(
            "/api/auth/refresh",
            json={"refresh_token": "old-refresh-token"}
        )

        assert response.status_code in [200, 201]