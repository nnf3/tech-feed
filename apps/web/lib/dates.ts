const tokyo: Intl.DateTimeFormatOptions = {
  timeZone: "Asia/Tokyo",
  year: "numeric",
  month: "short",
  day: "numeric",
};

export function formatDate(value: string, withTime = false) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat("ja-JP", withTime ? { ...tokyo, hour: "2-digit", minute: "2-digit" } : tokyo).format(
    date,
  );
}
