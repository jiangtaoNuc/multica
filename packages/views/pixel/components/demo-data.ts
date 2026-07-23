/**
 * Built-in demo snapshots for the pixel-art pipeline showcase.
 *
 * The original CODING HARNESS polled a Fastify BFF that derived these FSM
 * snapshots from the Multica CLI + GitHub API. To guarantee the page always
 * displays and interacts correctly inside Multica (no separate BFF
 * required), each snapshot here mirrors the BFF's output shape for a
 * representative issue at each lifecycle state. Switching tabs exercises the
 * heartbeat / firework / rocket / fail-shake animations exactly as the live
 * viz did.
 */
import { HARNESS_STATES, type HarnessSnapshot, type HarnessState, type NodeStatus } from "./types";

const NOW = "2026-07-23T02:42:00Z";

function makePerNode(currentIndex: number, durations: number[]): Record<HarnessState, NodeStatus> {
  return HARNESS_STATES.reduce((acc, state, idx) => {
    const entered = idx <= currentIndex;
    acc[state] = {
      state,
      enteredAt: entered ? "2026-07-23T02:00:00Z" : null,
      leftAt: idx < currentIndex ? "2026-07-23T02:05:00Z" : null,
      stayedMs: idx < currentIndex ? (durations[idx] ?? 0) : idx === currentIndex ? durations[idx] ?? 0 : 0,
    };
    return acc;
  }, {} as Record<HarnessState, NodeStatus>);
}

const baseDurations = [42_000, 138_000, 905_000, 1_200_000, 300_000, 75_000];

export const DEMO_SNAPSHOTS: HarnessSnapshot[] = [
  {
    issueId: "issue-601",
    identifier: "LOO-601",
    title: "Add login redirect guard",
    state: "issue_created",
    enteredAt: NOW,
    stayedMs: 42_000,
    perNode: makePerNode(0, baseDurations),
    meta: {
      prUrl: null,
      deployUrl: null,
      assignee: null,
      lastComment: "Issue triaged, awaiting agent pickup.",
      ciStatus: null,
      prDraft: false,
      prMerged: false,
      prClosed: false,
      deployFailed: false,
    },
    degraded: false,
    etag: "demo-601",
  },
  {
    issueId: "issue-602",
    identifier: "LOO-602",
    title: "Wire sidebar usage badge",
    state: "agent_picked_up",
    enteredAt: NOW,
    stayedMs: 138_000,
    perNode: makePerNode(1, baseDurations),
    meta: {
      prUrl: null,
      deployUrl: null,
      assignee: "Qoder-全栈开发工程师",
      lastComment: "On it — checking out the repo now.",
      ciStatus: null,
      prDraft: false,
      prMerged: false,
      prClosed: false,
      deployFailed: false,
    },
    degraded: false,
    etag: "demo-602",
  },
  {
    issueId: "issue-603",
    identifier: "LOO-603",
    title: "Pixel-art left menu entry",
    state: "coding",
    enteredAt: NOW,
    stayedMs: 905_000,
    perNode: makePerNode(2, baseDurations),
    meta: {
      prUrl: "https://github.com/jiangtaoNuc/multica/pull/88",
      deployUrl: null,
      assignee: "Qoder-全栈开发工程师",
      lastComment: "Opened draft PR, still wiring the route.",
      ciStatus: "pending",
      prDraft: true,
      prMerged: false,
      prClosed: false,
      deployFailed: false,
    },
    degraded: false,
    etag: "demo-603",
  },
  {
    issueId: "issue-604",
    identifier: "LOO-604",
    title: "Migrate pixel design tokens",
    state: "pr_opened",
    enteredAt: NOW,
    stayedMs: 1_200_000,
    perNode: makePerNode(3, baseDurations),
    meta: {
      prUrl: "https://github.com/jiangtaoNuc/multica/pull/89",
      deployUrl: null,
      assignee: "Qoder-全栈开发工程师",
      lastComment: "PR ready for review, CI is green.",
      ciStatus: "pass",
      prDraft: false,
      prMerged: false,
      prClosed: false,
      deployFailed: false,
    },
    degraded: false,
    etag: "demo-604",
  },
  {
    issueId: "issue-605",
    identifier: "LOO-605",
    title: "Merge pixel CSS keyframes",
    state: "pr_merged",
    enteredAt: NOW,
    stayedMs: 300_000,
    perNode: makePerNode(4, baseDurations),
    meta: {
      prUrl: "https://github.com/jiangtaoNuc/multica/pull/89",
      deployUrl: null,
      assignee: "Qoder-全栈开发工程师",
      lastComment: "Merged — deploy pending.",
      ciStatus: "pass",
      prDraft: false,
      prMerged: true,
      prClosed: false,
      deployFailed: false,
    },
    degraded: false,
    etag: "demo-605",
  },
  {
    issueId: "issue-606",
    identifier: "LOO-606",
    title: "Ship pixel pipeline page",
    state: "deployed",
    enteredAt: NOW,
    stayedMs: 75_000,
    perNode: makePerNode(5, baseDurations),
    meta: {
      prUrl: "https://github.com/jiangtaoNuc/multica/pull/89",
      deployUrl: "https://github.com/jiangtaoNuc/multica/actions/runs/9876543",
      assignee: "Qoder-全栈开发工程师",
      lastComment: "Deployed successfully.",
      ciStatus: "pass",
      prDraft: false,
      prMerged: true,
      prClosed: false,
      deployFailed: false,
    },
    degraded: false,
    etag: "demo-606",
  },
  {
    issueId: "issue-607",
    identifier: "LOO-607",
    title: "Deploy pixel page (failed)",
    state: "pr_merged",
    enteredAt: NOW,
    stayedMs: 300_000,
    perNode: makePerNode(4, baseDurations),
    meta: {
      prUrl: "https://github.com/jiangtaoNuc/multica/pull/90",
      deployUrl: null,
      assignee: "Qoder-全栈开发工程师",
      lastComment: "Merged but deploy workflow failed — retrying.",
      ciStatus: "pass",
      prDraft: false,
      prMerged: true,
      prClosed: false,
      deployFailed: true,
    },
    degraded: true,
    etag: "demo-607",
  },
];
