"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import type { Session } from "@/lib/auth/session";

function navClass(active: boolean) {
  return active ? "nav-icon nav-icon-active" : "nav-icon";
}

export function Header({ session, children }: { session: Session | null; children?: ReactNode }) {
  const pathname = usePathname();

  return (
    <header className="header">
      <a className="brand" href="/">
        <small>Local aggregator</small>
        <h1>Tech-Feed</h1>
      </a>
      <div className="header-actions">
        {children}
        <a
          className={navClass(pathname === "/")}
          href="/"
          aria-label="ホーム"
          title="ホーム"
          aria-current={pathname === "/" ? "page" : undefined}
        >
          <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
            <path
              d="M4.6 11.2 12 4.9l7.4 6.3"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.7"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M7.2 10.6V19h9.6v-8.4"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.7"
              strokeLinejoin="round"
            />
          </svg>
        </a>
        {session ? (
          <>
            <a
              className={navClass(pathname === "/history")}
              href="/history"
              aria-label="閲覧履歴"
              title="閲覧履歴"
              aria-current={pathname === "/history" ? "page" : undefined}
            >
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
                <circle cx="12" cy="12" r="7.2" fill="none" stroke="currentColor" strokeWidth="1.7" />
                <path d="M12 8.2v4.1l2.6 1.6" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
              </svg>
            </a>
            <a
              className={navClass(pathname === "/bookmarks")}
              href="/bookmarks"
              aria-label="ブックマーク"
              title="ブックマーク"
              aria-current={pathname === "/bookmarks" ? "page" : undefined}
            >
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
                <path
                  d="M7 4.5h10c.8 0 1.5.7 1.5 1.5v14l-6.5-3.4L5.5 20V6c0-.8.7-1.5 1.5-1.5z"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.7"
                  strokeLinejoin="round"
                />
              </svg>
            </a>
            <a
              className={navClass(pathname === "/account")}
              href="/account"
              aria-label="アカウント"
              title="アカウント"
              aria-current={pathname === "/account" ? "page" : undefined}
            >
              <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
                <circle cx="12" cy="8" r="3.2" fill="currentColor" />
                <path
                  d="M5.2 18.8c.7-3.2 3.4-5 6.8-5s6.1 1.8 6.8 5"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                />
              </svg>
            </a>
          </>
        ) : (
          <a className="auth-link" href="/login">
            ログイン
          </a>
        )}
      </div>
    </header>
  );
}
