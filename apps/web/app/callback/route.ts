import { NextRequest, NextResponse } from "next/server";
import { upsertUser } from "@/lib/users";
import { claimsFromIdToken, exchangeCode } from "@/lib/auth/oidc";
import {
  encodeSession,
  oauthCookie,
  readOAuthState,
  sessionCookie,
  sessionCookieOptions,
} from "@/lib/auth/session";

function appOrigin() {
  const redirect = process.env.OIDC_REDIRECT_URI ?? "http://127.0.0.1:3001/callback";
  return new URL(redirect).origin;
}

function fail(home: string, reason: string) {
  return NextResponse.redirect(new URL(`/?auth_error=${encodeURIComponent(reason)}`, home));
}

export async function GET(req: NextRequest) {
  const home = appOrigin();
  const url = new URL(req.url);
  const hydraError = url.searchParams.get("error");
  if (hydraError) {
    const desc = url.searchParams.get("error_description") ?? hydraError;
    return fail(home, desc);
  }

  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state");
  const pending = readOAuthState(req.cookies.get(oauthCookie)?.value);
  if (!code) {
    return fail(home, "authorization_code がありません");
  }
  if (!pending) {
    return fail(home, "ログイン用 Cookie がありません。http://127.0.0.1:3001 からやり直してください");
  }
  if (!state || pending.state !== state) {
    return fail(home, "state が一致しません");
  }

  try {
    const idToken = await exchangeCode(code, pending.verifier);
    const claims = await claimsFromIdToken(idToken);
    await upsertUser({ id: claims.sub, email: claims.email, name: claims.name });
    const res = NextResponse.redirect(new URL("/", home));
    res.cookies.set(sessionCookie, encodeSession(claims), sessionCookieOptions);
    res.cookies.delete(oauthCookie);
    return res;
  } catch (err) {
    const message = err instanceof Error ? err.message : "callback_failed";
    return fail(home, message);
  }
}
