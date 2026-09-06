#!/usr/bin/env bash
set -euo pipefail

IDP_DIR="${IDP_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../nnf3-idp" && pwd)}"
CLIENT_ID="${OIDC_CLIENT_ID:-tech-feed-web}"
REDIRECT_URI="${OIDC_REDIRECT_URI:-http://127.0.0.1:3001/callback}"
ENDPOINT="${HYDRA_ADMIN_URL:-http://127.0.0.1:4445}"

if [[ ! -d "$IDP_DIR" ]]; then
  echo "nnf3-idp が見つかりません: $IDP_DIR" >&2
  echo "IDP_DIR=/path/to/nnf3-idp を指定してください" >&2
  exit 1
fi

create_client() {
  docker compose -f "$IDP_DIR/compose.yaml" exec -T hydra \
    hydra create oauth2-client \
      --endpoint "$ENDPOINT" \
      --id "$CLIENT_ID" \
      --name "Tech-Feed Web" \
      --grant-type authorization_code,refresh_token \
      --response-type code \
      --scope openid,offline,offline_access,email,profile \
      --redirect-uri "$REDIRECT_URI" \
      --token-endpoint-auth-method none \
      --skip-consent \
      --skip-logout-consent
}

if create_client; then
  echo "created $CLIENT_ID ($REDIRECT_URI)"
  exit 0
fi

echo "create に失敗したので update を試します"
docker compose -f "$IDP_DIR/compose.yaml" exec -T hydra \
  hydra update oauth2-client "$CLIENT_ID" \
    --endpoint "$ENDPOINT" \
    --name "Tech-Feed Web" \
    --grant-type authorization_code,refresh_token \
    --response-type code \
    --scope openid,offline,offline_access,email,profile \
    --redirect-uri "$REDIRECT_URI" \
    --token-endpoint-auth-method none \
    --skip-consent \
    --skip-logout-consent
echo "updated $CLIENT_ID ($REDIRECT_URI)"
