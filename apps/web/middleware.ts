import { NextRequest, NextResponse } from "next/server";

// IdP の redirect_uri と Cookie を 127.0.0.1 に揃える。localhost だと callback で state が消える。
export function middleware(req: NextRequest) {
  const host = req.headers.get("host") ?? "";
  if (host === "localhost:3001") {
    const url = req.nextUrl.clone();
    url.hostname = "127.0.0.1";
    url.port = "3001";
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}
