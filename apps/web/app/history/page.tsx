import { ArticleLink } from "@/components/ArticleLink";
import { Header } from "@/components/Header";
import { getSession } from "@/lib/auth/session";
import { formatDate } from "@/lib/dates";
import { listHistory } from "@/lib/history";
import { redirect } from "next/navigation";

export const dynamic = "force-dynamic";

export default async function HistoryPage() {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  let items = [] as Awaited<ReturnType<typeof listHistory>>;
  let error = "";
  try {
    items = await listHistory(session.sub);
  } catch (err) {
    error = err instanceof Error ? err.message : "failed to load history";
  }

  return (
    <main>
      <Header session={session} />
      <p className="meta">{error ? `読み込みに失敗しました: ${error}` : `閲覧履歴 ${items.length} 件`}</p>
      {items.length === 0 && !error ? (
        <div className="empty">まだ閲覧履歴がありません。一覧の記事を開くとここに残ります。</div>
      ) : (
        <section className="list">
          {items.map((item) => (
            <article key={item.article_id} className="card">
              <ArticleLink
                record
                className="card-body"
                href={item.url}
                article={{
                  article_id: item.article_id,
                  url: item.url,
                  title: item.title,
                  source: item.source,
                }}
              >
                <div className="card-top">
                  <span className={`source source-${item.source}`}>{item.source}</span>
                  <span>{formatDate(item.viewed_at, true)}</span>
                  {item.view_count > 1 ? <span>{item.view_count}回</span> : null}
                </div>
                <h2>{item.title}</h2>
              </ArticleLink>
            </article>
          ))}
        </section>
      )}
    </main>
  );
}
