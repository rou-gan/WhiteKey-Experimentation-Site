# 「しろかぎ」のセキュリティ実験場

デザイン・バイ・セキュリティ（Design-by-Security）の考え方に基づき、コンテナや依存関係の脆弱性管理、サプライチェーンセキュリティ（SBOM）、セキュリティコンプライアンス検証（SCAP）を統合的に実践・検証するための実験場。

本番運用を想定したセキュアな CI/CD パイプラインおよび監視・監査プロセスを、Render.com および GitHub Actions 環境でローコスト（無料枠・低コスト）に検証することを目的とする。

> **これは本番向けテンプレートではありません。**  
> 学習・検証用の実験リポジトリです。`.env.example` の資格情報、テスト用パスワード、`internal/experiment/` や実験ブランチのドリルは、そのまま本番に持ち込まないでください。Deploy Hook や API キーなどの秘密は GitHub Actions の Secrets に置き、リポジトリ本文には書きません。

## リポジトリ構成

モノレポ構成。Go API は `backend/`、Next.js は `frontend/` に分離する（別コンテナ）。

```text
WhiteKey-Experimentation-Site/
├── backend/                 # Go API（DDD 風レイヤー、Docker、PostgreSQL）
│   └── openapi/             # OpenAPI 仕様
├── frontend/                # Next.js（App Router、standalone Docker）
├── render.yaml              # Render Blueprint（Postgres + 配線）
├── docs/                    # 設計ドキュメント
└── .github/workflows/       # backend / frontend 用 GitHub Actions
```


詳細設計:

- デプロイ方針: [docs/render.md](docs/render.md)
- Go アーキテクチャ: [docs/go-architecture.md](docs/go-architecture.md)
- Frontend アーキテクチャ: [docs/frontend-architecture.md](docs/frontend-architecture.md)
- セキュリティテスト計画: [docs/security-test-plan.md](docs/security-test-plan.md)
- ローカルセキュリティ Agent: [docs/security-agent-threat-model.md](docs/security-agent-threat-model.md)
- OpenAPI: [backend/openapi/openapi.yaml](backend/openapi/openapi.yaml)

## システムアーキテクチャ

```text
[GitHub Repository]
       ↓ (PR / main: backend/** または frontend/**)
[GitHub Actions — Backend Verify / Frontend Verify]
       ├── lint / test or build
       ├── Trivy filesystem scan
       ├── Syft SBOM (CycloneDX)
       └── Grype による SBOM 突き合わせ監査
                ↓
[人間] main からリリースタグ作成・push (例: v0.00.01)
       ↓
[GitHub Actions — Backend / Frontend Release Assets]（同一タグで並列）
       ├── タグ検証、image build / non-root
       ├── Trivy image / Syft / Grype
       └── 同一 GitHub Release へ各成果物を添付
                ↓
[人間] Backend / Frontend Pre-Deploy をそれぞれ手動実行
       ↓ （Render 接続後）
[Render.com]
       ├── Web Service: Go API（Docker / PostgreSQL）
       ├── Web Service: Next.js（別 Docker / standalone）
       └── Render Postgres（Free: whitekey-db）
```

アプリデータは Render Postgres（Free）に置く。スキーマ正本はアプリ埋め込みマイグレーション。Free は約 30 日で期限切れするため再作成前提（詳細は [docs/render.md](docs/render.md)）。

### 技術スタック

| 領域 | 採用技術 |
| --- | --- |
| ホスティング | Render.com（backend / frontend 別 Web Service + Postgres）。デプロイ接続は準備中 |
| Backend | Go（`net/http`）、Docker、PostgreSQL、CORS |
| Frontend | Next.js（App Router）、TypeScript、standalone Docker |
| バージョン管理・CI/CD | GitHub / GitHub Actions |
| セキュリティツール | golangci-lint / gosec、ESLint、Trivy、Syft（SBOM）、Grype。SCAP は今後 |

## セキュリティ検証項目

### Trivy による多層脆弱性スキャン

- **依存・ファイルシステム:** PR 時に `backend/` をスキャン（Critical / High で失敗）。
- **コンテナスキャン:** 手動 Pre-Deploy でイメージをスキャン。
- **CI/CD 統合:** 重要度に応じたビルドブロックを検証中。

### SBOM（Software Bill of Materials）と突き合わせ監査

- **生成（Syft）:** CycloneDX JSON を生成する（Verify は `backend/`、Release / Pre-Deploy はコンテナイメージ）。
- **監査（Grype）:** 生成した SBOM を入力に脆弱性スキャンし、High 以上で失敗させる（SBOM とスキャン結果の突き合わせ）。
- **保管:**
  - PR / 手動実行: Actions Artifacts（短期）
  - タグ push: GitHub Release 資産として長期保管（そのタグ向けに再生成）
- **補完:** Trivy による fs / image 直接スキャンも併用する。

### SCAP（Security Content Automation Protocol）

- 今後の拡張項目。セキュリティ設定や CIS Benchmark 等との突合実験を予定。

## GitHub Actions（現状）

