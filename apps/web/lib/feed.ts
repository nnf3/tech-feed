import { feedHeaders, feedURL } from "./feed-client";

export type Article = {
  id: string;
  source: string;
  url: string;
  title: string;
  summary: string;
  title_highlighted?: string;
  summary_highlighted?: string;
  tags: string[];
  published_at: string;
};

export async function listArticles(query = "", userID = "", tag = ""): Promise<Article[]> {
  const url = feedURL("/articles");
  if (query) {
    url.searchParams.set("q", query);
  }
  if (tag) {
    url.searchParams.set("tag", tag);
  }

  const res = await fetch(url, { headers: feedHeaders(userID), cache: "no-store" });
  if (!res.ok) {
    throw new Error(`feed api returned ${res.status}`);
  }

  const data = (await res.json()) as { articles?: Article[] };
  return data.articles ?? [];
}
