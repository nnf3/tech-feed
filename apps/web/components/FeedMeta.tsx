import { feedHref } from "@/lib/feed";
import { sourceListLabel } from "@/lib/sources";

export function FeedMeta({
  error,
  personalized,
  tag,
  query,
  sort,
}: {
  error: string;
  personalized: boolean;
  tag: string;
  query: string;
  sort: string;
}) {
  return (
    <p className="meta">
      {error
        ? `読み込みに失敗しました: ${error}`
        : `${sourceListLabel()}${personalized ? " · タグをクリックで関心" : ""}`}
      {tag ? (
        <>
          {" · "}
          <a className="tag-filter" href={feedHref({ query, sort })}>
            タグ {tag} を解除
          </a>
        </>
      ) : null}
    </p>
  );
}
