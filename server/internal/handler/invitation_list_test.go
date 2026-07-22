package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListWorkspaceInvitations(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/workspaces/"+testWorkspaceID+"/invitations", nil)
	req = withURLParams(req, "id", testWorkspaceID)
	testHandler.ListWorkspaceInvitations(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListWorkspaceInvitations: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []InvitationResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode invitations: %v", err)
	}
}

func TestListWorkspaceInvitations_InvalidWorkspace(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/workspaces/not-a-uuid/invitations", nil)
	req = withURLParams(req, "id", "not-a-uuid")
	testHandler.ListWorkspaceInvitations(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ListWorkspaceInvitations(bad ws): expected 400, got %d", w.Code)
	}
}
