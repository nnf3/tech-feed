export type HighlightPart = {
  text: string;
  hit: boolean;
};

const markRe = /<mark>(.*?)<\/mark>/gs;

function decodeEntities(raw: string) {
  return raw
    .replaceAll("&lt;", "<")
    .replaceAll("&gt;", ">")
    .replaceAll("&quot;", '"')
    .replaceAll("&#39;", "'")
    .replaceAll("&amp;", "&");
}

export function highlightParts(html: string): HighlightPart[] {
  const parts: HighlightPart[] = [];
  let last = 0;
  for (const match of html.matchAll(markRe)) {
    const start = match.index ?? 0;
    if (start > last) {
      parts.push({ text: decodeEntities(html.slice(last, start)), hit: false });
    }
    parts.push({ text: decodeEntities(match[1] ?? ""), hit: true });
    last = start + match[0].length;
  }
  if (last < html.length) {
    parts.push({ text: decodeEntities(html.slice(last)), hit: false });
  }
  return parts;
}
