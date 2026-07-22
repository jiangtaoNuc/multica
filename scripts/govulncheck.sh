#!/usr/bin/env bash
set -euo pipefail

# ==========================================================================
# Run govulncheck against the Go backend and enforce severity thresholds.
#
# HIGH / CRITICAL vulnerabilities fail the build.
# LOW / MODERATE vulnerabilities are printed as warnings but do not fail.
# ==========================================================================

GOVULNCHECK_VERSION="${GOVULNCHECK_VERSION:-v1.5.0}"
SERVER_DIR="${SERVER_DIR:-server}"
cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1 && [ -x /usr/local/go/bin/go ]; then
  export PATH="/usr/local/go/bin:$PATH"
fi

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

OUTFILE="$TMPDIR/govulncheck.json"

set +e
(cd "$SERVER_DIR" && go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" -format json ./...) > "$OUTFILE" 2>"$TMPDIR/stderr"
STATUS=$?
set -e

# Echo the stderr (human-readable progress) so it still appears in logs.
if [ -s "$TMPDIR/stderr" ]; then
  cat "$TMPDIR/stderr"
fi

# govulncheck exits 1 when it finds vulnerabilities and 0 when clean.
# Any other exit code is treated as a tooling error.
if [ "$STATUS" -ne 0 ] && [ "$STATUS" -ne 1 ]; then
  echo "govulncheck failed with exit code $STATUS"
  exit "$STATUS"
fi

# Parse JSON output for severities. Python is used because it is available on
# local developer machines and GitHub Actions runners without extra installs.
SUMMARY="$(python3 - "$OUTFILE" <<'PY'
import json
import sys

path = sys.argv[1]
counts = {"CRITICAL": 0, "HIGH": 0, "MODERATE": 0, "LOW": 0, "vulnerabilities": 0, "valid_lines": 0}
valid_lines = 0
osv_severity_by_id = {}
findings_by_id = {}


def iter_json_objects(text):
    # govulncheck -format json emits a stream of pretty-printed (multi-line)
    # JSON objects concatenated together, not one object per line. Walk the
    # stream with raw_decode so multi-line objects are decoded whole.
    decoder = json.JSONDecoder()
    idx = 0
    length = len(text)
    while idx < length:
        while idx < length and text[idx].isspace():
            idx += 1
        if idx >= length:
            break
        try:
            obj, end = decoder.raw_decode(text, idx)
        except json.JSONDecodeError:
            break
        idx = end
        yield obj


try:
    with open(path, "r", encoding="utf-8") as fh:
        content = fh.read()

    for obj in iter_json_objects(content):
        # Skip any top-level value that is not a JSON object (e.g. a bare
        # array element string), which is not a govulncheck event.
        if not isinstance(obj, dict):
            continue

        valid_lines += 1

        finding = obj.get("finding")
        if finding:
            osv_id = finding.get("osv")
            if osv_id:
                findings_by_id.setdefault(osv_id, []).append(finding.get("severity"))
            else:
                counts["vulnerabilities"] += 1
                sev = finding.get("severity")
                if sev in counts:
                    counts[sev] += 1
            continue

        osv = obj.get("osv")
        if osv:
            osv_id = osv.get("id")
            if not osv_id or osv_id not in osv_severity_by_id:
                max_score = None
                for entry in osv.get("severity", []):
                    score = entry.get("score")
                    if score is None:
                        continue
                    try:
                        score = float(score)
                    except (TypeError, ValueError):
                        continue
                    if max_score is None or score > max_score:
                        max_score = score
                if osv_id:
                    osv_severity_by_id[osv_id] = max_score
except FileNotFoundError:
    pass

# Prefer explicit finding severity; fall back to the highest CVSS score for the OSV.
for osv_id, severities in findings_by_id.items():
    counts["vulnerabilities"] += len(severities)
    for sev in severities:
        if sev in counts:
            counts[sev] += 1
        else:
            # No explicit severity: fall back to the OSV CVSS score for this ID.
            score = osv_severity_by_id.get(osv_id)
            if score is None:
                continue
            if score >= 9.0:
                counts["CRITICAL"] += 1
            elif score >= 7.0:
                counts["HIGH"] += 1
            elif score >= 4.0:
                counts["MODERATE"] += 1
            else:
                counts["LOW"] += 1

# Count OSVs that had no finding entry at all.
counted_ids = set(findings_by_id.keys())
for osv_id, score in osv_severity_by_id.items():
    if osv_id in counted_ids:
        continue
    if score is None:
        continue
    if score >= 9.0:
        counts["CRITICAL"] += 1
    elif score >= 7.0:
        counts["HIGH"] += 1
    elif score >= 4.0:
        counts["MODERATE"] += 1
    else:
        counts["LOW"] += 1

counts["valid_lines"] = valid_lines
print(f"{counts['CRITICAL']} {counts['HIGH']} {counts['MODERATE']} {counts['LOW']} {counts['vulnerabilities']} {counts['valid_lines']}")
PY
)"

read -r CRITICAL_COUNT HIGH_COUNT MODERATE_COUNT LOW_COUNT VULN_COUNT VALID_LINES <<< "$SUMMARY"

# If govulncheck exited non-zero but produced no parseable JSON, it is a
# tooling/network error rather than a vulnerability report.
if [ "$VALID_LINES" -eq 0 ] && [ "$STATUS" -ne 0 ]; then
  echo "govulncheck failed with exit code $STATUS and produced no parseable output"
  exit "$STATUS"
fi

if [ "$VULN_COUNT" -eq 0 ] && [ "$STATUS" -eq 0 ]; then
  echo "✓ No vulnerabilities found by govulncheck."
  exit 0
fi

echo ""
echo "govulncheck found vulnerabilities:"
printf "  CRITICAL: %s\n" "$CRITICAL_COUNT"
printf "  HIGH:     %s\n" "$HIGH_COUNT"
printf "  MODERATE: %s\n" "$MODERATE_COUNT"
printf "  LOW:      %s\n" "$LOW_COUNT"
echo ""

if [ "$CRITICAL_COUNT" -gt 0 ] || [ "$HIGH_COUNT" -gt 0 ]; then
  echo "✗ HIGH/CRITICAL vulnerabilities detected; failing build."
  exit 1
fi

echo "⚠ LOW/MODERATE vulnerabilities detected; treated as warnings only."
exit 0
