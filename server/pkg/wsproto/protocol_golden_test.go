package wsproto

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/multica-ai/multica/server/pkg/protocol"
)

var updateGolden = flag.Bool("update", false, "update golden snapshot files")

// goldenCase maps a snapshot file to the concrete payload type used for
// round-trip validation. The envelope type is read first, then the payload
// is decoded into the concrete value so field assertions are type-safe.
type goldenCase struct {
	file    string
	payload any
	assert  func(t *testing.T, p any)
}

func TestGoldenSnapshots(t *testing.T) {
	cases := []goldenCase{
		// daemon -> server
		{
			file:    "daemon_register.json",
			payload: &protocol.DaemonRegisterPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.DaemonRegisterPayload)
				if v.DaemonID == "" {
					t.Errorf("DaemonID must not be empty")
				}
				if len(v.Runtimes) == 0 {
					t.Errorf("Runtimes must not be empty")
				}
			},
		},
		{
			file:    "daemon_heartbeat.json",
			payload: &protocol.DaemonHeartbeatRequestPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.DaemonHeartbeatRequestPayload)
				if v.RuntimeID == "" {
					t.Errorf("RuntimeID must not be empty")
				}
			},
		},
		{
			file:    "task_progress.json",
			payload: &protocol.TaskProgressPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskProgressPayload)
				if v.TaskID == "" {
					t.Errorf("TaskID must not be empty")
				}
			},
		},
		{
			file:    "task_completed.json",
			payload: &protocol.TaskCompletedPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskCompletedPayload)
				if v.TaskID == "" {
					t.Errorf("TaskID must not be empty")
				}
			},
		},
		{
			file:    "task_message_text.json",
			payload: &protocol.TaskMessagePayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskMessagePayload)
				if v.TaskID == "" || v.Type != "text" {
					t.Errorf("expected text task message, got %+v", v)
				}
			},
		},
		{
			file:    "task_message_tool_use.json",
			payload: &protocol.TaskMessagePayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskMessagePayload)
				if v.TaskID == "" || v.Type != "tool_use" || v.Tool == "" {
					t.Errorf("expected tool_use task message, got %+v", v)
				}
			},
		},
		{
			file:    "task_message_tool_result.json",
			payload: &protocol.TaskMessagePayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskMessagePayload)
				if v.TaskID == "" || v.Type != "tool_result" || v.Tool == "" {
					t.Errorf("expected tool_result task message, got %+v", v)
				}
			},
		},
		// server -> daemon
		{
			file:    "task_dispatch.json",
			payload: &protocol.TaskDispatchPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskDispatchPayload)
				if v.TaskID == "" || v.IssueID == "" {
					t.Errorf("TaskID and IssueID must not be empty, got %+v", v)
				}
			},
		},
		{
			file:    "daemon_task_available.json",
			payload: &protocol.TaskAvailablePayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.TaskAvailablePayload)
				if v.RuntimeID == "" {
					t.Errorf("RuntimeID must not be empty")
				}
			},
		},
		{
			file:    "daemon_runtime_profiles_changed.json",
			payload: &protocol.RuntimeProfilesChangedPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.RuntimeProfilesChangedPayload)
				if v.WorkspaceID == "" {
					t.Errorf("WorkspaceID must not be empty")
				}
			},
		},
		{
			file:    "daemon_heartbeat_ack.json",
			payload: &protocol.DaemonHeartbeatAckPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.DaemonHeartbeatAckPayload)
				if v.RuntimeID == "" || v.Status == "" {
					t.Errorf("RuntimeID and Status must not be empty, got %+v", v)
				}
			},
		},
		{
			file:    "daemon_heartbeat_ack_update.json",
			payload: &protocol.DaemonHeartbeatAckPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.DaemonHeartbeatAckPayload)
				if v.RuntimeID == "" || v.PendingUpdate == nil {
					t.Errorf("expected pending update, got %+v", v)
				}
			},
		},
		{
			file:    "daemon_heartbeat_ack_local_skill_import.json",
			payload: &protocol.DaemonHeartbeatAckPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.DaemonHeartbeatAckPayload)
				if v.RuntimeID == "" || len(v.PendingLocalSkillImports) == 0 {
					t.Errorf("expected pending local skill imports, got %+v", v)
				}
			},
		},
		{
			file:    "daemon_heartbeat_ack_runtime_gone.json",
			payload: &protocol.DaemonHeartbeatAckPayload{},
			assert: func(t *testing.T, p any) {
				v := p.(*protocol.DaemonHeartbeatAckPayload)
				if v.RuntimeID == "" || !v.RuntimeGone || v.Status != protocol.HeartbeatStatusRuntimeGone {
					t.Errorf("expected runtime_gone ack, got %+v", v)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(strings.TrimSuffix(tc.file, ".json"), func(t *testing.T) {
			path := filepath.Join("testdata", tc.file)
			if *updateGolden {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Fatalf("golden file %s does not exist; create it manually or remove -update", path)
				}
			}

			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s: %v", path, err)
			}

			var msg protocol.Message
			if err := json.Unmarshal(raw, &msg); err != nil {
				t.Fatalf("unmarshal envelope %s: %v", path, err)
			}
			if msg.Type == "" {
				t.Errorf("golden %s missing envelope type", path)
			}

			if err := json.Unmarshal(msg.Payload, tc.payload); err != nil {
				t.Fatalf("unmarshal payload %s: %v", path, err)
			}
			if tc.assert != nil {
				tc.assert(t, tc.payload)
			}

			// Round-trip: re-serialize the envelope and compare to the original.
			// This catches accidental field additions, renames, or omitempty drift.
			repayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload: %v", err)
			}
			out := protocol.Message{Type: msg.Type, Payload: repayload}
			got, err := json.Marshal(out)
			if err != nil {
				t.Fatalf("marshal envelope: %v", err)
			}

			if *updateGolden {
				var buf bytes.Buffer
				if err := json.Indent(&buf, got, "", "  "); err != nil {
					t.Fatalf("indent updated golden: %v", err)
				}
				if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
					t.Fatalf("write updated golden %s: %v", path, err)
				}
				t.Logf("updated %s", path)
				return
			}

			want := normalizeJSON(t, raw)
			gotNorm := normalizeJSON(t, got)
			if want != gotNorm {
				t.Errorf("round-trip mismatch for %s\nwant: %s\ngot:  %s", path, want, gotNorm)
			}
		})
	}
}

func normalizeJSON(t *testing.T, data []byte) string {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("normalizeJSON unmarshal: %v", err)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("normalizeJSON marshal: %v", err)
	}
	return string(b)
}
