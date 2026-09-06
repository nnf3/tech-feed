"use client";

import type { ReactNode } from "react";
import { recordVisit } from "@/app/actions";
import type { BookmarkInput } from "@/lib/bookmarks";

export function ArticleLink({
  article,
  href,
  className,
  record = false,
  children,
}: {
  article: BookmarkInput;
  href: string;
  className?: string;
  record?: boolean;
  children: ReactNode;
}) {
  const log = () => {
    if (record) {
      void recordVisit(article);
    }
  };

  return (
    <a
      className={className}
      href={href}
      target="_blank"
      rel="noreferrer"
      onClick={log}
      onAuxClick={(event) => {
        if (event.button === 1) {
          log();
        }
      }}
    >
      {children}
    </a>
  );
}
