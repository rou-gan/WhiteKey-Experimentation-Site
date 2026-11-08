# Frontend アーキテクチャ（Next.js）

「しろかぎ」実験場のフロントエンド構成方針。`frontend/` に Next.js を置き、backend とは別コンテナ・別 Render Web Service で動かす。

関連:

- デプロイ方針: [render.md](./render.md)
- Go API: [go-architecture.md](./go-architecture.md)
- セキュリティテスト計画: [security-test-plan.md](./security-test-plan.md)

## 方針

- **シンプルな Next.js:** App Router + TypeScript + ESLint。過剰な状態管理や UI ライブラリは初期導入しない。
- **standalone 出力:** `output: 'standalone'` で Docker 化し、Render Web Service（Node）として起動する。
- **backend と分離:** 別 Dockerfile・別イメージ・別 Deploy Hook。同一モノレポ・同一リリースタグ（`v*.*.*`）を共有する。
- **人間主導デプロイ:** `main` へのマージだけでは Render へ出さない（backend と同方針）。
- **低コスト:** 無料枠・低コストを優先。外部フロント専用ダッシュボードは初期導入しない。

## ディレクトリ構成

```text
frontend/
├── app/                 # App Router（layout / page）
├── public/              # 静的アセット（最小）
├── Dockerfile           # multi-stage、non-root
├── docker-compose.yml   # ローカル起動
├── next.config.ts       # output: 'standalone'
├── package.json
├── package-lock.json
├── tsconfig.json
└── eslint.config.mjs
```

## 依存の向き

```text
[Browser]
    ↓ HTTPS
[Render Web Service — frontend (Next.js standalone)]
    ↓ HTTP (NEXT_PUBLIC_API_BASE_URL)
[Render Web Service — backend (Go API)]
    ↓
[CSV 仮想 DB]
```

- ブラウザは frontend オリジンから API を呼ぶため、backend は **CORS** で frontend オリジンを許可する。
- PoC では認証なし。秘密情報を `NEXT_PUBLIC_*` に置かない。

## 環境変数

| 変数 | 用途 | 例 |
| --- | --- | --- |
| `API_BASE_URL` | server 側から呼ぶ backend API ベース URL（推奨・実行時） | `http://localhost:8080` / Render backend URL |
| `NEXT_PUBLIC_API_BASE_URL` | 互換用フォールバック（ビルド時インラインあり） | 同上 |
| `PORT` | Next 待ち受け（コンテナ／Render） | `3000` |
| `HOSTNAME` | バインド先 | `0.0.0.0` |

未設定時の API ベースはローカル開発用に `http://localhost:8080` とする。Docker / Render では `API_BASE_URL` を使う。

## PoC 画面

- ブランド名「しろかぎ」を第一ビューに出す。
- backend `GET /` の JSON（`{"message":"..."}`）を取得して表示する。
- 取得失敗時は汎用メッセージのみ（内部エラー詳細は出さない）。

## ローカル起動

```bash
# 別ターミナルで backend
cd backend && docker compose up --build

# frontend
cd frontend
npm ci
npm run dev
# http://localhost:3000
```

Docker のみ:

```bash
cd frontend
docker compose up --build
```

## CI/CD 概要

| Workflow | 起動 | 内容 |
| --- | --- | --- |
| Frontend Verify | PR / `main`（`frontend/**`） | lint、build、Trivy FS、Syft、Grype |
| Frontend Release Assets | タグ `v*.*.*` | イメージビルド／スキャン、同一 GitHub Release へ成果物添付 |
| Frontend Pre-Deploy | 手動 | デプロイ前ゲート（Render 接続は準備中） |

詳細は [render.md](./render.md) / [security-test-plan.md](./security-test-plan.md)。
