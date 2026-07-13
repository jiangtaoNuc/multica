# Branch Protection for `main`

This document lists the required status checks that must pass before a pull request can be merged into `main`, and the steps repository owners need to enable them in the GitHub UI.

## Required status checks

The following checks must be configured as **required status checks** for the `main` branch. Names must match the GitHub Actions job IDs exactly (matrix jobs render as `job-name (matrix-value)`).

| Check name | Source | What it guards |
|------------|--------|----------------|
| `frontend` | `.github/workflows/ci.yml` | Frontend build, typecheck, lint, and tests |
| `backend-lint` | `.github/workflows/ci.yml` | Go static analysis / lint |
| `backend-test` | `.github/workflows/ci.yml` | Go unit tests with race detector |
| `backend-contract` | `.github/workflows/ci.yml` | OpenAPI / handler contract validation |
| `installer (ubuntu-latest)` | `.github/workflows/ci.yml` | `scripts/install.test.sh` on Ubuntu |
| `installer (macos-latest)` | `.github/workflows/ci.yml` | `scripts/install.test.sh` on macOS |
| `openapi-drift` | `.github/workflows/ci.yml` | Generated OpenAPI / TS client drift check |
| `sqlc-drift` | `.github/workflows/ci.yml` | `sqlc generate` output drift check |
| `migrate-roundtrip` | `.github/workflows/ci.yml` | Recent migrations `down → up` idempotency smoke test |

> **Note:** Some of these jobs (for example `backend-lint`, `backend-test`, `backend-contract`, `openapi-drift`, `sqlc-drift`, and `migrate-roundtrip`) are introduced by [LOO-343](https://github.com/jiangtaoNuc/multica/issues?q=is%3Aissue+LOO-343) Stage 4a. They must exist in `.github/workflows/ci.yml` on `main` before the matching required checks can be selected in the GitHub UI.

## Owner setup steps

A repository admin must enable the checks in the branch protection rules. GitHub does not expose this setting via the REST API for fine-grained required-check selection, so the UI steps below are the canonical procedure.

1. Open the repository on GitHub: `https://github.com/jiangtaoNuc/multica`.
2. Go to **Settings → Branches**.
3. In **Branch protection rules**, click **Add rule** (or edit the existing `main` rule).
4. In **Branch name pattern**, enter `main`.
5. Enable:
   - **Require a pull request before merging**
   - **Require status checks to pass before merging**
6. Under **Status checks that are required**, search for and add each check from the table above:
   - `frontend`
   - `backend-lint`
   - `backend-test`
   - `backend-contract`
   - `installer (ubuntu-latest)`
   - `installer (macos-latest)`
   - `openapi-drift`
   - `sqlc-drift`
   - `migrate-roundtrip`
7. Save the rule.

### Verification

After enabling the checks, open a test pull request against `main` and confirm:

- All nine checks appear in the PR status area.
- The PR cannot be merged until every required check passes.
- A failing required check blocks the merge button.

## Related documentation

- [Repository `Makefile`](../Makefile) — local verification via `make check`
- [Pull request template](../.github/PULL_REQUEST_TEMPLATE.md) — contributor checklist
- [Contributing guidelines](../CONTRIBUTING.md)
