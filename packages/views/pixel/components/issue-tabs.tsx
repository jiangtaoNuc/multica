"use client";

import { useT } from "../../i18n";
import type { IssueSummary } from "./types";

const SHADE_GLYPH = "▒▒▒";
const BRACKET_OPEN = "[";
const BRACKET_CLOSE = "]";

interface Props {
  issues: IssueSummary[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export function IssueTabs({ issues, selectedId, onSelect }: Props) {
  const { t } = useT("pixel");

  if (issues.length === 0) {
    return (
      <div
        style={{
          padding: "12px 16px",
          fontFamily: "var(--font-body)",
          fontSize: 24,
          color: "var(--text-dust)",
          textAlign: "center",
        }}
      >
        <span aria-hidden="true">{SHADE_GLYPH}</span>
        {" "}
        {t(($) => $.no_issues)}
        {" "}
        <span aria-hidden="true">{SHADE_GLYPH}</span>
      </div>
    );
  }

  return (
    <div
      style={{
        display: "flex",
        gap: 8,
        padding: "8px 16px",
        overflowX: "auto",
        borderBottom: "4px solid var(--ink-muted)",
      }}
    >
      {issues.map((issue) => {
        const isSelected = issue.id === selectedId;
        return (
          <button
            key={issue.id}
            type="button"
            onClick={() => onSelect(issue.id)}
            style={{
              fontFamily: "var(--font-heading)",
              fontSize: 10,
              padding: "8px 12px",
              background: isSelected ? "var(--accent-cyan)" : "transparent",
              color: isSelected ? "var(--bg-deep)" : "var(--text-bone)",
              border: `2px solid ${isSelected ? "var(--accent-cyan)" : "var(--ink-muted)"}`,
              cursor: "pointer",
              whiteSpace: "nowrap",
              imageRendering: "pixelated",
              transition: "none",
            }}
          >
            {BRACKET_OPEN}{issue.identifier}{BRACKET_CLOSE}
          </button>
        );
      })}
    </div>
  );
}
