import { BookmarkButton } from "@/components/BookmarkButton";
import { FeedMeta } from "@/components/FeedMeta";
import { Header } from "@/components/Header";
import { HighlightedText } from "@/components/HighlightedText";
import { Tags } from "@/components/Tags";
import { getSession } from "@/lib/auth/session";
import { listBookmarks } from "@/lib/bookmarks";
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
  searchParams: Promise<{
    q?: string;
    tag?: string;
    auth_error?: string;
    profile_error?: string;
    bookmark_error?: string;
  }>;
}) {
  const {
    q = "",
    tag = "",
    auth_error: authError = "",
    profile_error: profileError = "",
    bookmark_error: bookmarkError = "",
  } = await searchParams;
  const session = await getSession();
  const profile = session ? await getProfile(session.sub) : null;
  const interest = profile?.interest_tags ?? [];
  const exclude = profile?.exclude_tags ?? [];
  const saved = new Set<string>();
  if (session) {
    try {
      for (const item of await listBookmarks(session.sub)) {
        saved.add(item.article_id);
      }
    } catch {
      // 一覧は出して、保存状態だけ空にする
    }
  }
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
      {profileError ? <p className="meta">プロフィールを更新できませんでした: {profileError}</p> : null}
      {bookmarkError ? <p className="meta">ブックマークを更新できませんでした: {bookmarkError}</p> : null}
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
              {session ? (
                <BookmarkButton
                  saved={saved.has(article.id)}
                  article={{
                    article_id: article.id,
                    url: article.url,
                    title: article.title,
                    source: article.source,
                  }}
                />
              ) : null}
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
              <Tags
                tags={article.tags ?? []}
                interest={interest}
                exclude={exclude}
                active={tag}
                query={q}
                links={!session}
                profile={Boolean(session)}
              />
            </article>
          ))}
        </section>
      )}
    </main>
  );
}
