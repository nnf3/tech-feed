# tech-feed

エンジニア向けの技術記事アグリゲータ。ローカルは Docker Compose だけで動かします。

## 起動

```bash
docker compose up --build
```

- Web: http://localhost:3001（3000 は nnf3-idp の sample-web と衝突するため）
- Feed API: http://localhost:8080
- Elasticsearch: http://localhost:9200

起動時に [Zenn RSS](https://zenn.dev/feed) を取り込み、トップページに一覧します。ログインはまだありません。

## 構成

- `apps/web` — Next.js（フロント + BFF）
- `services/feed` — Go（収集 / 検索 / 並び替え）
- `infra/elasticsearch` — Kuromoji + ICU 入り Elasticsearch
