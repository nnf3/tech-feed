import postgres from "postgres";

const url = process.env.DATABASE_URL;
if (!url) {
  throw new Error("DATABASE_URL is not set");
}

export const sql = postgres(url, { max: 5 });

let ready: Promise<void> | null = null;

export function ensureUsersTable() {
  if (!ready) {
    ready = sql`
      CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT NOT NULL DEFAULT '',
        name TEXT NOT NULL DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
      )
    `.then(() => undefined);
  }
  return ready;
}

export async function getUser(id: string) {
  await ensureUsersTable();
  const rows = await sql<{ id: string; email: string; name: string }[]>`
    SELECT id, email, name FROM users WHERE id = ${id}
  `;
  return rows[0] ?? null;
}

export async function upsertUser(user: { id: string; email: string; name: string }) {
  await ensureUsersTable();
  await sql`
    INSERT INTO users (id, email, name)
    VALUES (${user.id}, ${user.email}, ${user.name})
    ON CONFLICT (id) DO UPDATE SET
      email = EXCLUDED.email,
      name = EXCLUDED.name,
      updated_at = now()
  `;
}
