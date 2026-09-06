export function Tags({
  tags,
  interest = [],
}: {
  tags: string[];
  interest?: string[];
}) {
  if (tags.length === 0) {
    return null;
  }

  const highlighted = new Set(interest);

  return (
    <ul className="tags">
      {tags.map((tag) => (
        <li key={tag} className={highlighted.has(tag) ? "tag tag-interest" : "tag"}>
          {tag}
        </li>
      ))}
    </ul>
  );
}
