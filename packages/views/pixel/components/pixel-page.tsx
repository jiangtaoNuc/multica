"use client";

import { useT } from "../../i18n";
import { PageHeader } from "../../layout/page-header";
import { Banner } from "./banner";
import { IssueTabs } from "./issue-tabs";
import { Pipeline } from "./pipeline";
import { PixelSidebar } from "./pixel-sidebar";
import { usePixelDemo } from "./use-pixel-demo";
import "../styles/pixel.css";

export function PixelPage() {
  const { t } = useT();
  const { issues, selectedId, snapshot, transition, onSelect } = usePixelDemo();

  return (
    <>
      <PageHeader>
        <h1 className="text-lg font-semibold">{t(($) => $.nav.pixel)}</h1>
      </PageHeader>
      <div className="pixel-root flex-1 min-h-0 flex flex-col overflow-hidden">
        <header
          style={{
            padding: "16px 24px",
            borderBottom: "4px solid var(--ink-muted)",
            fontFamily: "var(--font-heading)",
            fontSize: 14,
            letterSpacing: 2,
            color: "var(--accent-cyan)",
            display: "flex",
            alignItems: "center",
            gap: 12,
          }}
        >
          <span style={{ fontSize: 20 }}>▓▓▓</span>
          CODING HARNESS
          <span style={{ fontSize: 20 }}>▓▓▓</span>
        </header>

        <Banner
          message="Pixel-art issue lifecycle visualization — switch tabs to see each FSM state"
          type="info"
        />

        <IssueTabs issues={issues} selectedId={selectedId} onSelect={onSelect} />

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
                  color: "var(--text-dust)",
                  textAlign: "center",
                }}
              >
                Select an issue to view its pipeline
              </div>
            )}
          </div>

          {snapshot && <PixelSidebar snapshot={snapshot} />}
        </div>
      </div>
    </>
  );
}
