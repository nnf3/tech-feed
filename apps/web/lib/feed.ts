export type Article = {
  id: string;
  source: string;
  url: string;
  title: string;
  summary: string;
  tags: string[];
  published_at: string;
};

const feedAPI = process.env.FEED_API_URL ?? "http://localhost:8080";

export async function listArticles(query = "", userID = "", tag = ""): Promise<Article[]> {
  const url = new URL("/articles", feedAPI);
  if (query) {
    url.searchParams.set("q", query);
  }
  if (userID) {
    url.searchParams.set("user_id", userID);
  }
  if (tag) {
    url.searchParams.set("tag", tag);
  }

  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`feed api returned ${res.status}`);
  }

  const data = (await res.json()) as { articles?: Article[] };
  return data.articles ?? [];
}
