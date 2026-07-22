package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureServer starts an httptest.Server that replays canned responses for
// the CLI issue commands and asserts that inbound requests match the expected
// method/path/body shapes. It is intentionally strict: if a command starts
// hitting a new endpoint, the test fails rather than silently returning 404.
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()

	loadFixture := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("testdata", "cli_fixtures", name))
		if err != nil {
			t.Fatalf("load fixture %s: %v", name, err)
		}
		return data
	}

	issueGet := loadFixture("issue_get_response.json")
	issueList := loadFixture("issue_list_response.json")
	issueCreate := loadFixture("issue_create_response.json")
	commentAdd := loadFixture("comment_add_response.json")

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// issue ref resolution used by issue get and comment add.
		case r.Method == http.MethodGet && r.URL.Path == "/api/issues/MUL-123":
			json.NewEncoder(w).Encode(map[string]any{
				"id":         "11111111-1111-1111-1111-111111111111",
				"identifier": "MUL-123",
				"title":      "Fix login redirect",
			})

		// issue get MUL-123: fetch by resolved UUID.
		case r.Method == http.MethodGet && r.URL.Path == "/api/issues/11111111-1111-1111-1111-111111111111":
			w.Write(issueGet)

		// issue list
		case r.Method == http.MethodGet && r.URL.Path == "/api/issues":
			if got := r.URL.Query().Get("workspace_id"); got != "ws-fixture" {
				t.Errorf("list expected workspace_id=ws-fixture, got %q", got)
			}
			w.Write(issueList)

		// issue create
		case r.Method == http.MethodPost && r.URL.Path == "/api/issues":
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			if err := json.Unmarshal(body, &req); err != nil {
				t.Errorf("unmarshal create request: %v", err)
			}
			if req["title"] != "New feature request" {
				t.Errorf("create expected title %q, got %q", "New feature request", req["title"])
			}
			w.Write(issueCreate)

		// comment add MUL-123 --content "...": post the comment by resolved UUID.
		case r.Method == http.MethodPost && r.URL.Path == "/api/issues/11111111-1111-1111-1111-111111111111/comments":
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			if err := json.Unmarshal(body, &req); err != nil {
				t.Errorf("unmarshal comment request: %v", err)
			}
			if req["content"] != "Looking into this now." {
				t.Errorf("comment expected content %q, got %q", "Looking into this now.", req["content"])
			}
			w.Write(commentAdd)

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, `{"error":"unexpected fixture request"}`, http.StatusNotFound)
		}
	}))
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = outW
	os.Stderr = errW
	rootCmd.SetArgs(args)

	done := make(chan struct{})
	var outBuf, errBuf bytes.Buffer
	go func() {
		io.Copy(&outBuf, outR)
		io.Copy(&errBuf, errR)
		close(done)
	}()

	execErr := rootCmd.Execute()

	outW.Close()
	errW.Close()
	<-done
	outR.Close()
	errR.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr
	rootCmd.SetArgs(nil)

	return outBuf.String(), errBuf.String(), execErr
}

func TestCLIIssueGetFixture(t *testing.T) {
	srv := fixtureServer(t)
	defer srv.Close()

	stdout, stderr, err := runCLI(t,
		"issue", "get", "MUL-123",
		"--server-url", srv.URL,
		"--workspace-id", "ws-fixture",
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("issue get: %v\nstderr: %s", err, stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout)
	}
	if got["identifier"] != "MUL-123" {
		t.Errorf("expected identifier MUL-123, got %v", got["identifier"])
	}
	if got["title"] != "Fix login redirect" {
		t.Errorf("expected title 'Fix login redirect', got %v", got["title"])
	}
}

func TestCLIIssueListFixture(t *testing.T) {
	srv := fixtureServer(t)
	defer srv.Close()

	stdout, stderr, err := runCLI(t,
		"issue", "list",
		"--server-url", srv.URL,
		"--workspace-id", "ws-fixture",
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("issue list: %v\nstderr: %s", err, stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout)
	}
	issues, _ := got["issues"].([]any)
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestCLIIssueCreateFixture(t *testing.T) {
	srv := fixtureServer(t)
	defer srv.Close()

	stdout, stderr, err := runCLI(t,
		"issue", "create",
		"--title", "New feature request",
		"--description", "We need a dark mode toggle.",
		"--status", "todo",
		"--priority", "none",
		"--server-url", srv.URL,
		"--workspace-id", "ws-fixture",
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("issue create: %v\nstderr: %s", err, stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout)
	}
	if got["identifier"] != "MUL-125" {
		t.Errorf("expected identifier MUL-125, got %v", got["identifier"])
	}
}

func TestCLICommentAddFixture(t *testing.T) {
	srv := fixtureServer(t)
	defer srv.Close()

	stdout, stderr, err := runCLI(t,
		"issue", "comment", "add", "MUL-123",
		"--content", "Looking into this now.",
		"--server-url", srv.URL,
		"--workspace-id", "ws-fixture",
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("comment add: %v\nstderr: %s", err, stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout)
	}
	if got["content"] != "Looking into this now." {
		t.Errorf("expected content 'Looking into this now.', got %v", got["content"])
	}
	if !strings.Contains(stderr, "Comment added") {
		t.Errorf("expected stderr to contain 'Comment added', got %q", stderr)
	}
}

