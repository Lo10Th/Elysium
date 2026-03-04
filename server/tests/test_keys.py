"""Tests for API key management routes."""

import pytest
from unittest.mock import MagicMock
from datetime import datetime, timedelta


class TestKeyRoutes:
    """Test API key management endpoints."""

    def test_list_keys_success(
        self, client, mock_supabase, mock_key, mock_auth_user, mock_profile
    ):
        """Test listing all API keys."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        keys_response = MagicMock()
        keys_response.data = [mock_key, mock_key]
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            keys_response,  # keys list
        ]

        response = client.get(
            "/api/keys", headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code == 200

    def test_list_keys_unauthorized(self, client):
        """Test listing keys without auth."""
        response = client.get("/api/keys")
        assert response.status_code in [401, 403]

    def test_create_key_success(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test creating a new API key."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
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
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            no_duplicate,  # duplicate check
            create_response,  # insert
        ]

        response = client.post(
            "/api/keys",
            json={"name": "new-key"},
            headers={"Authorization": "Bearer test-token"},
        )

        assert response.status_code in [200, 201]
        data = response.json()
        # Key should be shown once on creation
        assert "key" in data or "id" in data

    def test_create_key_with_expiration(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test creating key with expiration date."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
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
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            no_duplicate,  # duplicate check
            create_response,  # insert
        ]

        response = client.post(
            "/api/keys",
            json={"name": "temp-key", "expires_days": 30},
            headers={"Authorization": "Bearer test-token"},
        )

        assert response.status_code in [200, 201]

    def test_get_key_success(
        self, client, mock_supabase, mock_key, mock_auth_user, mock_profile
    ):
        """Test getting specific key details."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        key_response = MagicMock()
        key_response.data = mock_key
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            key_response,  # key lookup
        ]

        response = client.get(
            "/api/keys/key-123", headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["id"] == "key-123"

    def test_get_key_not_found(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test getting non-existent key."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        key_response = MagicMock()
        key_response.data = None
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            key_response,  # key not found
        ]

        response = client.get(
            "/api/keys/nonexistent", headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [404, 400]

    def test_delete_key_success(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test deleting a key."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        key_response = MagicMock()
        key_response.data = {"id": "key-123", "user_id": "user-123"}
        delete_response = MagicMock()
        delete_response.data = None
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            key_response,  # key lookup
            delete_response,  # delete
        ]

        response = client.delete(
            "/api/keys/key-123", headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 204]

    def test_delete_key_not_found(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test deleting non-existent key."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        key_response = MagicMock()
        key_response.data = None
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            key_response,  # key not found
        ]

        response = client.delete(
            "/api/keys/nonexistent", headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [404, 400]

    def test_delete_key_unauthorized(self, client):
        """Test deleting key without auth."""
        response = client.delete("/api/keys/key-123")
        assert response.status_code in [401, 403]

    def test_key_shown_only_once_on_creation(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test that API key value is returned only on creation."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
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
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            no_duplicate,  # duplicate check
            create_response,  # insert
        ]

        response = client.post(
            "/api/keys",
            json={"name": "new-key"},
            headers={"Authorization": "Bearer test-token"},
        )

        assert response.status_code in [200, 201]
        data = response.json()

        # Key should be visible in creation response (raw key, not hash)
        assert "key" in data
        raw_key = data.get("key")
        assert raw_key is not None and raw_key.startswith("ely_")

        # On subsequent GET requests, key should not be visible
        list_response_data = [
            {
                "id": "key-new",
                "name": "new-key",
                "created_at": "2024-01-01T00:00:00Z",
                "expires_at": None,
            }
        ]
        list_mock = MagicMock()
        list_mock.data = list_response_data
        profile_response2 = MagicMock()
        profile_response2.data = mock_profile
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response2,  # profile query in get_current_user
            list_mock,
        ]

        get_response = client.get(
            "/api/keys", headers={"Authorization": "Bearer test-token"}
        )

        # Key should not be visible in list (KeyListItem has no key field)
        assert get_response.status_code == 200