| ワークフロー | 起動 | 内容 |
| --- | --- | --- |
| `Backend Verify` | PR、または `main` への push（`backend/**`） | Lint、Test、Trivy FS、Syft SBOM、Grype 監査、coverage Artifact |
| `Frontend Verify` | PR、または `main` への push（`frontend/**`） | ESLint、`next build`、Trivy FS、Syft SBOM、Grype 監査 |
| `Backend Release Assets` | タグ push（`v*.*.*`） | タグ検証、image ビルド／スキャン、**GitHub Release へ成果物添付** |
| `Frontend Release Assets` | タグ push（`v*.*.*`） | 同上（frontend 成果物を同一 Release へ添付） |
| `Backend Pre-Deploy` | 手動（`workflow_dispatch`） | デプロイ前ゲート（**Render デプロイは未接続**） |
| `Frontend Pre-Deploy` | 手動（`workflow_dispatch`） | 同上（frontend） |
| `ZAP DAST` | 手動（`workflow_dispatch`）、週次 schedule | runner 上で app 起動 → ZAP AF（Baseline + OpenAPI パッシブ）。**High で失敗**。PR マージゲートではない |
| `Semgrep` | PR、または `main` への push（`backend/**` / `frontend/**`） | `semgrep scan --error`（security-audit / OWASP / Go / JS）。クラウド不要。findings で失敗 |

デプロイ方針:

- `main` へのマージだけでは Render へデプロイしない。
- リリースタグ（`v0.00.00` 形式）を `main` から作成して push すると、backend / frontend のセキュリティ成果物付き Release が自動作成される。
- Render へのデプロイは、将来各 Pre-Deploy を手動実行して行う（サービスごとに別 Deploy Hook）。

テスト／スキャン結果は Actions のログ・Artifacts、およびリリースタグ時は Release 資産に残す（外部ダッシュボードは初期導入しない）。詳細は [docs/security-test-plan.md](docs/security-test-plan.md) / [docs/render.md](docs/render.md)。

## ローカルセキュリティレビュー（Phase A）

Cursor Automations / クラウド Agent は使わない。**push または PR 作成の直後**に、手元の Cursor で `/review-security` と Semgrep（Skills / CLI）を実行する（作成前は実行しない）。

1. `/review-security`（`branch changes`、base は `main`）
2. Semgrep（[`.cursor/skills/semgrep-review/SKILL.md`](.cursor/skills/semgrep-review/SKILL.md)。任意で `npx skills add semgrep/skills`）
3. 指摘を Severity / Location / Finding でまとめる（Semgrep は要約）
4. 該当 PR に `gh pr review --comment` で投稿する（`APPROVE` / `REQUEST_CHANGES` は使わない）

決定的ゲートは `Backend Verify` / `Frontend Verify` / `Semgrep`。ローカルレビューは advisory。脅威モデルは [docs/security-agent-threat-model.md](docs/security-agent-threat-model.md)。前提としてローカルで `gh auth status` が通っていること。

## 初期リリース仕様（PoC）

### 機能要件

- **エンドポイント:**
  - `GET /` — PostgreSQL 既定メッセージを JSON で返す
  - `GET /experiment/greet?name=` — 名前を埋め込んだ HTML 挨拶（埋め込み前に HTML エスケープ）
- **動作仕様:**
  1. 起動時に migrate し、`messages` から id=`1` を取得する。
  2. `"Hello World"` を JSON（`{"message":"Hello World"}`）で返す。
  3. `/experiment/greet` はクエリ `name` を HTML へ埋め込む前にエスケープし、反射型 XSS を防ぐ（OpenAPI: [backend/openapi/openapi.yaml](backend/openapi/openapi.yaml)）。

### ローカル起動（Docker）

```bash
cd backend
docker compose up --build
curl http://localhost:8080/
curl "http://localhost:8080/experiment/greet?name=%E3%81%97%E3%82%8D%E3%81%8B%E3%81%8E"
```

`api` と `db`（Postgres）が起動する。DB 資格情報は [backend/.env.example](backend/.env.example) を `backend/.env` にコピーして設定する（compose が読み込む）。その他: `ADDR`（既定 `:8080`）、`CORS_ALLOWED_ORIGINS`（例: `http://localhost:3000`）。

### Frontend（Next.js）

```bash
cd frontend
npm ci
npm run dev
# http://localhost:3000 — backend GET / の message を表示
```

環境変数（任意）: `API_BASE_URL` または `NEXT_PUBLIC_API_BASE_URL`（既定 `http://localhost:8080`）。詳細は [docs/frontend-architecture.md](docs/frontend-architecture.md)。

## 検証ロードマップ

- **Phase 1:** Docker 化、backend の Verify / Pre-Deploy Actions、Trivy 導入（進行中。Render デプロイ接続が残作業）
- **Phase 2:** Syft SBOM + Grype 突き合わせ監査の定着、依存スキャンの継続監視
- **Phase 2.5:** frontend Next.js + Verify / Release / Pre-Deploy（Render 接続は準備中）
- **Phase 3:** SCAP / Policy as Code、監査ログ収集の実験
- **Phase 4:** Auth0 による認証基盤の追加（frontend / backend への統合と保護 API の実験）
- **Phase 5:** 要認証の MCP サーバー構築（Auth0 と連携したアクセス制御の検証）
  - 実験（API キー + stdio）: [`mcp/message-server`](mcp/message-server)
