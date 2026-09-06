import { feedHeaders, feedURL } from "./feed-client";

export type User = {
  id: string;
  email: string;
  name: string;
};

export async function getUser(id: string): Promise<User | null> {
  const res = await fetch(feedURL("/me"), {
    headers: feedHeaders(id),
    cache: "no-store",
  });
  if (res.status === 404) {
    return null;
  }
  if (!res.ok) {
    throw new Error(`user lookup returned ${res.status}`);
  }
  return (await res.json()) as User;
}

export async function upsertUser(user: User): Promise<User> {
  const res = await fetch(feedURL("/me"), {
    method: "PUT",
    headers: feedHeaders(user.id, { "Content-Type": "application/json" }),
    body: JSON.stringify({ email: user.email, name: user.name }),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`user upsert returned ${res.status}`);
  }
  return (await res.json()) as User;
}
