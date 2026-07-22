package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/multica-ai/multica/server/internal/middleware"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// withInboxWorkspaceCtx injects the workspace+member context the real chi
// middleware chain sets. The inbox handlers read the workspace via
// ctxWorkspaceID (context only), so calling them directly without this
// yields "invalid workspace id".
func withInboxWorkspaceCtx(t *testing.T, req *http.Request) *http.Request {
	t.Helper()
	memberRow, err := testHandler.Queries.GetMemberByUserAndWorkspace(context.Background(), db.GetMemberByUserAndWorkspaceParams{
		UserID:      util.MustParseUUID(testUserID),
		WorkspaceID: util.MustParseUUID(testWorkspaceID),
	})
	if err != nil {
		t.Fatalf("load test member row: %v", err)
	}
	return req.WithContext(middleware.SetMemberContext(req.Context(), testWorkspaceID, memberRow))
}

// seedInboxItem inserts an inbox_item for the handler test user and returns its
// id. issueID may be "" to leave issue_id NULL. The row is removed on cleanup.
func seedInboxItem(t *testing.T, itemType string, read, archived bool, issueID string) string {
	t.Helper()

	var issueArg any
	if issueID == "" {
		issueArg = nil
	} else {
		issueArg = issueID
	}

	var id string
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO inbox_item (workspace_id, recipient_type, recipient_id, type, severity, issue_id, title, body, read, archived)
		VALUES ($1, 'member', $2, $3, 'info', $4, 'Inbox test item', 'body text', $5, $6)
		RETURNING id
	`, testWorkspaceID, testUserID, itemType, issueArg, read, archived).Scan(&id); err != nil {
		t.Fatalf("seed inbox item: %v", err)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM inbox_item WHERE id = $1`, id)
	})
	return id
}

// seedInboxTestIssue creates a done issue owned by the test user for
// issue-scoped inbox assertions.
func seedInboxTestIssue(t *testing.T, status string) string {
	t.Helper()

	var id string
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO issue (workspace_id, title, status, priority, creator_type, creator_id, number)
		VALUES ($1, 'inbox test issue', $2, 'medium', 'member', $3,
		        COALESCE((SELECT MAX(number) FROM issue WHERE workspace_id = $1), 0) + 1)
		RETURNING id
	`, testWorkspaceID, status, testUserID).Scan(&id); err != nil {
		t.Fatalf("seed inbox issue: %v", err)
	}
	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM issue WHERE id = $1`, id)
	})
	return id
}

