package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// seedPinTestIssue and seedPinTestProject create workspace-scoped entities the
// pin handlers can validate against.
func seedPinTestIssue(t *testing.T) string {
	t.Helper()

	var id string
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO issue (workspace_id, title, status, priority, creator_type, creator_id, number)
		VALUES ($1, 'pin test issue', 'todo', 'medium', 'member', $2,
		        COALESCE((SELECT MAX(number) FROM issue WHERE workspace_id = $1), 0) + 1)
		RETURNING id
	`, testWorkspaceID, testUserID).Scan(&id); err != nil {
		t.Fatalf("seed pin issue: %v", err)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM issue WHERE id = $1`, id)
	})
	return id
}

func seedPinTestProject(t *testing.T) string {
	t.Helper()

	var id string
	if err := testPool.QueryRow(context.Background(),
		`INSERT INTO project (workspace_id, title) VALUES ($1, 'pin test project') RETURNING id`,
		testWorkspaceID).Scan(&id); err != nil {
		t.Fatalf("seed pin project: %v", err)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM project WHERE id = $1`, id)
	})
	return id
}

func TestPinCreateListReorderDelete(t *testing.T) {
	issueID := seedPinTestIssue(t)
	projectID := seedPinTestProject(t)

	// Create an issue pin.
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/pins?workspace_id="+testWorkspaceID, CreatePinRequest{ItemType: "issue", ItemID: issueID})
	testHandler.CreatePin(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreatePin(issue): expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var issuePin PinnedItemResponse
	json.NewDecoder(w.Body).Decode(&issuePin)
	if issuePin.ItemType != "issue" || issuePin.ItemID != issueID {
		t.Fatalf("CreatePin(issue): unexpected payload %+v", issuePin)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM pinned_item WHERE workspace_id = $1 AND user_id = $2`, testWorkspaceID, testUserID)
	})

	// Create a project pin; its position should be strictly greater.
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/pins?workspace_id="+testWorkspaceID, CreatePinRequest{ItemType: "project", ItemID: projectID})
	testHandler.CreatePin(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreatePin(project): expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var projectPin PinnedItemResponse
	json.NewDecoder(w.Body).Decode(&projectPin)
	if projectPin.Position <= issuePin.Position {
		t.Fatalf("CreatePin: expected appended position > %v, got %v", issuePin.Position, projectPin.Position)
	}

	// Duplicate pin → 409.
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/pins?workspace_id="+testWorkspaceID, CreatePinRequest{ItemType: "issue", ItemID: issueID})
	testHandler.CreatePin(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("CreatePin(dup): expected 409, got %d: %s", w.Code, w.Body.String())
	}

	// List returns both pins.
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/pins?workspace_id="+testWorkspaceID, nil)
	testHandler.ListPins(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListPins: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var pins []PinnedItemResponse
	json.NewDecoder(w.Body).Decode(&pins)
	if len(pins) != 2 {
		t.Fatalf("ListPins: expected 2 pins, got %d", len(pins))
	}

	// Reorder: swap the two positions.
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/pins/reorder?workspace_id="+testWorkspaceID, ReorderPinsRequest{Items: []ReorderItem{
		{ID: issuePin.ID, Position: 10},
		{ID: projectPin.ID, Position: 5},
	}})
	testHandler.ReorderPins(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("ReorderPins: expected 204, got %d: %s", w.Code, w.Body.String())
	}
	var pos float64
	testPool.QueryRow(context.Background(), `SELECT position FROM pinned_item WHERE id = $1`, issuePin.ID).Scan(&pos)
	if pos != 10 {
		t.Fatalf("ReorderPins: expected issue pin position 10, got %v", pos)
	}

	// Delete the issue pin.
	w = httptest.NewRecorder()
	req = newRequest("DELETE", "/api/pins/issue/"+issueID+"?workspace_id="+testWorkspaceID, nil)
	req = withURLParams(req, "itemType", "issue", "itemId", issueID)
	testHandler.DeletePin(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DeletePin: expected 204, got %d: %s", w.Code, w.Body.String())
	}

	var remaining int
	testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM pinned_item WHERE workspace_id = $1 AND user_id = $2`,
		testWorkspaceID, testUserID).Scan(&remaining)
	if remaining != 1 {
		t.Fatalf("DeletePin: expected 1 remaining pin, got %d", remaining)
	}
}

func TestCreatePinValidation(t *testing.T) {
	// Invalid item_type.
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/pins?workspace_id="+testWorkspaceID, CreatePinRequest{ItemType: "bogus", ItemID: testWorkspaceID})
	testHandler.CreatePin(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("CreatePin(bad type): expected 400, got %d", w.Code)
	}

	// Missing item_id.
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/pins?workspace_id="+testWorkspaceID, CreatePinRequest{ItemType: "issue", ItemID: ""})
	testHandler.CreatePin(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("CreatePin(no id): expected 400, got %d", w.Code)
	}

	// Non-existent issue → 404.
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/pins?workspace_id="+testWorkspaceID, CreatePinRequest{
		ItemType: "issue",
		ItemID:   "00000000-0000-0000-0000-000000000000",
	})
	testHandler.CreatePin(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("CreatePin(missing issue): expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
