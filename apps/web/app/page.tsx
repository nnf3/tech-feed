import { FeedMeta } from "@/components/FeedMeta";
import { Header } from "@/components/Header";
import { HighlightedText } from "@/components/HighlightedText";
import { Tags } from "@/components/Tags";
import { getSession } from "@/lib/auth/session";
import { listArticles } from "@/lib/feed";
import { getProfile } from "@/lib/profiles";

export const dynamic = "force-dynamic";

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "short",
    day: "numeric",
  }).format(date);
}

export default async function HomePage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; tag?: string; auth_error?: string }>;
}) {
  const { q = "", tag = "", auth_error: authError = "" } = await searchParams;
  const session = await getSession();
  const interest = session ? (await getProfile(session.sub)).interest_tags : [];
  let articles = [] as Awaited<ReturnType<typeof listArticles>>;
  let error = "";

  try {
    articles = await listArticles(q, session?.sub ?? "", tag);
  } catch (err) {
    error = err instanceof Error ? err.message : "failed to load feed";
  }

  return (
    <main>
      <Header session={session}>
        <form className="toolbar" action="/">
          <input
            type="search"
            name="q"
            defaultValue={q}
            placeholder="キーワード"
            aria-label="記事を検索"
          />
          {tag ? <input type="hidden" name="tag" value={tag} /> : null}
          <button type="submit">検索</button>
        </form>
      </Header>

      {authError ? <p className="meta">ログインに失敗しました: {authError}</p> : null}
      <FeedMeta
        error={error}
        count={articles.length}
        personalized={Boolean(session)}
        tag={tag}
        query={q}
      />

      {articles.length === 0 && !error ? (
        <div className="empty">
          {tag
            ? `タグ「${tag}」の記事はありません。タグを外して一覧に戻ってください。`
            : "まだ記事がありません。取り込みを待って更新してください。"}
        </div>
      ) : (
        <section className="list">
          {articles.map((article) => (
            <article key={article.id} className="card">
              <a className="card-body" href={article.url} target="_blank" rel="noreferrer">
                <div className="card-top">
                  <span className={`source source-${article.source}`}>{article.source}</span>
                  <span>{formatDate(article.published_at)}</span>
                </div>
                <h2>
                  <HighlightedText text={article.title} highlighted={article.title_highlighted} />
                </h2>
                {article.summary ? (
                  <p>
                    <HighlightedText text={article.summary} highlighted={article.summary_highlighted} />
                  </p>
                ) : null}
              </a>
              <Tags tags={article.tags ?? []} interest={interest} active={tag} query={q} links />
            </article>
          ))}
        </section>
      )}
    </main>
  );
}
