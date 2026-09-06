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
        ) : (
          <a className="auth-link" href="/login">
            ログイン
          </a>
        )}
      </div>
    </header>
  );
}
