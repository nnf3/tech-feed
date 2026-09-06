import { feedHeaders, feedURL } from "./feed-client";
import type { BookmarkInput } from "./bookmarks";

export type HistoryEntry = {
  user_id: string;
  article_id: string;
  url: string;
  title: string;
  source: string;
  viewed_at: string;
  view_count: number;
};

export async function listHistory(userID: string): Promise<HistoryEntry[]> {
  const res = await fetch(feedURL("/me/history"), {
    headers: feedHeaders(userID),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`history list returned ${res.status}`);
  }
  const data = (await res.json()) as { history?: HistoryEntry[] };
  return data.history ?? [];
}

export async function recordHistory(userID: string, article: BookmarkInput): Promise<void> {
  const res = await fetch(feedURL("/me/history"), {
    method: "PUT",
    headers: feedHeaders(userID, { "Content-Type": "application/json" }),
    body: JSON.stringify(article),
    cache: "no-store",
  });
  if (!res.ok) {
    const detail = await res.text();
    throw new Error(detail.trim() || `history record returned ${res.status}`);
  }
}
