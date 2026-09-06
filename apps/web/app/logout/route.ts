import { NextResponse } from "next/server";
import { sessionCookie } from "@/lib/auth/session";

export async function POST() {
  const home = process.env.OIDC_REDIRECT_URI ?? "http://127.0.0.1:3001/callback";
  const res = NextResponse.redirect(new URL("/", home), { status: 303 });
  res.cookies.delete(sessionCookie);
  return res;
}
