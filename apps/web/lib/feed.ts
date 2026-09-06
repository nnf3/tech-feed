import { feedHeaders, feedURL } from "./feed-client";

export type FeedSort = "new" | "recommended";

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

export type FeedPage = {
  articles: Article[];
  next?: string;
};

export function feedHref({
  query = "",
  tag = "",
  sort = "",
}: {
  query?: string;
  tag?: string;
  sort?: string;
} = {}) {
  const params = new URLSearchParams();
  if (query) {
    params.set("q", query);
  }
  if (tag) {
    params.set("tag", tag);
  }
  if (sort && sort !== "new") {
    params.set("sort", sort);
  }
  const qs = params.toString();
  return qs ? `/?${qs}` : "/";
}

export function normalizeFeedSort(raw = "", canRecommend = false): FeedSort {
  return canRecommend && raw === "recommended" ? "recommended" : "new";
}

export async function listArticles({
  query = "",
  userID = "",
  tag = "",
  sort = "",
  after = "",
}: {
  query?: string;
  userID?: string;
  tag?: string;
  sort?: string;
  after?: string;
} = {}): Promise<FeedPage> {
  const url = feedURL("/articles");
  if (query) {
    url.searchParams.set("q", query);
  }
  if (tag) {
    url.searchParams.set("tag", tag);
  }
  if (sort && sort !== "new") {
    url.searchParams.set("sort", sort);
  }
  if (after) {
    url.searchParams.set("after", after);
  }

  const res = await fetch(url, { headers: feedHeaders(userID), cache: "no-store" });
  if (!res.ok) {
    throw new Error(`feed api returned ${res.status}`);
  }

  const data = (await res.json()) as FeedPage;
  return { articles: data.articles ?? [], next: data.next ?? "" };
}
