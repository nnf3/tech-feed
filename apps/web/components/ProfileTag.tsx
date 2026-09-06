"use client";

import { useTransition } from "react";
import { toggleProfileTag } from "@/app/actions";

export function ProfileTag({ tag, className }: { tag: string; className: string }) {
  const [pending, start] = useTransition();

  return (
    <button
      type="button"
      className={className}
      disabled={pending}
      title="クリックで関心、Option / Alt + クリックで除外"
      aria-label={`${tag}をプロフィールに反映`}
      onClick={(event) => {
        event.preventDefault();
        const kind = event.altKey || event.shiftKey ? "exclude" : "interest";
        start(() => toggleProfileTag(tag, kind));
      }}
    >
      {tag}
    </button>
  );
}
