import { Header } from "@/components/Header";
import { BookmarkButton } from "@/components/BookmarkButton";
import { getSession } from "@/lib/auth/session";
import { listBookmarks } from "@/lib/bookmarks";
import { redirect } from "next/navigation";

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
              <a className="card-body" href={bookmark.url} target="_blank" rel="noreferrer">
                <div className="card-top">
                  <span className={`source source-${bookmark.source}`}>{bookmark.source}</span>
                  <span>{formatDate(bookmark.created_at)}</span>
                </div>
                <h2>{bookmark.title}</h2>
              </a>
            </article>
          ))}
        </section>
      )}
    </main>
  );
}
