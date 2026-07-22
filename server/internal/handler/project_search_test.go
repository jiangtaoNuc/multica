package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchProjects(t *testing.T) {
	var projectID string
	if err := testPool.QueryRow(context.Background(),
		`INSERT INTO project (workspace_id, title, description) VALUES ($1, 'Searchable Alpha Project', 'unique zebrafinch marker') RETURNING id`,
		testWorkspaceID).Scan(&projectID); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM project WHERE id = $1`, projectID)
	})

	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/projects/search?workspace_id="+testWorkspaceID+"&q=zebrafinch", nil)
	testHandler.SearchProjects(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("SearchProjects: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if _, ok := resp["projects"]; !ok {
		t.Fatalf("SearchProjects: response missing projects key: %s", w.Body.String())
	}
}

func TestSearchProjects_MissingQuery(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/projects/search?workspace_id="+testWorkspaceID, nil)
	testHandler.SearchProjects(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("SearchProjects(no q): expected 400, got %d", w.Code)
	}
}
