import { feedHref } from "@/lib/feed";

export function SortNav({
  query,
  tag,
  sort,
  recommend,
}: {
  query: string;
  tag: string;
  sort: string;
  recommend: boolean;
}) {
  return (
    <nav className="sorts" aria-label="並び順">
      <a
        className={sort === "new" ? "sort-link sort-active" : "sort-link"}
        href={feedHref({ query, tag, sort: "new" })}
        aria-current={sort === "new" ? "page" : undefined}
      >
        新着
      </a>
      {recommend ? (
        <a
          className={sort === "recommended" ? "sort-link sort-active" : "sort-link"}
          href={feedHref({ query, tag, sort: "recommended" })}
          aria-current={sort === "recommended" ? "page" : undefined}
        >
          おすすめ
        </a>
      ) : null}
    </nav>
  );
}
