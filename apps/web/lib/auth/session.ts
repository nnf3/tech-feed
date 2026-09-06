import { createHmac, timingSafeEqual } from "crypto";
import { cookies } from "next/headers";
import { authConfig } from "./config";

export const sessionCookie = "tf_session";
export const oauthCookie = "tf_oauth";

export type Session = {
  sub: string;
  email: string;
  name: string;
};

type OAuthState = {
  state: string;
  verifier: string;
};

function sign(payload: string, secret: string) {
  return createHmac("sha256", secret).update(payload).digest("base64url");
}

function encode(value: unknown, secret: string) {
  const payload = Buffer.from(JSON.stringify(value)).toString("base64url");
  return `${payload}.${sign(payload, secret)}`;
}

function decode<T>(raw: string, secret: string): T | null {
  const [payload, sig] = raw.split(".");
  if (!payload || !sig) {
    return null;
  }
  const expected = sign(payload, secret);
  const a = Buffer.from(sig);
  const b = Buffer.from(expected);
  if (a.length !== b.length || !timingSafeEqual(a, b)) {
    return null;
  }
  return JSON.parse(Buffer.from(payload, "base64url").toString()) as T;
}

export const sessionCookieOptions = {
  httpOnly: true,
  sameSite: "lax" as const,
  secure: false,
  path: "/",
  maxAge: 60 * 60 * 24 * 7,
};

export const oauthCookieOptions = {
  httpOnly: true,
  sameSite: "lax" as const,
  secure: false,
  path: "/",
  maxAge: 60 * 10,
};

export function encodeSession(session: Session) {
  return encode(session, authConfig().sessionSecret);
}

export function encodeOAuthState(value: OAuthState) {
  return encode(value, authConfig().sessionSecret);
}

export function readOAuthState(raw: string | undefined): OAuthState | null {
  if (!raw) {
    return null;
  }
  return decode<OAuthState>(raw, authConfig().sessionSecret);
}

export async function getSession(): Promise<Session | null> {
  const jar = await cookies();
  const raw = jar.get(sessionCookie)?.value;
  if (!raw) {
    return null;
  }
  return decode<Session>(raw, authConfig().sessionSecret);
}
