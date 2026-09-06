import type { ReactNode } from "react";
import type { Session } from "@/lib/auth/session";

export function Header({ session, children }: { session: Session | null; children?: ReactNode }) {
  return (
    <header className="header">
      <a className="brand" href="/">
        <small>Local aggregator</small>
        <h1>Tech-Feed</h1>
      </a>
      <div className="header-actions">
        {children}
        {session ? (
          <>
            <a className="account-icon" href="/bookmarks" aria-label="ブックマーク" title="ブックマーク">
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
            <a className="account-icon" href="/account" aria-label="アカウント" title="アカウント">
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
