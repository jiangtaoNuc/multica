package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWorkspaceAgentRunCounts(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/agents/run-counts?workspace_id="+testWorkspaceID, nil)
	testHandler.GetWorkspaceAgentRunCounts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetWorkspaceAgentRunCounts: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []AgentRunCount
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode run counts: %v", err)
	}
}

func TestGetWorkspaceAgentActivity30d(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/agents/activity-30d?workspace_id="+testWorkspaceID, nil)
	testHandler.GetWorkspaceAgentActivity30d(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetWorkspaceAgentActivity30d: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []AgentActivityBucket
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode activity: %v", err)
	}
}
