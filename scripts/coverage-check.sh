#!/usr/bin/env bash
# scripts/coverage-check.sh — enforce per-package coverage thresholds.
#
# Usage:
#   scripts/coverage-check.sh [path/to/coverage.out]
#
# Defaults to server/coverage.out (run from repo root).

set -euo pipefail

cd "$(dirname "$0")/.."

COVERAGE_FILE="${1:-server/coverage.out}"

if [ ! -f "$COVERAGE_FILE" ]; then
  echo "coverage-check: file not found: $COVERAGE_FILE" >&2
  exit 1
fi

MODULE="github.com/multica-ai/multica/server"

# Thresholds for core packages. The issue specified internal/api; this repo's
# API layer lives in internal/handler, so the gate is applied there.
declare -A THRESHOLD
THRESHOLD["$MODULE/pkg/taskfailure"]=80
THRESHOLD["$MODULE/pkg/agent"]=60
THRESHOLD["$MODULE/internal/handler"]=60

failed=0

for pkg in $(printf '%s\n' "${!THRESHOLD[@]}" | sort); do
  thr="${THRESHOLD[$pkg]}"
  suffix="${pkg#$MODULE/}"
  dir="server/$suffix"

  if [ ! -d "$dir" ]; then
    printf "MISSING  %s (threshold %d%%, directory %s not found)\n" "$pkg" "$thr" "$dir"
    continue
  fi

  # Aggregate statement coverage for this exact package.
  stats=$(awk -v pkg="$pkg" '
    /^mode:/ { next }
    {
      idx = index($0, ":")
      if (idx == 0) next
      path = substr($0, 1, idx - 1)
      sub(/\/[^\/]+$/, "", path)
      if (path != pkg) next
      rest = substr($0, idx + 1)
      split(rest, a, " ")
      stmts = a[2] + 0
      count = a[3] + 0
      total += stmts
      if (count > 0) covered += stmts
    }
    END { print covered + 0, total + 0 }
  ' "$COVERAGE_FILE")

  covered="${stats% *}"
  total="${stats#* }"

  if [ "$total" -eq 0 ]; then
    pct="0.0"
  else
    pct=$(awk -v c="$covered" -v t="$total" 'BEGIN { printf "%.1f", c / t * 100 }')
  fi

  if awk -v p="$pct" -v t="$thr" 'BEGIN { exit (p + 0 < t + 0) ? 0 : 1 }'; then
    printf "FAIL     %s coverage %s%% (threshold %d%%)\n" "$pkg" "$pct" "$thr"
    failed=1
  else
    printf "PASS     %s coverage %s%% (threshold %d%%)\n" "$pkg" "$pct" "$thr"
  fi
done

exit "$failed"
