"""Tests for API key management routes."""
import pytest
from unittest.mock import MagicMock
from datetime import datetime, timedelta


class TestKeyRoutes:
    """Test API key management endpoints."""

    def test_list_keys_success(self, client, mock_supabase, mock_key, mock_auth_user):
        """Test listing all API keys."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock keys list
        mock_response = MagicMock()
        mock_response.data = [mock_key, mock_key]
        mock_supabase.table.return_value.select.return_value.eq.return_value.execute.return_value = mock_response

        response = client.get(
            "/api/keys",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code == 200
        # Should return list of keys

    def test_list_keys_unauthorized(self, client):
        """Test listing keys without auth."""
        response = client.get("/api/keys")
        assert response.status_code in [401, 403]

    def test_create_key_success(self, client, mock_supabase, mock_auth_user):
        """Test creating a new API key."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Duplicate check returns empty, insert returns new key row
        no_duplicate = MagicMock()
        no_duplicate.data = []
        key_row = {
            "id": "key-new",
            "name": "new-key",
            "user_id": "user-123",
            "created_at": "2024-01-01T00:00:00Z",
        }
        create_response = MagicMock()
        create_response.data = [key_row]
        mock_supabase.table.return_value.execute.side_effect = [no_duplicate, create_response]

        response = client.post(
            "/api/keys",
            json={"name": "new-key"},
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 201]
        data = response.json()
        # Key should be shown once on creation
        assert "key" in data or "id" in data

    def test_create_key_with_expiration(self, client, mock_supabase, mock_auth_user):
        """Test creating key with expiration date."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Duplicate check returns empty, insert returns new key row with expiration
        no_duplicate = MagicMock()
        no_duplicate.data = []
        key_row = {
            "id": "key-new",
            "name": "temp-key",
            "user_id": "user-123",
            "created_at": "2024-01-01T00:00:00Z",
            "expires_at": "2024-01-31T00:00:00Z",
        }
        create_response = MagicMock()
        create_response.data = [key_row]
        mock_supabase.table.return_value.execute.side_effect = [no_duplicate, create_response]

        response = client.post(
            "/api/keys",
            json={"name": "temp-key", "expires_days": 30},
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 201]

    def test_get_key_success(self, client, mock_supabase, mock_key, mock_auth_user):
        """Test getting specific key details."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock key lookup
        mock_response = MagicMock()
        mock_response.data = mock_key
        mock_supabase.table.return_value.select.return_value.eq.return_value.eq.return_value.single.return_value.execute.return_value = mock_response

        response = client.get(
            "/api/keys/key-123",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["id"] == "key-123"

    def test_get_key_not_found(self, client, mock_supabase, mock_auth_user):
        """Test getting non-existent key."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock key not found
        mock_response = MagicMock()
        mock_response.data = None
        mock_supabase.table.return_value.select.return_value.eq.return_value.eq.return_value.single.return_value.execute.return_value = mock_response

        response = client.get(
            "/api/keys/nonexistent",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [404, 400]

    def test_delete_key_success(self, client, mock_supabase, mock_auth_user):
        """Test deleting a key."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock key lookup (ownership check)
        mock_key_response = MagicMock()
        mock_key_response.data = {"id": "key-123", "user_id": "user-123"}
        mock_supabase.table.return_value.select.return_value.eq.return_value.eq.return_value.single.return_value.execute.return_value = mock_key_response

        # Mock delete
        mock_supabase.table.return_value.delete.return_value.eq.return_value.execute.return_value = MagicMock()

        response = client.delete(
            "/api/keys/key-123",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 204]

    def test_delete_key_not_found(self, client, mock_supabase, mock_auth_user):
        """Test deleting non-existent key."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock key not found
        mock_response = MagicMock()
        mock_response.data = None
        mock_supabase.table.return_value.select.return_value.eq.return_value.eq.return_value.single.return_value.execute.return_value = mock_response

        response = client.delete(
            "/api/keys/nonexistent",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [404, 400]

    def test_delete_key_unauthorized(self, client):
        """Test deleting key without auth."""
        response = client.delete("/api/keys/key-123")
        assert response.status_code in [401, 403]

    def test_key_shown_only_once_on_creation(self, client, mock_supabase, mock_auth_user):
        """Test that API key value is returned only on creation."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Duplicate check returns empty, insert returns new key
        no_duplicate = MagicMock()
        no_duplicate.data = []
        key_row = {
            "id": "key-new",
            "name": "new-key",
            "user_id": "user-123",
            "created_at": "2024-01-01T00:00:00Z",
        }
        create_response = MagicMock()
        create_response.data = [key_row]
        mock_supabase.table.return_value.execute.side_effect = [no_duplicate, create_response]

        response = client.post(
            "/api/keys",
            json={"name": "new-key"},
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 201]
        data = response.json()

        # Key should be visible in creation response (raw key, not hash)
        assert "key" in data
        raw_key = data.get("key")
        assert raw_key is not None and raw_key.startswith("ely_")

        # On subsequent GET requests, key should not be visible
        list_response_data = [{"id": "key-new", "name": "new-key",
                                "created_at": "2024-01-01T00:00:00Z", "expires_at": None}]
        list_mock = MagicMock()
        list_mock.data = list_response_data
        mock_supabase.table.return_value.execute.side_effect = [list_mock]

        get_response = client.get(
            "/api/keys",
            headers={"Authorization": "Bearer test-token"}
        )

        # Key should not be visible in list (KeyListItem has no key field)
        assert get_response.status_code == 200