import { createHash, randomBytes } from "crypto";
import { createRemoteJWKSet, jwtVerify } from "jose";
import { authConfig } from "./config";

export function newPkce() {
  const verifier = randomBytes(32).toString("base64url");
  const challenge = createHash("sha256").update(verifier).digest("base64url");
  const state = randomBytes(16).toString("base64url");
  return { verifier, challenge, state };
}

export function authorizationUrl(state: string, challenge: string) {
  const cfg = authConfig();
  const url = new URL(cfg.authorizationUrl);
  url.searchParams.set("client_id", cfg.clientId);
  url.searchParams.set("redirect_uri", cfg.redirectUri);
  url.searchParams.set("response_type", "code");
  url.searchParams.set("scope", cfg.scope);
  url.searchParams.set("state", state);
  url.searchParams.set("code_challenge", challenge);
  url.searchParams.set("code_challenge_method", "S256");
  return url.toString();
}

type TokenResponse = {
  id_token?: string;
  access_token?: string;
  error?: string;
  error_description?: string;
};

export async function exchangeCode(code: string, verifier: string) {
  const cfg = authConfig();
  const body = new URLSearchParams({
    grant_type: "authorization_code",
    code,
    redirect_uri: cfg.redirectUri,
    client_id: cfg.clientId,
    code_verifier: verifier,
  });

  const res = await fetch(cfg.tokenUrl, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
    cache: "no-store",
  });
  const data = (await res.json()) as TokenResponse;
  if (!res.ok || !data.id_token) {
    throw new Error(data.error_description || data.error || `token exchange failed: ${res.status}`);
  }
  return data.id_token;
}

export async function claimsFromIdToken(idToken: string) {
  const cfg = authConfig();
  const JWKS = createRemoteJWKSet(new URL(cfg.jwksUrl));
  const { payload } = await jwtVerify(idToken, JWKS, {
    issuer: cfg.issuer,
    audience: cfg.clientId,
  });

  const email = typeof payload.email === "string" ? payload.email : "";
  const name =
    typeof payload.name === "string"
      ? payload.name
      : [payload.given_name, payload.family_name].filter((part) => typeof part === "string").join(" ");

  if (typeof payload.sub !== "string" || !payload.sub) {
    throw new Error("id_token に sub がありません");
  }

  return { sub: payload.sub, email, name };
}
