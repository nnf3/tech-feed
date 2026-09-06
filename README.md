# tech-feed

エンジニア向けの技術記事アグリゲータ。ローカルは Docker Compose だけで動かします。

## 起動

```bash
docker compose up --build
```

- Web: http://localhost:3001（3000 は nnf3-idp の sample-web と衝突するため）
- Feed API: http://localhost:8080
- Crawler: http://localhost:8081
- Elasticsearch: http://localhost:9200
- Postgres: localhost:5433（5432 は nnf3-idp と衝突するため。ユーザー / プロフィール用。記事検索は ES）

`crawler` プロセスが起動時と `CRAWL_INTERVAL`（既定 15 分）ごとに [Zenn RSS](https://zenn.dev/feed) と [Qiita API](https://qiita.com/api/v2/items) を取り込みます。Zenn のトピックは記事詳細 API で補完します。`CRAWL_INTERVAL=0` で定期実行だけ止められます。feed は一覧とユーザーだけを担当します。

ログインは [nnf3-idp](https://github.com/nnf3/nnf3-idp) の Hydra を使います。ブラウザは `http://localhost:3001` ではなく **`http://127.0.0.1:3001`** で開いてください。

```bash
./scripts/register-idp-client.sh
```

IdP 側で登録またはログインすると `http://127.0.0.1:3001/callback` に戻ります。ユーザーは feed が Postgres の `users` に保存します。`/me` と `/me/profile`、一覧のパーソナライズは BFF の `FEED_INTERNAL_TOKEN` と `X-User-ID` が必要です。ユーザー ID はパスやクエリには載せません。

`compose.yaml` は開発用です。Go と Next.js はボリュームマウントしているので、ソースを保存すればコンテナ内で再ビルド / ホットリロードされます。`go.mod` や `package.json` を変えたときだけ `docker compose up --build` し直してください。本番向けの焼き込みイメージは各ディレクトリの `Dockerfile` です。

## 構成

- `apps/web` — Next.js（フロント + BFF）。OIDC セッションと画面。Postgres は触らない
- `services/feed` — Go。`cmd/server` が一覧とユーザー、`cmd/crawler` が定期収集。記事は Elasticsearch、ユーザーとプロフィールは Postgres
- `infra/elasticsearch` — Kuromoji + ICU 入り Elasticsearch
