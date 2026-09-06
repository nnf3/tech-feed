const feedAPI = process.env.FEED_API_URL ?? "http://localhost:8080";

function internalToken() {
  const value = process.env.FEED_INTERNAL_TOKEN;
  if (!value) {
    throw new Error("FEED_INTERNAL_TOKEN is not set");
  }
  return value;
}

export function feedURL(path: string) {
  return new URL(path, feedAPI);
}

export function feedHeaders(userID = "", extra?: HeadersInit): HeadersInit {
  const headers = new Headers(extra);
  headers.set("Authorization", `Bearer ${internalToken()}`);
  if (userID) {
    headers.set("X-User-ID", userID);
  }
  return headers;
}
