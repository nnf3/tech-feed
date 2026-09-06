"use client";

import { useTransition } from "react";
import { toggleBookmark } from "@/app/actions";
import type { BookmarkInput } from "@/lib/bookmarks";

export function BookmarkButton({
  article,
  saved,
  from,
}: {
  article: BookmarkInput;
  saved: boolean;
  from?: string;
}) {
  const [pending, start] = useTransition();

  return (
    <button
      type="button"
      className={`bookmark-btn${saved ? " bookmark-btn-on" : ""}`}
      disabled={pending}
      title={saved ? "ブックマークを外す" : "ブックマークする"}
      aria-label={saved ? `${article.title}のブックマークを外す` : `${article.title}をブックマーク`}
      aria-pressed={saved}
      onClick={(event) => {
        event.preventDefault();
        start(() => toggleBookmark({ ...article, saved, from }));
      }}
    >
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
        <path
          d="M7 4.5h10c.8 0 1.5.7 1.5 1.5v14l-6.5-3.4L5.5 20V6c0-.8.7-1.5 1.5-1.5z"
          fill={saved ? "currentColor" : "none"}
          stroke="currentColor"
          strokeWidth="1.7"
          strokeLinejoin="round"
        />
      </svg>
    </button>
  );
}
