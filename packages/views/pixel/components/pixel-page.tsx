"use client";

import { useT } from "../../i18n";
import { PageHeader } from "../../layout/page-header";
import { Banner } from "./banner";
import { IssueTabs } from "./issue-tabs";
import { Pipeline } from "./pipeline";
import { PixelSidebar } from "./pixel-sidebar";
import { usePixelDemo } from "./use-pixel-demo";
import "../styles/pixel.css";

const BLOCK_GLYPH = "▓▓▓";
const APP_TITLE = "CODING HARNESS";

export function PixelPage() {
  const { t } = useT("pixel");
  const { t: tLayout } = useT("layout");
  const { issues, selectedId, snapshot, transition, onSelect } = usePixelDemo();

  return (
    <>
      <PageHeader>
        <h1 className="text-lg font-semibold">{tLayout(($) => $.nav.pixel)}</h1>
      </PageHeader>
      <div className="pixel-root flex-1 min-h-0 flex flex-col overflow-hidden bg-white">
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
          <span aria-hidden="true" style={{ fontSize: 20 }}>{BLOCK_GLYPH}</span>
          {APP_TITLE}
          <span aria-hidden="true" style={{ fontSize: 20 }}>{BLOCK_GLYPH}</span>
        </header>

        <Banner
          message={t(($) => $.banner_message)}
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
