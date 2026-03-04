"""Direct unit tests for the service layer.

These tests exercise service methods directly with a mocked Supabase client,
covering error paths and edge cases that are not reachable through the thin
route-level tests (which mock at the route level).
"""

import pytest
from unittest.mock import MagicMock, patch
from fastapi import HTTPException

from app.services.emblem_service import EmblemService, _row_to_emblem
from app.services.key_service import KeyService
from app.services.auth_service import AuthService
from app.models import User, EmblemCreate, EmblemUpdate, KeyCreate


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def make_supabase():
    """Return a fresh MagicMock configured to behave like a Supabase client."""
    sb = MagicMock()
    q = MagicMock()
    for m in (
        "select", "eq", "neq", "gt", "gte", "lt", "lte",
        "ilike", "or_", "order", "range", "single", "limit",
        "insert", "update", "delete", "upsert", "maybe_single", "rpc",
    ):
        getattr(q, m).return_value = q
    sb.table.return_value = q
    sb.rpc.return_value = q
    return sb, q


def mock_response(data=None):
    r = MagicMock()
    r.data = data if data is not None else []
    return r


SAMPLE_USER = User(id="u-1", email="a@b.com", username="alice")

SAMPLE_ROW = {
    "id": "e-1",
    "name": "my-api",
    "description": "A great API for testing",
    "author_id": "u-1",
    "profiles": {"username": "alice"},
    "category": "finance",
    "tags": ["tag1"],
    "license": "MIT",
    "repository_url": None,
    "homepage_url": None,
    "latest_version": "1.0.0",
    "downloads_count": 42,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z",
}


# ---------------------------------------------------------------------------
# EmblemService._row_to_emblem
# ---------------------------------------------------------------------------

class TestRowToEmblem:
    def test_with_profiles(self):
        emblem = _row_to_emblem(SAMPLE_ROW)
        assert emblem.author_name == "alice"
        assert emblem.downloads_count == 42

    def test_without_profiles(self):
        row = {**SAMPLE_ROW, "profiles": None}
        emblem = _row_to_emblem(row)
        assert emblem.author_name is None

    def test_profiles_not_dict(self):
        row = {**SAMPLE_ROW, "profiles": "unexpected-string"}
        emblem = _row_to_emblem(row)
        assert emblem.author_name is None


# ---------------------------------------------------------------------------
# EmblemService.validate_yaml
# ---------------------------------------------------------------------------

VALID_YAML = """apiVersion: v1
name: new-api
version: 1.0.0
description: A new API for testing purposes that is long enough
baseUrl: https://api.example.com
actions:
  list:
    description: List all items
    method: GET
    path: /items
"""

class TestValidateYaml:
    def test_valid_yaml_returns_dict(self):
        result = EmblemService.validate_yaml(VALID_YAML)
        assert isinstance(result, dict)
        assert result["name"] == "new-api"

    def test_invalid_yaml_raises_400(self):
        with pytest.raises(HTTPException) as exc:
            EmblemService.validate_yaml("key: [unclosed")
        assert exc.value.status_code == 400
        assert "YAML" in exc.value.detail

    def test_schema_validation_failure_raises_400(self):
        # Missing required `actions` field
        with pytest.raises(HTTPException) as exc:
            EmblemService.validate_yaml("apiVersion: v1\nname: bad")
        assert exc.value.status_code == 400


# ---------------------------------------------------------------------------
# EmblemService.list_emblems
# ---------------------------------------------------------------------------

