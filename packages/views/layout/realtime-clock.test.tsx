import { render, screen, act } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { RealtimeClock } from "./realtime-clock";

describe("RealtimeClock", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-15T13:45:30"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders the current time after mount", () => {
    render(<RealtimeClock />);
    act(() => {
      vi.advanceTimersByTime(0);
    });
    expect(screen.getByLabelText("Current time")).toHaveTextContent("13:45:30");
  });

  it("advances every second", () => {
    render(<RealtimeClock />);
    act(() => {
      vi.advanceTimersByTime(0);
    });
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(screen.getByLabelText("Current time")).toHaveTextContent("13:45:33");
  });
});
