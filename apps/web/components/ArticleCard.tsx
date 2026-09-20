import { ArticleLink } from "@/components/ArticleLink";
import { BookmarkButton } from "@/components/BookmarkButton";
import { HighlightedText } from "@/components/HighlightedText";
import { Tags } from "@/components/Tags";
import { formatDate } from "@/lib/dates";
import type { Article } from "@/lib/feed";
import { articleKindLabel, articleLevelLabel } from "@/lib/feed";

export function ArticleCard({
  article,
  signedIn,
  saved,
  interest,
  exclude,
  tag,
  query,
  sort,
}: {
  article: Article;
  signedIn: boolean;
  saved: boolean;
  interest: string[];
  exclude: string[];
  tag: string;
  query: string;
  sort: string;
}) {
  return (
    <article className="card">
      {signedIn ? (
        <BookmarkButton
          saved={saved}
          article={{
            article_id: article.id,
            url: article.url,
            title: article.title,
            source: article.source,
          }}
        />
      ) : null}
      <ArticleLink
        className="card-body"
        href={article.url}
        record={signedIn}
        article={{
          article_id: article.id,
          url: article.url,
          title: article.title,
          source: article.source,
        }}
      >
        <div className="card-top">
          <span className={`source source-${article.source}`}>{article.source}</span>
          <span>{formatDate(article.published_at)}</span>
          {article.kind && articleKindLabel[article.kind] ? (
            <span className="signal">{articleKindLabel[article.kind]}</span>
          ) : null}
          {article.level && articleLevelLabel[article.level] ? (
            <span className="signal">{articleLevelLabel[article.level]}</span>
          ) : null}
          {article.promo ? <span className="signal signal-promo">PR</span> : null}
        </div>
        <h2>
          <HighlightedText text={article.title} highlighted={article.title_highlighted} />
        </h2>
        {article.summary ? (
          <p>
            <HighlightedText text={article.summary} highlighted={article.summary_highlighted} />
          </p>
        ) : null}
      </ArticleLink>
      <Tags
        tags={article.tags ?? []}
        interest={interest}
        exclude={exclude}
        active={tag}
        query={query}
        sort={sort}
        links={!signedIn}
        profile={signedIn}
      />
    </article>
  );
}
