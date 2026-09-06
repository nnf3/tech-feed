import { tagHref } from "@/components/Tags";
import { sourceListLabel } from "@/lib/sources";

export function FeedMeta({
  error,
  count,
  personalized,
  tag,
  query,
}: {
  error: string;
  count: number;
  personalized: boolean;
  tag: string;
  query: string;
}) {
  return (
    <p className="meta">
      {error
        ? `読み込みに失敗しました: ${error}`
        : `${count} 件 · ${sourceListLabel()}${personalized ? " · プロフィール反映" : ""}`}
      {tag ? (
        <>
          {" · "}
          <a className="tag-filter" href={tagHref(tag, query, tag)}>
            タグ {tag} を解除
          </a>
        </>
      ) : null}
    </p>
  );
}
