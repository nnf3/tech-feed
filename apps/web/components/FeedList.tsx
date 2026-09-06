"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { loadMoreFeed } from "@/app/actions";
import { ArticleCard } from "@/components/ArticleCard";
import type { Article, FeedPage } from "@/lib/feed";

export function FeedList({
  initial,
  query,
  tag,
  sort,
  signedIn,
  saved,
  interest,
  exclude,
}: {
  initial: FeedPage;
  query: string;
  tag: string;
  sort: string;
  signedIn: boolean;
  saved: string[];
  interest: string[];
  exclude: string[];
}) {
  const [articles, setArticles] = useState<Article[]>(initial.articles);
  const [next, setNext] = useState(initial.next ?? "");
  const [pending, setPending] = useState(false);
  const [loadError, setLoadError] = useState("");
  const nextRef = useRef(next);
  const pendingRef = useRef(false);
  const sentinel = useRef<HTMLDivElement>(null);
  const savedSet = new Set(saved);

  nextRef.current = next;

  const loadMore = useCallback(async () => {
    if (!nextRef.current || pendingRef.current) {
      return;
    }
    pendingRef.current = true;
    setPending(true);
    try {
      const page = await loadMoreFeed({ query, tag, sort, after: nextRef.current });
      setArticles((current) => {
        const seen = new Set(current.map((item) => item.id));
        return [...current, ...page.articles.filter((item) => !seen.has(item.id))];
      });
      setNext(page.next ?? "");
      setLoadError("");
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "failed to load more");
    } finally {
      pendingRef.current = false;
      setPending(false);
    }
  }, [query, tag, sort]);

  useEffect(() => {
    const node = sentinel.current;
    if (!node || !next) {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          void loadMore();
        }
      },
      { rootMargin: "240px" },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [loadMore, next]);

  return (
    <section className="list">
      {articles.map((article) => (
        <ArticleCard
          key={article.id}
          article={article}
          signedIn={signedIn}
          saved={savedSet.has(article.id)}
          interest={interest}
          exclude={exclude}
          tag={tag}
          query={query}
          sort={sort}
        />
      ))}
      {next ? <div ref={sentinel} className="feed-more" aria-hidden="true" /> : null}
      {pending ? <p className="feed-more">読み込み中…</p> : null}
      {loadError ? (
        <p className="feed-more">
          続きを読めませんでした。
          <button type="button" onClick={() => void loadMore()}>
            再試行
          </button>
        </p>
      ) : null}
    </section>
  );
}
