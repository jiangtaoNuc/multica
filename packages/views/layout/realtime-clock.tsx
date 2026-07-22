"use client";

import { useEffect, useState } from "react";
import { Clock } from "lucide-react";
import { cn } from "@multica/ui/lib/utils";

function formatTime(date: Date): string {
  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
}

export function RealtimeClock({ className }: { className?: string }) {
  const [now, setNow] = useState<Date | null>(null);

  useEffect(() => {
    setNow(new Date());
    const id = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <div
      className={cn(
        "pointer-events-none absolute top-2 right-3 z-20 flex items-center gap-1.5 rounded-md bg-background/70 px-2 py-1 text-xs font-medium tabular-nums text-muted-foreground backdrop-blur",
        className,
      )}
      aria-label="Current time"
    >
      <Clock className="size-3.5" />
      <span suppressHydrationWarning>{now ? formatTime(now) : "--:--:--"}</span>
    </div>
  );
}
