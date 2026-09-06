export type Profile = {
  user_id: string;
  interest_tags: string[];
  exclude_tags: string[];
};

const feedAPI = process.env.FEED_API_URL ?? "http://localhost:8080";

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
  const res = await fetch(new URL(`/users/${encodeURIComponent(userID)}/profile`, feedAPI), {
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`profile lookup returned ${res.status}`);
  }
  return (await res.json()) as Profile;
}

export async function upsertProfile(userID: string, profile: Pick<Profile, "interest_tags" | "exclude_tags">): Promise<Profile> {
  const res = await fetch(new URL(`/users/${encodeURIComponent(userID)}/profile`, feedAPI), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(profile),
    cache: "no-store",
  });
  if (!res.ok) {
    const detail = await res.text();
    throw new Error(detail.trim() || `profile upsert returned ${res.status}`);
  }
  return (await res.json()) as Profile;
}
