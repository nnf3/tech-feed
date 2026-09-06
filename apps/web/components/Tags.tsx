import { ProfileTag } from "@/components/ProfileTag";
import { feedHref } from "@/lib/feed";

export function Tags({
  tags,
  interest = [],
  exclude = [],
  active = "",
  query = "",
  sort = "",
  links = false,
  profile = false,
}: {
  tags: string[];
  interest?: string[];
  exclude?: string[];
  active?: string;
  query?: string;
  sort?: string;
  links?: boolean;
  profile?: boolean;
}) {
  if (tags.length === 0) {
    return null;
  }

  const interested = new Set(interest.map((item) => item.toLowerCase()));
  const excluded = new Set(exclude.map((item) => item.toLowerCase()));

  return (
    <ul className="tags">
      {tags.map((tag) => {
        const className = [
          "tag",
          interested.has(tag.toLowerCase()) ? "tag-interest" : "",
          excluded.has(tag.toLowerCase()) ? "tag-exclude" : "",
          tag.toLowerCase() === active.toLowerCase() ? "tag-active" : "",
        ]
          .filter(Boolean)
          .join(" ");

        return (
          <li key={tag}>
            {profile ? (
              <ProfileTag tag={tag} className={className} />
            ) : links ? (
              <a className={className} href={feedHref({ query, tag: tag === active ? "" : tag, sort })}>
                {tag}
              </a>
            ) : (
              <span className={className}>{tag}</span>
            )}
          </li>
        );
      })}
    </ul>
  );
}
