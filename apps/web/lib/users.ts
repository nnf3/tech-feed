export type User = {
  id: string;
  email: string;
  name: string;
};

const feedAPI = process.env.FEED_API_URL ?? "http://localhost:8080";

export async function getUser(id: string): Promise<User | null> {
  const res = await fetch(new URL(`/users/${encodeURIComponent(id)}`, feedAPI), {
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
  const res = await fetch(new URL("/users", feedAPI), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(user),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`user upsert returned ${res.status}`);
  }
  return (await res.json()) as User;
}
