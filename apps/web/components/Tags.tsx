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
  active = "",
  query = "",
  links = false,
}: {
  tags: string[];
  interest?: string[];
  active?: string;
  query?: string;
  links?: boolean;
}) {
  if (tags.length === 0) {
    return null;
  }

  const highlighted = new Set(interest);

  return (
    <ul className="tags">
      {tags.map((tag) => {
        const className = [
          "tag",
          highlighted.has(tag) ? "tag-interest" : "",
          tag === active ? "tag-active" : "",
        ]
          .filter(Boolean)
          .join(" ");

        return (
          <li key={tag}>
            {links ? (
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
