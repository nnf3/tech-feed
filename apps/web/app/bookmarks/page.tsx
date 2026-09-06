import { ArticleLink } from "@/components/ArticleLink";
import { BookmarkButton } from "@/components/BookmarkButton";
import { Header } from "@/components/Header";
import { getSession } from "@/lib/auth/session";
import { listBookmarks } from "@/lib/bookmarks";
import { formatDate } from "@/lib/dates";
import { redirect } from "next/navigation";

export const dynamic = "force-dynamic";

export default async function BookmarksPage({
  searchParams,
}: {
  searchParams: Promise<{ bookmark_error?: string }>;
}) {
  const { bookmark_error: bookmarkError = "" } = await searchParams;
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  let bookmarks = [] as Awaited<ReturnType<typeof listBookmarks>>;
  let error = "";
  try {
    bookmarks = await listBookmarks(session.sub);
  } catch (err) {
    error = err instanceof Error ? err.message : "failed to load bookmarks";
  }

  return (
    <main>
      <Header session={session} />
      {bookmarkError ? <p className="meta">ブックマークを更新できませんでした: {bookmarkError}</p> : null}
      <p className="meta">
        {error ? `読み込みに失敗しました: ${error}` : `ブックマーク ${bookmarks.length} 件`}
      </p>
      {bookmarks.length === 0 && !error ? (
        <div className="empty">まだブックマークがありません。一覧のマークから保存できます。</div>
      ) : (
        <section className="list">
          {bookmarks.map((bookmark) => (
            <article key={bookmark.article_id} className="card">
              <BookmarkButton
                saved
                from="bookmarks"
                article={{
                  article_id: bookmark.article_id,
                  url: bookmark.url,
                  title: bookmark.title,
                  source: bookmark.source,
                }}
              />
              <ArticleLink
                record
                className="card-body"
                href={bookmark.url}
                article={{
                  article_id: bookmark.article_id,
                  url: bookmark.url,
                  title: bookmark.title,
                  source: bookmark.source,
                }}
              >
                <div className="card-top">
                  <span className={`source source-${bookmark.source}`}>{bookmark.source}</span>
                  <span>{formatDate(bookmark.created_at)}</span>
                </div>
                <h2>{bookmark.title}</h2>
              </ArticleLink>
            </article>
          ))}
        </section>
      )}
    </main>
  );
}
