# Contract Testing — OpenAPI Single Source of Truth

This document explains how the OpenAPI spec (`server/api/openapi.yaml`) serves
as the single source of truth for the Multica REST API, and what developers
must do when adding or modifying API handlers.

## Architecture

```
server/api/openapi.yaml          ← single source of truth (hand-edited)
        │
        ├──► openapi-typescript  → packages/core/api/generated.ts   (TS types)
        ├──► oapi-codegen        → server/api/openapi_types.gen.go  (Go types, planned)
        └──► redocly lint        → CI validation
```

The spec is **contract-first**: developers edit `openapi.yaml` directly, then
regenerate downstream artifacts. CI gates ensure the generated files never
drift from the spec.

## Current Coverage (Batch 1)

| Endpoint group | Routes |
|---|---|
| `/api/issues/*` | List, create, update, delete, search, grouped, batch, comments, timeline, subscribers, reactions, attachments, labels, metadata, pull-requests, tasks, rerun |
| `/api/comments/*` | Update, delete, resolve/unresolve, reactions |
| `/api/agents/*` | List, create, from-template, get, update, archive, restore, cancel-tasks, tasks, skills, env |
| `/api/squads/*` | List, create, get, update, delete, members, member status, member role |

## Developer Workflow

### Adding a new endpoint

1. **Edit the spec.** Open `server/api/openapi.yaml` and add the path item
   with method, parameters, request body, and response schemas. Use
   `$ref` to reference shared schemas under `components/schemas/`.

2. **Generate TS types.** Run:
   ```bash
   pnpm run generate:api
   ```
   This regenerates `packages/core/api/generated.ts`. Commit the updated file.

3. **Generate Go types (optional, for new handler files).** Run:
   ```bash
   cd server && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
     --config api/oapi-codegen.yaml api/openapi.yaml
   ```
   This regenerates `server/api/openapi_types.gen.go`. Commit the updated file.

4. **Implement the handler.** Add the Go handler in `server/internal/handler/`
   and register the route in `server/cmd/server/router.go`.

5. **Lint the spec.** Run:
   ```bash
   pnpm openapi-lint
   # or: make openapi-lint
   ```

6. **Commit everything together** — the spec change, generated files, and
   handler implementation.

### Modifying an existing endpoint

1. Update the relevant path item / schema in `openapi.yaml`.
2. Regenerate: `pnpm run generate:api`
3. Update the Go handler and frontend code as needed.
4. Commit spec + generated files + code changes together.

### CI Drift Gates

CI runs two checks on every PR:

| Step | Command | Fails when |
|---|---|---|
| `Lint OpenAPI spec` | `pnpm openapi-lint` | Spec has structural errors or Redocly rule violations |
| `Verify generated API types` | `pnpm run generate:api && git diff --exit-code` | `generated.ts` is out of sync with `openapi.yaml` |

If the drift gate fails, run `pnpm run generate:api` locally and commit the
updated `generated.ts`.

## Local Commands

```bash
pnpm run generate:api     # Regenerate TS types from spec
pnpm openapi-lint          # Lint spec with Redocly
make openapi-lint          # Same via Makefile
make openapi-drift         # Verify generated TS is in sync
```

## Key Files

| File | Purpose |
|---|---|
| `server/api/openapi.yaml` | OpenAPI 3.1 spec — the single source of truth |
| `server/api/oapi-codegen.yaml` | oapi-codegen config (Go type generation) |
| `packages/core/api/generated.ts` | Auto-generated TypeScript types (do not hand-edit) |
| `packages/core/api/contract.ts` | Convenience type aliases from generated types |
| `.github/workflows/ci.yml` | CI steps: `openapi-lint` + drift gate |

## Future Work

- Expand coverage to remaining endpoint groups (workspaces, projects, labels,
  runtimes, autopilots, chat, inbox, etc.).
- Integrate oapi-codegen generated Go types into handler signatures.
- Add contract conformance tests (e.g., `oapi-codegen` runtime validation)
  that verify handler responses match the spec at runtime.
