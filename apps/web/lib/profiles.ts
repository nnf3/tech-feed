import { feedHeaders, feedURL } from "./feed-client";

export type Profile = {
  user_id: string;
  interest_tags: string[];
  exclude_tags: string[];
};

function tagKey(tag: string) {
  return tag.trim().toLowerCase();
}

function withoutTag(tags: string[], tag: string) {
  const key = tagKey(tag);
  return tags.filter((item) => tagKey(item) !== key);
}

function withTag(tags: string[], tag: string) {
  const key = tagKey(tag);
  if (!key || tags.some((item) => tagKey(item) === key)) {
    return tags;
  }
  return [...tags, key];
}

export function toggleTagLists(
  interest: string[],
  exclude: string[],
  tag: string,
  kind: "interest" | "exclude",
) {
  const key = tagKey(tag);
  if (!key) {
    return { interest, exclude };
  }
  const inInterest = interest.some((item) => tagKey(item) === key);
  const inExclude = exclude.some((item) => tagKey(item) === key);

  if (kind === "interest") {
    if (inInterest) {
      return { interest: withoutTag(interest, key), exclude };
    }
    return { interest: withTag(interest, key), exclude: withoutTag(exclude, key) };
  }
  if (inExclude) {
    return { interest, exclude: withoutTag(exclude, key) };
  }
  return { interest: withoutTag(interest, key), exclude: withTag(exclude, key) };
}

export function parseTags(raw: FormDataEntryValue | null): string[] {
  if (typeof raw !== "string") {
    return [];
  }
  return raw
    .split(/[,，、\s]+/)
    .map((tag) => tag.trim())
    .filter(Boolean);
}

export async function getProfile(userID: string): Promise<Profile> {
  const res = await fetch(feedURL("/me/profile"), {
    headers: feedHeaders(userID),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`profile lookup returned ${res.status}`);
  }
  return (await res.json()) as Profile;
}

export async function upsertProfile(userID: string, profile: Pick<Profile, "interest_tags" | "exclude_tags">): Promise<Profile> {
  const res = await fetch(feedURL("/me/profile"), {
    method: "PUT",
    headers: feedHeaders(userID, { "Content-Type": "application/json" }),
    body: JSON.stringify(profile),
    cache: "no-store",
  });
  if (!res.ok) {
    const detail = await res.text();
    throw new Error(detail.trim() || `profile upsert returned ${res.status}`);
  }
  return (await res.json()) as Profile;
}
