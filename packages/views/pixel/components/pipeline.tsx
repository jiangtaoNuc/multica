"use client";

import { useState, useEffect } from "react";
import {
  HARNESS_STATES,
  STATE_LABELS,
  STATE_SHORT,
  type HarnessSnapshot,
  type HarnessState,
} from "./types";

const GLYPH_CHECK = "✓";
const GLYPH_DOT = "●";
const GLYPH_FIREWORK = "✦ ✧ ✦";
const GLYPH_ROCKET = "🚀";

function formatDuration(ms: number): string {
  if (ms <= 0) return "--";
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  const h = Math.floor(m / 60);
  if (h > 0) return `${h}h ${m % 60}m`;
  if (m > 0) return `${m}m ${s % 60}s`;
  return `${s}s`;
}

interface NodeCardProps {
  state: HarnessState;
  currentIndex: number;
  stateIndex: number;
  stayedMs: number;
  isFailed: boolean;
}

function NodeCard({ state, currentIndex, stateIndex, stayedMs, isFailed }: NodeCardProps) {
  const [showLightUp, setShowLightUp] = useState(false);
  const isCompleted = stateIndex < currentIndex;
  const isCurrent = stateIndex === currentIndex;

  useEffect(() => {
    if (!isCurrent) return;
    setShowLightUp(true);
    const t = setTimeout(() => setShowLightUp(false), 1000);
    return () => clearTimeout(t);
  }, [isCurrent]);

  const bg = isCompleted
    ? "var(--accent-lime)"
    : isCurrent
      ? "var(--accent-cyan)"
      : "var(--ink-muted)";

  const textColor = isCompleted || isCurrent ? "var(--bg-deep)" : "var(--text-dust)";

  const animClass = isCurrent ? "anim-heartbeat" : isFailed ? "anim-shake" : "";

  return (
    <div
      className={animClass}
      style={{
        width: "var(--node-size)",
        height: "var(--node-size)",
        background: bg,
        border: `4px solid ${
          isFailed
            ? "var(--accent-red)"
            : isCurrent
              ? "var(--accent-cyan)"
              : isCompleted
                ? "var(--accent-lime)"
                : "var(--ink-muted)"
        }`,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 2,
        position: "relative",
        animation: isCurrent
          ? "pixel-heartbeat 1.2s steps(8) infinite"
          : isFailed
            ? "pixel-fail-shake 0.4s steps(6) 3"
            : showLightUp
              ? "pixel-node-light-up 1s steps(12) 1"
              : "none",
        imageRendering: "pixelated",
        flexShrink: 0,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-heading)",
          fontSize: 8,
          color: textColor,
          textAlign: "center",
          lineHeight: 1.2,
        }}
      >
        {STATE_SHORT[state]}
      </span>
      {isCompleted && <span aria-hidden="true" style={{ fontSize: 16, color: "var(--bg-deep)" }}>{GLYPH_CHECK}</span>}
      {isCurrent && <span aria-hidden="true" style={{ fontSize: 12, color: "var(--bg-deep)" }}>{GLYPH_DOT}</span>}
      <span
        style={{
          fontFamily: "var(--font-body)",
          fontSize: 14,
          color: textColor,
          position: "absolute",
          bottom: -24,
          whiteSpace: "nowrap",
        }}
      >
        {formatDuration(isCurrent || isCompleted ? stayedMs : 0)}
      </span>
      <span
        style={{
          fontFamily: "var(--font-body)",
          fontSize: 13,
          color: "var(--text-dust)",
          position: "absolute",
          bottom: -42,
          whiteSpace: "nowrap",
        }}
      >
        {STATE_LABELS[state]}
      </span>
    </div>
  );
}

interface PipeProps {
  active: boolean;
}

function Pipe({ active }: PipeProps) {
  return (
    <div
      className={active ? "anim-dataflow" : ""}
      style={{
        flex: 1,
        minWidth: 40,
        height: "var(--pipe-height)",
        background: active
          ? "repeating-linear-gradient(90deg, var(--accent-cyan) 0, var(--accent-cyan) 8px, var(--bg-deep) 8px, var(--bg-deep) 16px)"
          : "var(--ink-muted)",
        animation: active ? "pixel-data-flow 0.6s steps(6) infinite" : "none",
        imageRendering: "pixelated",
        alignSelf: "center",
      }}
    />
  );
}

interface Props {
  snapshot: HarnessSnapshot;
  transition: { from: string; to: string } | null;
}

export function Pipeline({ snapshot, transition }: Props) {
  const currentIndex = HARNESS_STATES.indexOf(snapshot.state);
  const isDeployFailed = snapshot.meta.deployFailed;
  const isPrClosed = snapshot.meta.prClosed;

  return (
    <div style={{ position: "relative" }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          padding: "48px 24px 60px",
          justifyContent: "center",
          minWidth: 700,
        }}
      >
        {HARNESS_STATES.map((state, idx) => {
          const isFailed =
            (state === "pr_merged" && isDeployFailed) ||
            (state === "agent_picked_up" && isPrClosed);

          return (
            <span
              key={state}
              style={{ display: "flex", alignItems: "center", flex: idx < HARNESS_STATES.length - 1 ? 1 : "none" }}
            >
              <NodeCard
                state={state}
                currentIndex={currentIndex}
                stateIndex={idx}
                stayedMs={snapshot.perNode[state]?.stayedMs ?? 0}
                isFailed={isFailed}
              />
              {idx < HARNESS_STATES.length - 1 && <Pipe active={idx < currentIndex} />}
            </span>
          );
        })}
      </div>

      {transition?.to === "pr_merged" && (
        <div
          aria-hidden="true"
          className="anim-firework"
          style={{
            position: "absolute",
            top: 0,
            left: "65%",
            fontSize: 32,
            animation: "pixel-firework 2s steps(24) 1",
            pointerEvents: "none",
          }}
        >
          {GLYPH_FIREWORK}
        </div>
      )}

      {transition?.to === "deployed" && (
        <div
          aria-hidden="true"
          className="anim-rocket"
          style={{
            position: "absolute",
            bottom: 40,
            left: "65%",
            fontSize: 24,
            animation: "pixel-rocket-launch 2.5s steps(30) 1",
            pointerEvents: "none",
          }}
        >
          {GLYPH_ROCKET}
        </div>
      )}
    </div>
  );
}
