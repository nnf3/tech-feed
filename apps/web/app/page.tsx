import { refreshFeed } from "./actions";
import { Header } from "@/components/Header";
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
  searchParams: Promise<{ q?: string; auth_error?: string }>;
}) {
  const { q = "", auth_error: authError = "" } = await searchParams;
  const session = await getSession();
  const interest = session ? (await getProfile(session.sub)).interest_tags : [];
  let articles = [] as Awaited<ReturnType<typeof listArticles>>;
  let error = "";

  try {
    articles = await listArticles(q, session?.sub ?? "");
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
          <button type="submit">検索</button>
        </form>
        <form action={refreshFeed}>
          <button type="submit">再取得</button>
        </form>
      </Header>

      {authError ? <p className="meta">ログインに失敗しました: {authError}</p> : null}
      <p className="meta">
        {error
          ? `読み込みに失敗しました: ${error}`
          : `${articles.length} 件 · Zenn RSS${session ? " · プロフィール反映" : ""}`}
      </p>

      {articles.length === 0 && !error ? (
        <div className="empty">まだ記事がありません。再取得を押すか、少し待って更新してください。</div>
      ) : (
        <section className="list">
          {articles.map((article) => (
            <a key={article.id} className="card" href={article.url} target="_blank" rel="noreferrer">
              <div className="card-top">
                <span className="source">{article.source}</span>
                <span>{formatDate(article.published_at)}</span>
              </div>
              <h2>{article.title}</h2>
              {article.summary ? <p>{article.summary}</p> : null}
              <Tags tags={article.tags ?? []} interest={interest} />
            </a>
          ))}
        </section>
      )}
    </main>
  );
}
