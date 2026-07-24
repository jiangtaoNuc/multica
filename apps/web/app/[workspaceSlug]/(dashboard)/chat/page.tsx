"use client";

import { useEffect } from "react";
import { useChatStore } from "@multica/core/chat";
import { ChatWindow } from "@multica/views/chat";

export default function ChatPage() {
  const setOpen = useChatStore((s) => s.setOpen);

  useEffect(() => {
    setOpen(true);
  }, [setOpen]);

  return (
    <div className="flex flex-1 flex-col min-h-0">
      <ChatWindow variant="page" />
    </div>
  );
}
