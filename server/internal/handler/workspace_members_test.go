package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListMembers(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/workspaces/"+testWorkspaceID+"/members", nil)
	req = withURLParams(req, "id", testWorkspaceID)
	testHandler.ListMembers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListMembers: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var members []MemberResponse
	if err := json.NewDecoder(w.Body).Decode(&members); err != nil {
		t.Fatalf("decode members: %v", err)
	}
	var found bool
	for _, m := range members {
		if m.UserID == testUserID && m.Role == "owner" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListMembers: did not find owner member for test user")
	}
}

func TestListMembersWithUser(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/workspaces/"+testWorkspaceID+"/members-with-user", nil)
	req = withURLParams(req, "id", testWorkspaceID)
	testHandler.ListMembersWithUser(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListMembersWithUser: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var members []MemberWithUserResponse
	if err := json.NewDecoder(w.Body).Decode(&members); err != nil {
		t.Fatalf("decode members: %v", err)
	}
	var found bool
	for _, m := range members {
		if m.UserID == testUserID {
			if m.Email != handlerTestEmail || m.Name != handlerTestName {
				t.Fatalf("ListMembersWithUser: unexpected user fields %+v", m)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("ListMembersWithUser: did not find test user")
	}
}

func TestListMembersWithUser_InvalidWorkspace(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/workspaces/not-a-uuid/members-with-user", nil)
	req = withURLParams(req, "id", "not-a-uuid")
	testHandler.ListMembersWithUser(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ListMembersWithUser(bad ws): expected 400, got %d", w.Code)
	}
}
