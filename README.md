# 「しろかぎ」のセキュリティ実験場

デザイン・バイ・セキュリティ（Design-by-Security）の考え方に基づき、コンテナや依存関係の脆弱性管理、サプライチェーンセキュリティ（SBOM）、セキュリティコンプライアンス検証（SCAP）を統合的に実践・検証するための実験場。

本番運用を想定したセキュアな CI/CD パイプラインおよび監視・監査プロセスを、Render.com および GitHub Actions 環境でローコスト（無料枠・低コスト）に検証することを目的とする。

## リポジトリ構成

モノレポ構成。Go API は `backend/`、Next.js は `frontend/` に分離する（別コンテナ）。

```text
WhiteKey-Experimentation-Site/
├── backend/                 # Go API（DDD 風レイヤー、Docker、CSV 仮想 DB）
│   └── openapi/             # OpenAPI 仕様
├── frontend/                # Next.js（App Router、standalone Docker）
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
       ├── Web Service: Go API（Docker / CSV）
       └── Web Service: Next.js（別 Docker / standalone）
```

PostgreSQL などのマネージド DB は採用しない（無料枠の寿命・低コスト方針のため）。データはリポジトリ同梱の CSV を仮想 DB とする。詳細は [docs/render.md](docs/render.md)。

### 技術スタック

| 領域 | 採用技術 |
| --- | --- |
| ホスティング | Render.com（backend / frontend 別 Web Service）。デプロイ接続は準備中 |
| Backend | Go（`net/http`）、Docker、CSV 仮想 DB、CORS |
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

デプロイ方針:

- `main` へのマージだけでは Render へデプロイしない。
- リリースタグ（`v0.00.00` 形式）を `main` から作成して push すると、backend / frontend のセキュリティ成果物付き Release が自動作成される。
- Render へのデプロイは、将来各 Pre-Deploy を手動実行して行う（サービスごとに別 Deploy Hook）。

テスト／スキャン結果は Actions のログ・Artifacts、およびリリースタグ時は Release 資産に残す（外部ダッシュボードは初期導入しない）。詳細は [docs/security-test-plan.md](docs/security-test-plan.md) / [docs/render.md](docs/render.md)。

## ローカルセキュリティレビュー（Phase A）

Cursor Automations / クラウド Agent は使わない。**push または PR 作成の直後**に、手元の Cursor で `/review-security` を実行する（作成前は実行しない）。

1. `/review-security`（`branch changes`、base は `main`）
2. 指摘を Severity / Location / Finding でまとめる
3. 該当 PR に `gh pr review --comment` で投稿する（`APPROVE` / `REQUEST_CHANGES` は使わない）

必須のマージゲートは引き続き `Backend Verify`。脅威モデルと詳細手順は [docs/security-agent-threat-model.md](docs/security-agent-threat-model.md)。前提としてローカルで `gh auth status` が通っていること。

## 初期リリース仕様（PoC）

### 機能要件

- **エンドポイント:**
  - `GET /` — CSV 既定メッセージを JSON で返す
  - `GET /experiment/greet?name=` — 名前を埋め込んだ HTML 挨拶（埋め込み前に HTML エスケープ）
- **動作仕様:**
  1. 同梱 CSV（`backend/data/messages.csv`）からメッセージを取得する。
  2. `"Hello World"` を JSON（`{"message":"Hello World"}`）で返す。
  3. `/experiment/greet` はクエリ `name` を HTML へ埋め込む前にエスケープし、反射型 XSS を防ぐ（OpenAPI: [backend/openapi/openapi.yaml](backend/openapi/openapi.yaml)）。

### ローカル起動（Docker）

```bash
cd backend
docker compose up --build
curl http://localhost:8080/
curl "http://localhost:8080/experiment/greet?name=%E3%81%97%E3%82%8D%E3%81%8B%E3%81%8E"
```

環境変数（任意）: `ADDR`（既定 `:8080`）、`DATA_CSV_PATH`（コンテナ内既定 `/app/data/messages.csv`）、`CORS_ALLOWED_ORIGINS`（例: `http://localhost:3000`）。

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
