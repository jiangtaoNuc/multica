"use client";

import { useCallback, useEffect, useState } from "react";
import { DEMO_SNAPSHOTS } from "./demo-data";
import type { HarnessSnapshot, IssueSummary } from "./types";

const DEMO_ISSUES: IssueSummary[] = DEMO_SNAPSHOTS.map((s) => ({
  id: s.issueId,
  identifier: s.identifier,
  title: s.title,
  status: s.state,
  updatedAt: s.enteredAt ?? "",
}));

/**
 * Self-contained stand-in for the CODING HARNESS BFF polling hooks
 * (useIssues / useHarness). Reads from built-in demo snapshots so the pixel
 * pipeline always renders inside Multica without a separate BFF. The
 * transition tracking mirrors the live hook: when the selected issue's FSM
 * state changes, a {from, to} transition is emitted for 3s — driving the
 * firework (S5) and rocket (S6) animations.
 */
const FIRST_SNAPSHOT = DEMO_SNAPSHOTS[0]!;

export function usePixelDemo() {
  const [selectedId, setSelectedId] = useState<string>(FIRST_SNAPSHOT.issueId);
  const [transition, setTransition] = useState<{ from: string; to: string } | null>(null);

  const snapshot: HarnessSnapshot =
    DEMO_SNAPSHOTS.find((s) => s.issueId === selectedId) ?? FIRST_SNAPSHOT;

  const handleSelect = useCallback(
    (id: string) => {
      setSelectedId((prevId) => {
        const prev = DEMO_SNAPSHOTS.find((s) => s.issueId === prevId);
        const next = DEMO_SNAPSHOTS.find((s) => s.issueId === id);
        if (prev && next && prev.state !== next.state) {
          setTransition({ from: prev.state, to: next.state });
        }
        return id;
      });
    },
    [],
  );

  useEffect(() => {
    if (!transition) return;
    const t = setTimeout(() => setTransition(null), 3000);
    return () => clearTimeout(t);
  }, [transition]);

  return {
    issues: DEMO_ISSUES,
    selectedId,
    snapshot,
    transition,
    onSelect: handleSelect,
  };
}
