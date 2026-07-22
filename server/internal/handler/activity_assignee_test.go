package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAssigneeFrequency(t *testing.T) {
	// Seed an issue created by the test user, assigned to the test user
	// (member). CountCreatedIssueAssignees should then surface this pair.
	var issueID string
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO issue (workspace_id, title, status, priority, creator_type, creator_id, assignee_type, assignee_id, number)
		VALUES ($1, 'assignee freq issue', 'todo', 'medium', 'member', $2, 'member', $2,
		        COALESCE((SELECT MAX(number) FROM issue WHERE workspace_id = $1), 0) + 1)
		RETURNING id
	`, testWorkspaceID, testUserID).Scan(&issueID); err != nil {
		t.Fatalf("seed assignee issue: %v", err)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM issue WHERE id = $1`, issueID)
	})

	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/assignee-frequency?workspace_id="+testWorkspaceID, nil)
	testHandler.GetAssigneeFrequency(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetAssigneeFrequency: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var entries []AssigneeFrequencyEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode assignee frequency: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.AssigneeType == "member" && e.AssigneeID == testUserID && e.Frequency >= 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("GetAssigneeFrequency: expected member entry for test user, got %+v", entries)
	}
}
