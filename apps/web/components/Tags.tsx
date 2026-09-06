import { ProfileTag } from "@/components/ProfileTag";

export function tagHref(tag: string, query = "", active = "") {
  const params = new URLSearchParams();
  if (query) {
    params.set("q", query);
  }
  if (tag !== active) {
    params.set("tag", tag);
  }
  const qs = params.toString();
  return qs ? `/?${qs}` : "/";
}

export function Tags({
  tags,
  interest = [],
  exclude = [],
  active = "",
  query = "",
  links = false,
  profile = false,
}: {
  tags: string[];
  interest?: string[];
  exclude?: string[];
  active?: string;
  query?: string;
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
              <a className={className} href={tagHref(tag, query, active)}>
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
