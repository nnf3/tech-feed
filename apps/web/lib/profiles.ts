import { feedHeaders, feedURL } from "./feed-client";

export type Profile = {
  user_id: string;
  interest_tags: string[];
  exclude_tags: string[];
};

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
