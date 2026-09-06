function required(name: string) {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} is not set`);
  }
  return value;
}

export function authConfig() {
  return {
    clientId: process.env.OIDC_CLIENT_ID ?? "tech-feed-web",
    redirectUri: process.env.OIDC_REDIRECT_URI ?? "http://127.0.0.1:3001/callback",
    // ブラウザが開く Hydra。issuer と一致させる。
    authorizationUrl: process.env.OIDC_AUTH_URL ?? "http://127.0.0.1:4444/oauth2/auth",
    issuer: process.env.OIDC_ISSUER ?? "http://127.0.0.1:4444",
    // コンテナ内の Next からホストの Hydra へ。
    tokenUrl: process.env.OIDC_TOKEN_URL ?? "http://host.docker.internal:4444/oauth2/token",
    jwksUrl: process.env.OIDC_JWKS_URL ?? "http://host.docker.internal:4444/.well-known/jwks.json",
    sessionSecret: required("AUTH_SECRET"),
    scope: "openid offline offline_access email profile",
    settingsUrl: process.env.IDP_SETTINGS_URL ?? "http://127.0.0.1:4455/settings",
  };
}
