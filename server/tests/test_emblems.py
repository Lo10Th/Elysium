"""Tests for emblem routes."""

import pytest
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient

VALID_EMBLEM_YAML = """apiVersion: v1
name: new-api
version: 1.0.0
description: A new API for testing purposes
baseUrl: https://api.example.com
actions:
  list:
    description: List all items
    method: GET
    path: /items
"""


class TestEmblemRoutes:
    """Test emblem CRUD operations."""

    def test_list_emblems_success(self, client, mock_supabase, mock_emblem):
        """Test listing all emblems."""
        # Mock Supabase query
        mock_response = MagicMock()
        mock_response.data = [mock_emblem, mock_emblem]
        mock_supabase.table.return_value.select.return_value.execute.return_value = (
            mock_response
        )

        response = client.get("/api/emblems")

        assert response.status_code == 200
        # Should return list of emblems

    def test_list_emblems_with_category_filter(
        self, client, mock_supabase, mock_emblem
    ):
        """Test listing emblems filtered by category."""
        mock_response = MagicMock()
        mock_response.data = [mock_emblem]
        mock_supabase.table.return_value.select.return_value.eq.return_value.execute.return_value = mock_response

        response = client.get("/api/emblems?category=payments")

        assert response.status_code == 200

    def test_list_emblems_with_search(self, client, mock_supabase, mock_emblem):
        """Test searching emblems."""
        mock_response = MagicMock()
        mock_response.data = [mock_emblem]
        mock_supabase.table.return_value.select.return_value.ilike.return_value.execute.return_value = mock_response

        response = client.get("/api/emblems?search=test")

        assert response.status_code == 200

    def test_get_emblem_by_name_success(self, client, mock_supabase, mock_emblem):
        """Test getting emblem by name."""
        mock_response = MagicMock()
        mock_response.data = mock_emblem
        mock_supabase.table.return_value.select.return_value.eq.return_value.single.return_value.execute.return_value = mock_response

        response = client.get("/api/emblems/test-api")

        assert response.status_code == 200
        data = response.json()
        assert data["name"] == "test-api"

    def test_get_emblem_not_found(self, client, mock_supabase):
        """Test getting non-existent emblem."""
        # Mock Supabase error
        mock_response = MagicMock()
        mock_response.data = None
        mock_supabase.table.return_value.select.return_value.eq.return_value.single.return_value.execute.return_value = mock_response

        response = client.get("/api/emblems/nonexistent")

        assert response.status_code in [404, 400]

    def test_get_emblem_version_success(self, client, mock_supabase, mock_emblem):
        """Test getting specific emblem version."""
        # Mock emblem lookup
        emblem_response = MagicMock()
        emblem_response.data = {"id": "emblem-123"}

        # Mock version lookup
        version_data = {
            "version": "1.0.0",
            "yaml_content": "apiVersion: v1\nname: test-api",
            "changelog": None,
            "published_at": "2024-01-01T00:00:00Z",
        }
        version_response = MagicMock()
        version_response.data = version_data

        mock_supabase.table.return_value.execute.side_effect = [
            emblem_response,  # First call: get emblem ID
            version_response,  # Second call: get version
        ]

        response = client.get("/api/emblems/test-api/1.0.0")

        assert response.status_code == 200
        data = response.json()
        assert data["version"] == "1.0.0"

    def test_create_emblem_success(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test creating a new emblem."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        no_duplicate = MagicMock()
        no_duplicate.data = []
        new_emblem_row = {
            "id": "emblem-new",
            "name": "new-api",
            "description": "New API for testing purposes",
            "author_id": "user-123",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z",
        }
        create_response = MagicMock()
        create_response.data = [new_emblem_row]
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            no_duplicate,  # duplicate check
            create_response,  # insert emblem
            MagicMock(),  # insert version
            MagicMock(),  # insert pull
        ]

        response = client.post(
            "/api/emblems",
            json={
                "name": "new-api",
                "version": "1.0.0",
                "description": "New API for testing purposes",
                "yaml_content": VALID_EMBLEM_YAML,
                "license": "MIT",
            },
            headers={"Authorization": "Bearer test-token"},
        )

        assert response.status_code in [200, 201]

    def test_create_emblem_unauthorized(self, client):
        """Test creating emblem without auth."""
        response = client.post(
            "/api/emblems",
            json={"name": "new-api", "version": "1.0.0", "license": "MIT"},
        )

        assert response.status_code in [401, 403]

    def test_create_emblem_duplicate_name(
        self, client, mock_supabase, mock_auth_user, mock_profile
    ):
        """Test creating emblem with duplicate name."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        duplicate_response = MagicMock()
        duplicate_response.data = [{"id": "emblem-existing"}]
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            duplicate_response,  # duplicate check returns existing
        ]

        response = client.post(
            "/api/emblems",
            json={
                "name": "existing-api",
                "version": "1.0.0",
                "description": "An existing API description",
                "yaml_content": VALID_EMBLEM_YAML,
                "license": "MIT",
            },
            headers={"Authorization": "Bearer test-token"},
        )

        assert response.status_code in [400, 409]

    def test_update_emblem_success(
        self, client, mock_supabase, mock_auth_user, mock_emblem, mock_profile
    ):
        """Test updating an emblem."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        emblem_response = MagicMock()
        emblem_response.data = mock_emblem
        no_existing_version = MagicMock()
        no_existing_version.data = []
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            emblem_response,  # get emblem ownership check
            no_existing_version,  # check existing version
            MagicMock(),  # insert new version
            MagicMock(),  # update emblem
        ]

        response = client.put(
            "/api/emblems/test-api",
            json={
                "yaml_content": VALID_EMBLEM_YAML,
                "version": "2.0.0",
                "description": "Updated description here",
            },
            headers={"Authorization": "Bearer test-token"},
        )

        assert response.status_code == 200

    def test_delete_emblem_success(
        self, client, mock_supabase, mock_auth_user, mock_emblem, mock_profile
    ):
        """Test deleting an emblem."""
        profile_response = MagicMock()
        profile_response.data = mock_profile
        emblem_response = MagicMock()
        emblem_response.data = mock_emblem
        delete_response = MagicMock()
        delete_response.data = None
        mock_supabase.table.return_value.execute.side_effect = [
            profile_response,  # profile query in get_current_user
            emblem_response,  # get emblem for ownership check
            delete_response,  # delete operation
        ]

        response = client.delete(
            "/api/emblems/test-api", headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 204]
