/**
 * Contract types derived from the OpenAPI spec (server/api/openapi.yaml).
 *
 * These types are the single source of truth for the API wire shapes.
 * Hand-written types in packages/core/types/ should be compatible with
 * these; the OpenAPI drift gate (CI) ensures the generated file stays
 * in sync with the spec.
 */
import type { components, operations } from "./generated";

// ── Schema types ──────────────────────────────────────────────────────
export type IssueResponse = components["schemas"]["IssueResponse"];
export type CreateIssueRequest = components["schemas"]["CreateIssueRequest"];
export type UpdateIssueRequest = components["schemas"]["UpdateIssueRequest"];
export type SearchIssueResponse = components["schemas"]["SearchIssueResponse"];
export type CommentResponse = components["schemas"]["CommentResponse"];
export type CreateCommentRequest = components["schemas"]["CreateCommentRequest"];
export type AgentResponse = components["schemas"]["AgentResponse"];
export type CreateAgentRequest = components["schemas"]["CreateAgentRequest"];
export type UpdateAgentRequest = components["schemas"]["UpdateAgentRequest"];
export type SquadResponse = components["schemas"]["SquadResponse"];
export type CreateSquadRequest = components["schemas"]["CreateSquadRequest"];
export type UpdateSquadRequest = components["schemas"]["UpdateSquadRequest"];
export type SquadMemberResponse = components["schemas"]["SquadMemberResponse"];
export type LabelResponse = components["schemas"]["LabelResponse"];
export type ReactionResponse = components["schemas"]["ReactionResponse"];

// ── Operation response helpers ────────────────────────────────────────
export type ListIssuesResponse =
  operations["listIssues"]["responses"]["200"]["content"]["application/json"];
export type GetIssueResponse =
  operations["getIssue"]["responses"]["200"]["content"]["application/json"];
export type ListAgentsResponse =
  operations["listAgents"]["responses"]["200"]["content"]["application/json"];
export type ListSquadsResponse =
  operations["listSquads"]["responses"]["200"]["content"]["application/json"];
export type ListCommentsResponse =
  operations["listComments"]["responses"]["200"]["content"]["application/json"];