class TestListEmblems:
    def test_success(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response([SAMPLE_ROW])
        result = EmblemService.list_emblems(sb, None, 20, 0)
        assert len(result) == 1
        assert result[0].name == "my-api"

    def test_with_category_filter(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response([SAMPLE_ROW])
        result = EmblemService.list_emblems(sb, "finance", 10, 0)
        assert len(result) == 1

    def test_db_failure_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("DB down")
        with pytest.raises(HTTPException) as exc:
            EmblemService.list_emblems(sb, None, 20, 0)
        assert exc.value.status_code == 500
        assert exc.value.detail == "Internal server error"


# ---------------------------------------------------------------------------
# EmblemService.search_emblems — fallback path
# ---------------------------------------------------------------------------

class TestSearchEmblems:
    def test_ilike_fallback_on_rpc_failure(self):
        sb, q = make_supabase()
        # First execute (RPC) raises an error, second (ILIKE) succeeds.
        q.execute.side_effect = [Exception("RPC not found"), mock_response([SAMPLE_ROW])]
        result = EmblemService.search_emblems(sb, "my", None, "downloads", 10, 0)
        assert len(result) == 1

    def test_both_paths_fail_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("total failure")
        with pytest.raises(HTTPException) as exc:
            EmblemService.search_emblems(sb, "query", None, "downloads", 10, 0)
        assert exc.value.status_code == 500

    def test_sort_by_recent(self):
        sb, q = make_supabase()
        q.execute.side_effect = [Exception("no rpc"), mock_response([SAMPLE_ROW])]
        result = EmblemService.search_emblems(sb, "q", None, "recent", 5, 0)
        assert len(result) == 1

    def test_sort_by_name(self):
        sb, q = make_supabase()
        q.execute.side_effect = [Exception("no rpc"), mock_response([SAMPLE_ROW])]
        result = EmblemService.search_emblems(sb, "q", None, "name", 5, 0)
        assert len(result) == 1

    def test_metachar_escaping(self):
        """Verify LIKE metacharacters are escaped in fallback search."""
        sb, q = make_supabase()
        q.execute.side_effect = [Exception("no rpc"), mock_response([])]
        result = EmblemService.search_emblems(sb, "100%_off\\path", None, "downloads", 5, 0)
        assert result == []


# ---------------------------------------------------------------------------
# EmblemService.get_emblem
# ---------------------------------------------------------------------------

class TestGetEmblem:
    def test_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            EmblemService.get_emblem(sb, "missing")
        assert exc.value.status_code == 404

    def test_db_error_with_not_found_msg_raises_404(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("row not found")
        with pytest.raises(HTTPException) as exc:
            EmblemService.get_emblem(sb, "missing")
        assert exc.value.status_code == 404

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("connection error")
        with pytest.raises(HTTPException) as exc:
            EmblemService.get_emblem(sb, "any")
        assert exc.value.status_code == 500


# ---------------------------------------------------------------------------
# EmblemService.get_emblem_version
# ---------------------------------------------------------------------------

class TestGetEmblemVersion:
    def test_emblem_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            EmblemService.get_emblem_version(sb, "no-such", "1.0.0")
        assert exc.value.status_code == 404

    def test_version_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.side_effect = [
            mock_response({"id": "e-1"}),  # emblem found
            mock_response(None),            # version not found
        ]
        with pytest.raises(HTTPException) as exc:
            EmblemService.get_emblem_version(sb, "my-api", "9.9.9")
        assert exc.value.status_code == 404

    def test_success(self):
        sb, q = make_supabase()
        q.execute.side_effect = [
            mock_response({"id": "e-1"}),
            mock_response({
                "yaml_content": "...",
                "changelog": "initial",
                "published_at": "2024-01-01T00:00:00Z",
            }),
        ]
        result = EmblemService.get_emblem_version(sb, "my-api", "1.0.0")
        assert result["version"] == "1.0.0"
        assert "yaml_content" in result

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("boom")
        with pytest.raises(HTTPException) as exc:
            EmblemService.get_emblem_version(sb, "my-api", "1.0.0")
        assert exc.value.status_code == 500


# ---------------------------------------------------------------------------
# EmblemService.create_emblem
# ---------------------------------------------------------------------------

class TestCreateEmblem:
    def _body(self):
        return EmblemCreate(
            name="new-api",
            description="A brand new API for testing things",
            yaml_content=VALID_YAML,
            version="1.0.0",
        )

    def test_duplicate_name_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response([{"id": "existing"}])
        with pytest.raises(HTTPException) as exc:
            EmblemService.create_emblem(sb, self._body(), SAMPLE_USER)
        assert exc.value.status_code == 400
        assert "already exists" in exc.value.detail

    def test_insert_returns_no_data_raises_500(self):
        sb, q = make_supabase()
        # First call: no duplicates; second call: insert returns empty
        q.execute.side_effect = [
            mock_response([]),              # duplicate check → empty
            mock_response(None),            # insert → no data
        ]
        with pytest.raises(HTTPException) as exc:
            EmblemService.create_emblem(sb, self._body(), SAMPLE_USER)
        assert exc.value.status_code == 500

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("db error")
        with pytest.raises(HTTPException) as exc:
            EmblemService.create_emblem(sb, self._body(), SAMPLE_USER)
        assert exc.value.status_code == 500


# ---------------------------------------------------------------------------
# EmblemService.update_emblem
# ---------------------------------------------------------------------------

class TestUpdateEmblem:
    def _body(self):
        return EmblemUpdate(
            yaml_content=VALID_YAML,
            version="2.0.0",
            description="Updated description text here",
        )

    def test_emblem_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            EmblemService.update_emblem(sb, "missing", self._body(), SAMPLE_USER)
        assert exc.value.status_code == 404

    def test_not_author_raises_403(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(
            {**SAMPLE_ROW, "author_id": "other-user"}
        )
        with pytest.raises(HTTPException) as exc:
            EmblemService.update_emblem(sb, "my-api", self._body(), SAMPLE_USER)
        assert exc.value.status_code == 403

    def test_version_already_exists_raises_400(self):
        sb, q = make_supabase()
        q.execute.side_effect = [
            mock_response(SAMPLE_ROW),          # get emblem (author matches)
            mock_response([{"id": "v-1"}]),      # version check → exists
        ]
        body = EmblemUpdate(yaml_content=VALID_YAML, version="1.0.0")
        with pytest.raises(HTTPException) as exc:
            EmblemService.update_emblem(sb, "my-api", body, SAMPLE_USER)
        assert exc.value.status_code == 400

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("oops")
        with pytest.raises(HTTPException) as exc:
            EmblemService.update_emblem(sb, "my-api", self._body(), SAMPLE_USER)
        assert exc.value.status_code == 500


# ---------------------------------------------------------------------------
# EmblemService.delete_emblem
# ---------------------------------------------------------------------------

class TestDeleteEmblem:
    def test_emblem_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            EmblemService.delete_emblem(sb, "missing", SAMPLE_USER)
        assert exc.value.status_code == 404

    def test_not_author_raises_403(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(
            {**SAMPLE_ROW, "author_id": "other-user"}
        )
        with pytest.raises(HTTPException) as exc:
            EmblemService.delete_emblem(sb, "my-api", SAMPLE_USER)
        assert exc.value.status_code == 403

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("db down")
        with pytest.raises(HTTPException) as exc:
            EmblemService.delete_emblem(sb, "my-api", SAMPLE_USER)
        assert exc.value.status_code == 500


# ---------------------------------------------------------------------------
# KeyService
# ---------------------------------------------------------------------------

class TestKeyServiceListKeys:
    def test_success_empty(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response([])
        result = KeyService.list_keys(sb, "u-1")
        assert result == []

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("db fail")
        with pytest.raises(HTTPException) as exc:
            KeyService.list_keys(sb, "u-1")
        assert exc.value.status_code == 500


class TestKeyServiceCreateKey:
    def _body(self):
        return KeyCreate(name="my-key")

    def test_duplicate_name_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response([{"id": "k-existing"}])
        with pytest.raises(HTTPException) as exc:
            KeyService.create_key(sb, "u-1", self._body())
        assert exc.value.status_code == 400

    def test_insert_no_data_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = [
            mock_response([]),      # duplicate check → none
            mock_response(None),    # insert → no data
        ]
        with pytest.raises(HTTPException) as exc:
            KeyService.create_key(sb, "u-1", self._body())
        assert exc.value.status_code == 500

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = [
            mock_response([]),
            Exception("db error"),
        ]
        with pytest.raises(HTTPException) as exc:
            KeyService.create_key(sb, "u-1", self._body())
        assert exc.value.status_code == 500

    def test_with_expiration(self):
        sb, q = make_supabase()
        row = {
            "id": "k-new",
            "created_at": "2024-01-01T00:00:00Z",
            "expires_at": "2024-01-31T00:00:00Z",
        }
        q.execute.side_effect = [
            mock_response([]),
            mock_response([row]),
        ]
        result = KeyService.create_key(sb, "u-1", KeyCreate(name="temp", expires_days=30))
        assert result.expires_at is not None
        assert result.key.startswith("ely_")


class TestKeyServiceGetKey:
    def test_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            KeyService.get_key(sb, "u-1", "k-missing")
        assert exc.value.status_code == 404

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("db down")
        with pytest.raises(HTTPException) as exc:
            KeyService.get_key(sb, "u-1", "k-1")
        assert exc.value.status_code == 500


class TestKeyServiceDeleteKey:
    def test_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            KeyService.delete_key(sb, "u-1", "k-missing")
        assert exc.value.status_code == 404

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("db down")
        with pytest.raises(HTTPException) as exc:
            KeyService.delete_key(sb, "u-1", "k-1")
        assert exc.value.status_code == 500


# ---------------------------------------------------------------------------
# AuthService
# ---------------------------------------------------------------------------

class TestAuthServiceGetUser:
    def test_invalid_token_raises_401(self):
        sb, _ = make_supabase()
        sb.auth.get_user.side_effect = Exception("bad token")
        with pytest.raises(HTTPException) as exc:
            AuthService.get_user_from_token(sb, "bad-token")
        assert exc.value.status_code == 401

    def test_no_user_raises_401(self):
        sb, _ = make_supabase()
        resp = MagicMock()
        resp.user = None
        sb.auth.get_user.return_value = resp
        with pytest.raises(HTTPException) as exc:
            AuthService.get_user_from_token(sb, "token")
        assert exc.value.status_code == 401

    def test_optional_returns_none_on_failure(self):
        sb, _ = make_supabase()
        sb.auth.get_user.side_effect = Exception("bad")
        result = AuthService.get_user_from_token_optional(sb, "token")
        assert result is None

    def test_optional_returns_none_when_no_user(self):
        sb, _ = make_supabase()
        resp = MagicMock()
        resp.user = None
        sb.auth.get_user.return_value = resp
        result = AuthService.get_user_from_token_optional(sb, "token")
        assert result is None


class TestAuthServiceLogin:
    def test_invalid_credentials_raises_401(self):
        sb, _ = make_supabase()
        sb.auth.sign_in_with_password.side_effect = Exception("wrong password")
        with pytest.raises(HTTPException) as exc:
            AuthService.login(sb, "a@b.com", "wrong")
        assert exc.value.status_code == 401

    def test_no_session_raises_401(self):
        sb, _ = make_supabase()
        resp = MagicMock()
        resp.user = MagicMock()
        resp.session = None
        sb.auth.sign_in_with_password.return_value = resp
        with pytest.raises(HTTPException) as exc:
            AuthService.login(sb, "a@b.com", "pass")
        assert exc.value.status_code == 401


class TestAuthServiceRegister:
    def test_username_taken_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({"id": "existing"})
        with pytest.raises(HTTPException) as exc:
            AuthService.register(sb, "a@b.com", "pass1234", "alice", "https://fe.app")
        assert exc.value.status_code == 400
        assert "Username" in exc.value.detail

    def test_already_registered_email_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        sb.auth.sign_up.side_effect = Exception("User already registered")
        with pytest.raises(HTTPException) as exc:
            AuthService.register(sb, "a@b.com", "pass1234", "alice", "https://fe.app")
        assert exc.value.status_code == 400
        assert "Email already registered" in exc.value.detail

    def test_no_user_in_response_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        resp = MagicMock()
        resp.user = None
        sb.auth.sign_up.return_value = resp
        with pytest.raises(HTTPException) as exc:
            AuthService.register(sb, "a@b.com", "pass1234", "alice", "https://fe.app")
        assert exc.value.status_code == 400


class TestAuthServiceRefreshToken:
    def test_no_session_raises_401(self):
        sb, _ = make_supabase()
        resp = MagicMock()
        resp.session = None
        sb.auth.refresh_session.return_value = resp
        with pytest.raises(HTTPException) as exc:
            AuthService.refresh_token(sb, "bad-refresh")
        assert exc.value.status_code == 401

    def test_exception_raises_401(self):
        sb, _ = make_supabase()
        sb.auth.refresh_session.side_effect = Exception("expired")
        with pytest.raises(HTTPException) as exc:
            AuthService.refresh_token(sb, "refresh")
        assert exc.value.status_code == 401


class TestAuthServiceForgotPassword:
    def test_always_returns_safe_message(self):
        sb, _ = make_supabase()
        sb.auth.reset_password_for_email.side_effect = Exception("any error")
        result = AuthService.forgot_password(sb, "a@b.com", "https://fe.app")
        assert "reset instructions" in result["message"]

    def test_success(self):
        sb, _ = make_supabase()
        sb.auth.reset_password_for_email.return_value = None
        result = AuthService.forgot_password(sb, "a@b.com", "https://fe.app")
        assert "message" in result


class TestAuthServiceResetPassword:
    def test_invalid_token_raises_400(self):
        sb, _ = make_supabase()
        sb.auth.verify_oauth_token.side_effect = Exception("invalid token")
        with pytest.raises(HTTPException) as exc:
            AuthService.reset_password(sb, "bad-token", "newpass123")
        assert exc.value.status_code == 400

    def test_update_user_no_user_raises_400(self):
        sb, _ = make_supabase()
        sb.auth.verify_oauth_token.return_value = None
        resp = MagicMock()
        resp.user = None
        sb.auth.update_user.return_value = resp
        with pytest.raises(HTTPException) as exc:
            AuthService.reset_password(sb, "token", "newpass123")
        assert exc.value.status_code == 400


class TestAuthServiceGetProfile:
    def test_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            AuthService.get_profile(sb, "u-1", "a@b.com")
        assert exc.value.status_code == 404

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("boom")
        with pytest.raises(HTTPException) as exc:
            AuthService.get_profile(sb, "u-1", "a@b.com")
        assert exc.value.status_code == 500

    def test_success(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "username": "alice",
            "bio": None,
            "avatar_url": None,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z",
        })
        result = AuthService.get_profile(sb, "u-1", "a@b.com")
        assert result["username"] == "alice"


class TestAuthServiceUpdateProfile:
    def test_username_taken_raises_400(self):
        sb, q = make_supabase()
        # neq check returns an existing conflicting user
        q.execute.return_value = mock_response({"id": "other"})
        with pytest.raises(HTTPException) as exc:
            AuthService.update_profile(sb, "u-1", "a@b.com", "taken", None, None)
        assert exc.value.status_code == 400

    def test_no_profile_data_raises_404(self):
        sb, q = make_supabase()
        q.execute.side_effect = [
            mock_response(None),  # username check → no conflict
            mock_response(None),  # update → no data returned
        ]
        with pytest.raises(HTTPException) as exc:
            AuthService.update_profile(sb, "u-1", "a@b.com", "newname", None, None)
        assert exc.value.status_code == 404

    def test_no_changes_fetches_current_profile(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "username": "alice", "bio": None, "avatar_url": None,
            "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z",
        })
        result = AuthService.update_profile(sb, "u-1", "a@b.com", None, None, None)
        assert result["username"] == "alice"

    def test_list_profile_data_unwrapped(self):
        """Profile returned as a list (e.g. from update) is handled correctly."""
        sb, q = make_supabase()
        profile_row = {
            "username": "bob", "bio": "hi", "avatar_url": None,
            "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-02T00:00:00Z",
        }
        q.execute.side_effect = [
            mock_response(None),         # username conflict check → none
            mock_response([profile_row]),  # update returns list
        ]
        result = AuthService.update_profile(sb, "u-1", "a@b.com", "bob", None, None)
        assert result["username"] == "bob"

    def test_db_error_raises_500(self):
        sb, q = make_supabase()
        q.execute.side_effect = Exception("db fail")
        with pytest.raises(HTTPException) as exc:
            AuthService.update_profile(sb, "u-1", "a@b.com", "alice", None, None)
        assert exc.value.status_code == 500


class TestAuthServiceOAuth:
    def test_oauth_start_db_error_raises_500(self):
        sb, _ = make_supabase()
        sb.auth.sign_in_with_oauth.side_effect = Exception("provider down")
        with pytest.raises(HTTPException) as exc:
            AuthService.oauth_start(sb, "github", "https://app/cb", "https://fe.app", {})
        assert exc.value.status_code == 500

    def test_oauth_callback_error_param_redirects(self):
        sb, _ = make_supabase()
        states = {"state-abc": "https://app/cb"}
        result = AuthService.oauth_callback(
            sb, "github", "", "state-abc", "access_denied", "https://fe.app", states
        )
        assert result.status_code == 307
        assert "error=access_denied" in result.headers["location"]

    def test_oauth_callback_missing_code_raises_400(self):
        sb, _ = make_supabase()
        with pytest.raises(HTTPException) as exc:
            AuthService.oauth_callback(
                sb, "github", "", "", "", "https://fe.app", {}
            )
        assert exc.value.status_code == 400

    def test_oauth_callback_exchange_failure_redirects(self):
        sb, _ = make_supabase()
        sb.auth.exchange_code_for_session.side_effect = Exception("failed")
        states = {"st": "https://app/cb"}
        result = AuthService.oauth_callback(
            sb, "github", "code-xyz", "st", "", "https://fe.app", states
        )
        assert "error=oauth_failed" in result.headers["location"]


class TestAuthServiceDeviceCode:
    def test_create_device_code_returns_fields(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        result = AuthService.create_device_code(sb, "Elysium CLI", "https://fe.app")
        assert "device_code" in result
        assert "user_code" in result
        assert result["expires_in"] == 600

    def test_get_device_status_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            AuthService.get_device_status(sb, "ABCD-1234")
        assert exc.value.status_code == 404

    def test_get_device_status_success(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "user_code": "ABCD-1234",
            "verified_at": None,
            "client_name": "Elysium CLI",
            "expires_at": "2099-01-01T00:00:00Z",
        })
        result = AuthService.get_device_status(sb, "ABCD-1234")
        assert result["verified"] is False

    def test_verify_device_code_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            AuthService.verify_device_code(sb, "ABCD-1234", SAMPLE_USER)
        assert exc.value.status_code == 404

    def test_verify_device_code_already_verified_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "user_code": "ABCD-1234",
            "verified_at": "2024-01-01T00:00:00Z",
            "expires_at": "2099-01-01T00:00:00Z",
        })
        with pytest.raises(HTTPException) as exc:
            AuthService.verify_device_code(sb, "ABCD-1234", SAMPLE_USER)
        assert exc.value.status_code == 400

    def test_verify_device_code_expired_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "user_code": "ABCD-1234",
            "verified_at": None,
            "expires_at": "2000-01-01T00:00:00+00:00",  # past date
        })
        with pytest.raises(HTTPException) as exc:
            AuthService.verify_device_code(sb, "ABCD-1234", SAMPLE_USER)
        assert exc.value.status_code == 400

    def test_poll_device_token_not_found_raises_404(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response(None)
        with pytest.raises(HTTPException) as exc:
            AuthService.poll_device_token(sb, "dc-abc")
        assert exc.value.status_code == 404

    def test_poll_device_token_expired_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "device_code": "dc-abc",
            "expires_at": "2000-01-01T00:00:00+00:00",
        })
        with pytest.raises(HTTPException) as exc:
            AuthService.poll_device_token(sb, "dc-abc")
        assert exc.value.status_code == 400

    def test_poll_device_token_pending_raises_400(self):
        sb, q = make_supabase()
        q.execute.return_value = mock_response({
            "device_code": "dc-abc",
            "expires_at": "2099-01-01T00:00:00+00:00",
            "verified_at": None,
            "user_id": None,
        })
        with pytest.raises(HTTPException) as exc:
            AuthService.poll_device_token(sb, "dc-abc")
        assert exc.value.status_code == 400
        assert "pending" in exc.value.detail.lower()

    def test_generate_device_code_length(self):
        code = AuthService.generate_device_code()
        assert len(code) == 40

    def test_generate_user_code_format(self):
        code = AuthService.generate_user_code()
        assert len(code) == 9  # XXXX-XXXX
        assert code[4] == "-"


class TestAuthServiceLogout:
    def test_sign_out_error_raises_400(self):
        sb, _ = make_supabase()
        sb.auth.sign_out.side_effect = Exception("session gone")
        with pytest.raises(HTTPException) as exc:
            AuthService.logout(sb)
        assert exc.value.status_code == 400
