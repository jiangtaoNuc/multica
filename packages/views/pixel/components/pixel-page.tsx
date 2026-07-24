"use client";

import { useState } from "react";
import { useT } from "../../i18n";
import { PageHeader } from "../../layout/page-header";
import { Banner } from "./banner";
import { IssueTabs } from "./issue-tabs";
import { Pipeline } from "./pipeline";
import { PixelSidebar } from "./pixel-sidebar";
import { HARNESS_STATES, STATE_LABELS } from "./types";
import { usePixelDemo } from "./use-pixel-demo";
import "../styles/pixel.css";

const STATUS_ALL = "all";

export function PixelPage() {
  const { t } = useT("pixel");
  const { t: tLayout } = useT("layout");
  const { issues, selectedId, snapshot, transition, onSelect } = usePixelDemo();
  const [statusFilter, setStatusFilter] = useState<string>(STATUS_ALL);

  const filteredIssues = statusFilter === STATUS_ALL
    ? issues
    : issues.filter((i) => i.status === statusFilter);

  const statusCounts = Object.fromEntries(
    HARNESS_STATES.map((s) => [s, issues.filter((i) => i.status === s).length])
  );

  return (
    <>
      <PageHeader>
        <h1 className="text-lg font-semibold">{tLayout(($) => $.nav.pixel)}</h1>
      </PageHeader>
      <div className="pixel-root flex-1 min-h-0 flex flex-col overflow-hidden bg-white">
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 6,
            padding: "8px 16px",
            overflowX: "auto",
            borderBottom: "4px solid var(--ink-muted)",
          }}
        >
          {[{ value: STATUS_ALL, label: t(($) => $.status_all), count: issues.length }, ...HARNESS_STATES.map((s) => ({
            value: s,
            label: STATE_LABELS[s],
            count: statusCounts[s] ?? 0,
          }))].map((chip) => {
            const isSelected = chip.value === statusFilter;
            return (
              <button
                key={chip.value}
                type="button"
                onClick={() => setStatusFilter(chip.value)}
                style={{
                  fontFamily: "var(--font-heading)",
                  fontSize: 8,
                  padding: "6px 10px",
                  background: isSelected ? "var(--accent-cyan)" : "transparent",
                  color: isSelected ? "var(--bg-deep)" : "var(--text-bone)",
                  border: `2px solid ${isSelected ? "var(--accent-cyan)" : "var(--ink-muted)"}`,
                  cursor: "pointer",
                  whiteSpace: "nowrap",
                  imageRendering: "pixelated",
                  transition: "none",
                  display: "flex",
                  alignItems: "center",
                  gap: 4,
                }}
              >
                {chip.label}
                <span style={{ fontSize: 8, opacity: isSelected ? 0.7 : 0.5 }}>{chip.count}</span>
              </button>
            );
          })}
        </div>

        <Banner
          message={t(($) => $.banner_message)}
          type="info"
        />

        <IssueTabs issues={filteredIssues} selectedId={selectedId} onSelect={onSelect} />

        <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
          <div
            style={{
              flex: 1,
              overflowX: "auto",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              padding: 16,
            }}
          >
            {snapshot ? (
              <Pipeline snapshot={snapshot} transition={transition} />
            ) : (
              <div
                style={{
                  fontFamily: "var(--font-body)",
                  fontSize: 24,
                  color: "#374151",
                  textAlign: "center",
                }}
              >
                {t(($) => $.select_issue)}
              </div>
            )}
          </div>

          {snapshot && <PixelSidebar snapshot={snapshot} />}
        </div>
      </div>
    </>
  );
}
