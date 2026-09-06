import { feedHeaders, feedURL } from "./feed-client";

export type Bookmark = {
  user_id: string;
  article_id: string;
  url: string;
  title: string;
  source: string;
  created_at: string;
};

export type BookmarkInput = {
  article_id: string;
  url: string;
  title: string;
  source: string;
};

export async function listBookmarks(userID: string): Promise<Bookmark[]> {
  const res = await fetch(feedURL("/me/bookmarks"), {
    headers: feedHeaders(userID),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`bookmark list returned ${res.status}`);
  }
  const data = (await res.json()) as { bookmarks?: Bookmark[] };
  return data.bookmarks ?? [];
}

export async function addBookmark(userID: string, bookmark: BookmarkInput): Promise<Bookmark> {
  const res = await fetch(feedURL("/me/bookmarks"), {
    method: "PUT",
    headers: feedHeaders(userID, { "Content-Type": "application/json" }),
    body: JSON.stringify(bookmark),
    cache: "no-store",
  });
  if (!res.ok) {
    const detail = await res.text();
    throw new Error(detail.trim() || `bookmark add returned ${res.status}`);
  }
  return (await res.json()) as Bookmark;
}

export async function removeBookmark(userID: string, articleID: string): Promise<void> {
  const res = await fetch(feedURL(`/me/bookmarks/${encodeURIComponent(articleID)}`), {
    method: "DELETE",
    headers: feedHeaders(userID),
    cache: "no-store",
  });
  if (!res.ok) {
    const detail = await res.text();
    throw new Error(detail.trim() || `bookmark remove returned ${res.status}`);
  }
}
