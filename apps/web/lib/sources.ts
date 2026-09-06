export const sources = [
  { id: "zenn", label: "Zenn" },
  { id: "qiita", label: "Qiita" },
] as const;

export function sourceListLabel() {
  return sources.map((source) => source.label).join(" / ");
}
