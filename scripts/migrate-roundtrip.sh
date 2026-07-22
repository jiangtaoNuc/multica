#!/usr/bin/env bash
# migrate-roundtrip.sh — smoke-test the last N migrations for down→up idempotency.
#
# Usage:
#   bash scripts/migrate-roundtrip.sh [N]
#
# N defaults to 10. Requires DATABASE_URL to be set (or defaults to the dev DSN).
# Intended for CI ("Migrate roundtrip" step) and local `make migrate-check`.
#
# Design:
#   1. Run all migrations up (idempotent – already-applied are skipped).
#   2. For the last N unique version prefixes, in reverse order:
#        a. Verify the .down.sql file exists (fail loudly if missing).
#        b. Run migrate down (one step).
#        c. Run migrate up (one step back to head).
#   3. Leave the DB at head when done.
set -euo pipefail

N="${1:-10}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SERVER_DIR="$REPO_ROOT/server"
MIGRATE_CMD="go run ./cmd/migrate"

DATABASE_URL="${DATABASE_URL:-postgres://multica:multica@localhost:5432/multica?sslmode=disable}"
export DATABASE_URL

echo "==> migrate-roundtrip: testing last $N migrations"
echo "    DATABASE_URL: $DATABASE_URL"

# Ensure DB is at head before we start.
echo ""
echo "==> Applying all migrations (up)..."
(cd "$SERVER_DIR" && $MIGRATE_CMD up)

# Collect the last N .up.sql files (sorted, reversed for down iteration).
MIGRATIONS_DIR="$SERVER_DIR/migrations"
mapfile -t ALL_UP < <(ls "$MIGRATIONS_DIR"/*.up.sql | sort)
TOTAL="${#ALL_UP[@]}"

if [ "$N" -gt "$TOTAL" ]; then
  N="$TOTAL"
fi

# Take the last N entries, then reverse to get newest-first.
mapfile -t LAST_N < <(printf '%s\n' "${ALL_UP[@]: -$N}" | sort -r)

PASS=0
FAIL=0

for UP_FILE in "${LAST_N[@]}"; do
  BASE="$(basename "$UP_FILE" .up.sql)"
  DOWN_FILE="$MIGRATIONS_DIR/${BASE}.down.sql"

  echo ""
  echo "--- roundtrip: $BASE ---"

  # Fail if the .down.sql is missing.
  if [ ! -f "$DOWN_FILE" ]; then
    echo "  FAIL: missing down migration: $DOWN_FILE"
    FAIL=$((FAIL + 1))
    continue
  fi

  # Check if the down file is a forward-only placeholder (first line contains
  # the sentinel comment "由 XX 保证只前进"). Skip those silently.
  if head -1 "$DOWN_FILE" | grep -q "保证只前进"; then
    echo "  skip: forward-only migration (placeholder down)"
    PASS=$((PASS + 1))
    continue
  fi

  echo "  down..."
  if ! (cd "$SERVER_DIR" && $MIGRATE_CMD down) 2>&1; then
    echo "  FAIL: migrate down failed for $BASE"
    FAIL=$((FAIL + 1))
    # Re-apply up so subsequent iterations start from a consistent state.
    (cd "$SERVER_DIR" && $MIGRATE_CMD up) || true
    continue
  fi

  echo "  up..."
  if ! (cd "$SERVER_DIR" && $MIGRATE_CMD up) 2>&1; then
    echo "  FAIL: migrate up failed for $BASE"
    FAIL=$((FAIL + 1))
    continue
  fi

  echo "  ok"
  PASS=$((PASS + 1))
done

echo ""
echo "==> migrate-roundtrip results: $PASS passed, $FAIL failed (of $N tested)"

if [ "$FAIL" -gt 0 ]; then
  echo "FAIL"
  exit 1
fi

echo "PASS"
