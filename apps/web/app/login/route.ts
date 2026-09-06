import { NextResponse } from "next/server";
import { authorizationUrl, newPkce } from "@/lib/auth/oidc";
import { encodeOAuthState, oauthCookie, oauthCookieOptions } from "@/lib/auth/session";

export async function GET() {
  const { verifier, challenge, state } = newPkce();
  const res = NextResponse.redirect(authorizationUrl(state, challenge));
  res.cookies.set(oauthCookie, encodeOAuthState({ state, verifier }), oauthCookieOptions);
  return res;
}