func TestListInboxAndCountUnread(t *testing.T) {
	issueID := seedInboxTestIssue(t, "in_progress")
	unreadID := seedInboxItem(t, "mention", false, false, issueID)
	seedInboxItem(t, "assignment", true, false, "")

	// List returns the un-archived items, enriched with issue status.
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/inbox?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.ListInbox(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListInbox: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var items []InboxItemResponse
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("decode inbox list: %v", err)
	}
	var found *InboxItemResponse
	for i := range items {
		if items[i].ID == unreadID {
			found = &items[i]
		}
	}
	if found == nil {
		t.Fatalf("ListInbox did not return seeded unread item")
	}
	if found.IssueStatus == nil || *found.IssueStatus != "in_progress" {
		t.Fatalf("ListInbox: expected issue_status in_progress, got %v", found.IssueStatus)
	}

	// CountUnread reflects the unread item.
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/inbox/unread-count?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.CountUnreadInbox(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("CountUnreadInbox: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var countResp map[string]int64
	json.NewDecoder(w.Body).Decode(&countResp)
	if countResp["count"] < 1 {
		t.Fatalf("CountUnreadInbox: expected >=1 unread, got %d", countResp["count"])
	}
}

func TestMarkInboxRead(t *testing.T) {
	id := seedInboxItem(t, "mention", false, false, "")

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/inbox/"+id+"/read?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	req = withURLParam(req, "id", id)
	testHandler.MarkInboxRead(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("MarkInboxRead: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp InboxItemResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Read {
		t.Fatalf("MarkInboxRead: expected read=true, got %+v", resp)
	}
}

func TestArchiveInboxItemAndSiblings(t *testing.T) {
	issueID := seedInboxTestIssue(t, "done")
	target := seedInboxItem(t, "mention", false, false, issueID)
	sibling := seedInboxItem(t, "assignment", false, false, issueID)

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/inbox/"+target+"/archive?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	req = withURLParam(req, "id", target)
	testHandler.ArchiveInboxItem(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ArchiveInboxItem: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp InboxItemResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Archived {
		t.Fatalf("ArchiveInboxItem: expected archived=true, got %+v", resp)
	}

	// The sibling sharing the same issue must also be archived.
	var siblingArchived bool
	if err := testPool.QueryRow(context.Background(),
		`SELECT archived FROM inbox_item WHERE id = $1`, sibling).Scan(&siblingArchived); err != nil {
		t.Fatalf("load sibling: %v", err)
	}
	if !siblingArchived {
		t.Fatalf("ArchiveInboxItem: expected sibling on same issue to be archived")
	}
}

func TestMarkAllInboxRead(t *testing.T) {
	seedInboxItem(t, "mention", false, false, "")
	seedInboxItem(t, "assignment", false, false, "")

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/inbox/read-all?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.MarkAllInboxRead(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("MarkAllInboxRead: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// No unread items should remain for this user in the workspace.
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/inbox/unread-count?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.CountUnreadInbox(w, req)
	var countResp map[string]int64
	json.NewDecoder(w.Body).Decode(&countResp)
	if countResp["count"] != 0 {
		t.Fatalf("MarkAllInboxRead: expected 0 unread after mark-all, got %d", countResp["count"])
	}
}

func TestArchiveAllInbox(t *testing.T) {
	seedInboxItem(t, "mention", false, false, "")
	seedInboxItem(t, "assignment", true, false, "")

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/inbox/archive-all?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.ArchiveAllInbox(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ArchiveAllInbox: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Nothing should be returned by List (all archived).
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/inbox?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.ListInbox(w, req)
	var items []InboxItemResponse
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 0 {
		t.Fatalf("ArchiveAllInbox: expected empty list, got %d items", len(items))
	}
}

func TestArchiveAllReadInbox(t *testing.T) {
	readID := seedInboxItem(t, "mention", true, false, "")
	unreadID := seedInboxItem(t, "assignment", false, false, "")

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/inbox/archive-read?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.ArchiveAllReadInbox(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ArchiveAllReadInbox: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var readArchived, unreadArchived bool
	testPool.QueryRow(context.Background(), `SELECT archived FROM inbox_item WHERE id = $1`, readID).Scan(&readArchived)
	testPool.QueryRow(context.Background(), `SELECT archived FROM inbox_item WHERE id = $1`, unreadID).Scan(&unreadArchived)
	if !readArchived {
		t.Fatalf("ArchiveAllReadInbox: expected read item archived")
	}
	if unreadArchived {
		t.Fatalf("ArchiveAllReadInbox: expected unread item untouched")
	}
}

func TestArchiveCompletedInbox(t *testing.T) {
	doneIssue := seedInboxTestIssue(t, "done")
	openIssue := seedInboxTestIssue(t, "in_progress")
	doneItem := seedInboxItem(t, "mention", false, false, doneIssue)
	openItem := seedInboxItem(t, "mention", false, false, openIssue)

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/inbox/archive-completed?workspace_id="+testWorkspaceID, nil)
	req = withInboxWorkspaceCtx(t, req)
	testHandler.ArchiveCompletedInbox(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ArchiveCompletedInbox: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var doneArchived, openArchived bool
	testPool.QueryRow(context.Background(), `SELECT archived FROM inbox_item WHERE id = $1`, doneItem).Scan(&doneArchived)
	testPool.QueryRow(context.Background(), `SELECT archived FROM inbox_item WHERE id = $1`, openItem).Scan(&openArchived)
	if !doneArchived {
		t.Fatalf("ArchiveCompletedInbox: expected item on done issue archived")
	}
	if openArchived {
		t.Fatalf("ArchiveCompletedInbox: expected item on open issue untouched")
	}
}
