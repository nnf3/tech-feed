import { FeedList } from "@/components/FeedList";
import { FeedMeta } from "@/components/FeedMeta";
import { Header } from "@/components/Header";
import { SortNav } from "@/components/SortNav";
import { getSession } from "@/lib/auth/session";
import { listBookmarks } from "@/lib/bookmarks";
import { listArticles, normalizeFeedSort, type FeedPage } from "@/lib/feed";
import { getProfile } from "@/lib/profiles";

export const dynamic = "force-dynamic";

export default async function HomePage({
  searchParams,
}: {
  searchParams: Promise<{
    q?: string;
    tag?: string;
    sort?: string;
    auth_error?: string;
    profile_error?: string;
    bookmark_error?: string;
  }>;
}) {
  const {
    q = "",
    tag = "",
    sort: sortParam = "",
    auth_error: authError = "",
    profile_error: profileError = "",
    bookmark_error: bookmarkError = "",
  } = await searchParams;
  const session = await getSession();
  const sort = normalizeFeedSort(sortParam, Boolean(session));
  const profile = session ? await getProfile(session.sub) : null;
  const interest = profile?.interest_tags ?? [];
  const exclude = profile?.exclude_tags ?? [];
  const saved: string[] = [];
  if (session) {
    try {
      for (const item of await listBookmarks(session.sub)) {
        saved.push(item.article_id);
      }
    } catch {
      // 一覧は出して、保存状態だけ空にする
    }
  }
  let page: FeedPage = { articles: [] };
  let error = "";

  try {
    page = await listArticles({ query: q, userID: session?.sub ?? "", tag, sort });
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
          {sort !== "new" ? <input type="hidden" name="sort" value={sort} /> : null}
        </form>
      </Header>

      {authError ? <p className="meta">ログインに失敗しました: {authError}</p> : null}
      {profileError ? <p className="meta">プロフィールを更新できませんでした: {profileError}</p> : null}
      {bookmarkError ? <p className="meta">ブックマークを更新できませんでした: {bookmarkError}</p> : null}
      <SortNav query={q} tag={tag} sort={sort} recommend={Boolean(session)} />
      <FeedMeta error={error} personalized={Boolean(session)} tag={tag} query={q} sort={sort} />

      {page.articles.length === 0 && !error ? (
        <div className="empty">
          {tag
            ? `タグ「${tag}」の記事はありません。タグを外して一覧に戻ってください。`
            : "まだ記事がありません。取り込みを待って更新してください。"}
        </div>
      ) : page.articles.length > 0 ? (
        <FeedList
          key={`${sort}|${q}|${tag}`}
          initial={page}
          query={q}
          tag={tag}
          sort={sort}
          signedIn={Boolean(session)}
          saved={saved}
          interest={interest}
          exclude={exclude}
        />
      ) : null}
    </main>
  );
}
