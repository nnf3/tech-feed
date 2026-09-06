import { highlightParts } from "@/lib/highlight";

export function HighlightedText({
  text,
  highlighted,
}: {
  text: string;
  highlighted?: string;
}) {
  if (!highlighted) {
    return text;
  }

  return (
    <>
      {highlightParts(highlighted).map((part, index) =>
        part.hit ? <mark key={index}>{part.text}</mark> : <span key={index}>{part.text}</span>,
      )}
    </>
  );
}
