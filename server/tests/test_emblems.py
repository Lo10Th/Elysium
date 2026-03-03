"""Tests for emblem routes."""
import pytest
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient


class TestEmblemRoutes:
    """Test emblem CRUD operations."""

    def test_list_emblems_success(self, client, mock_supabase, mock_emblem):
        """Test listing all emblems."""
        # Mock Supabase query
        mock_response = MagicMock()
        mock_response.data = [mock_emblem, mock_emblem]
        mock_supabase.table.return_value.select.return_value.execute.return_value = mock_response

        response = client.get("/api/emblems")

        assert response.status_code == 200
        # Should return list of emblems

    def test_list_emblems_with_category_filter(self, client, mock_supabase, mock_emblem):
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
            "readme_content": "# Test API"
        }
        version_response = MagicMock()
        version_response.data = version_data
        
        mock_supabase.table.return_value.select.return_value.eq.return_value.single.return_value.execute.side_effect = [
            emblem_response,  # First call: get emblem ID
            version_response  # Second call: get version
        ]

        response = client.get("/api/emblems/test-api/1.0.0")

        assert response.status_code == 200
        data = response.json()
        assert data["version"] == "1.0.0"

    def test_create_emblem_success(self, client, mock_supabase, mock_auth_user):
        """Test creating a new emblem."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock emblem creation
        mock_response = MagicMock()
        mock_response.data = {
            "id": "emblem-new",
            "name": "new-api",
            "description": "New API",
            "author_id": "user-123"
        }
        mock_supabase.table.return_value.insert.return_value.execute.return_value = mock_response

        # Mock version creation
        mock_supabase.table.return_value.insert.return_value.execute.return_value = mock_response

        response = client.post(
            "/api/emblems",
            json={
                "name": "new-api",
                "version": "1.0.0",
                "description": "New API",
                "license": "MIT"
            },
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 201]

    def test_create_emblem_unauthorized(self, client):
        """Test creating emblem without auth."""
        response = client.post(
            "/api/emblems",
            json={
                "name": "new-api",
                "version": "1.0.0",
                "license": "MIT"
            }
        )

        assert response.status_code in [401, 403]

    def test_create_emblem_duplicate_name(self, client, mock_supabase, mock_auth_user):
        """Test creating emblem with duplicate name."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock existing emblem
        mock_response = MagicMock()
        mock_response.data = [{"id": "emblem-existing"}]
        mock_supabase.table.return_value.select.return_value.eq.return_value.execute.return_value = mock_response

        response = client.post(
            "/api/emblems",
            json={
                "name": "existing-api",
                "version": "1.0.0",
                "license": "MIT"
            },
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [400, 409]

    def test_update_emblem_success(self, client, mock_supabase, mock_auth_user, mock_emblem):
        """Test updating an emblem."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock emblem ownership check
        emblem_response = MagicMock()
        emblem_response.data = mock_emblem
        mock_supabase.table.return_value.select.return_value.eq.return_value.single.return_value.execute.return_value = emblem_response

        # Mock update
        mock_response = MagicMock()
        mock_response.data = mock_emblem
        mock_supabase.table.return_value.update.return_value.eq.return_value.execute.return_value = mock_response

        response = client.put(
            "/api/emblems/test-api",
            json={"description": "Updated description"},
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code == 200

    def test_delete_emblem_success(self, client, mock_supabase, mock_auth_user, mock_emblem):
        """Test deleting an emblem."""
        # Mock auth
        mock_auth_response = MagicMock()
        mock_auth_response.user = mock_auth_user
        mock_supabase.auth.get_user.return_value = mock_auth_response

        # Mock emblem ownership
        emblem_response = MagicMock()
        emblem_response.data = mock_emblem
        mock_supabase.table.return_value.select.return_value.eq.return_value.single.return_value.execute.return_value = emblem_response

        # Mock delete
        mock_supabase.table.return_value.delete.return_value.eq.return_value.execute.return_value = MagicMock()

        response = client.delete(
            "/api/emblems/test-api",
            headers={"Authorization": "Bearer test-token"}
        )

        assert response.status_code in [200, 204]