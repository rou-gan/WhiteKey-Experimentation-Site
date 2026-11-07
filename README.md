# 「しろかぎ」のセキュリティ実験場

デザイン・バイ・セキュリティ（Design-by-Security）の考え方に基づき、コンテナや依存関係の脆弱性管理、サプライチェーンセキュリティ（SBOM）、セキュリティコンプライアンス検証（SCAP）を統合的に実践・検証するための実験場。

本番運用を想定したセキュアな CI/CD パイプラインおよび監視・監査プロセスを、Render.com および GitHub Actions 環境でローコスト（無料枠・低コスト）に検証することを目的とする。

## リポジトリ構成

モノレポ構成。Go API は `backend/`、将来のフロントエンドは `frontend/` に分離する。

```text
WhiteKey-Experimentation-Site/
├── backend/                 # Go API（DDD 風レイヤー、Docker、CSV 仮想 DB）
├── frontend/                # フロントエンド（準備中）
├── docs/                    # 設計ドキュメント
└── .github/workflows/       # backend 用 GitHub Actions
```

詳細設計:

- デプロイ方針: [docs/render.md](docs/render.md)
- Go アーキテクチャ: [docs/go-architecture.md](docs/go-architecture.md)
- セキュリティテスト計画: [docs/security-test-plan.md](docs/security-test-plan.md)

## システムアーキテクチャ

```text
[GitHub Repository]
       ↓ (PR / push: backend/**)
[GitHub Actions — Backend Verify]
       ├── golangci-lint (+ gosec)
       ├── go test (+ coverage Artifact)
       ├── Trivy filesystem scan
       └── SBOM Artifact (CycloneDX)
                ↓
[人間] リリースタグ作成 (例: v0.00.01) + Actions 手動実行
       ↓
[GitHub Actions — Backend Pre-Deploy]
       ├── go test
       ├── Docker image build
       ├── non-root 構成チェック
       ├── Trivy image scan
       └── SBOM Artifact
                ↓ （Render 接続後）
[Render.com]
       └── Go API Server (Docker / CSV 同梱)
```

PostgreSQL などのマネージド DB は採用しない（無料枠の寿命・低コスト方針のため）。データはリポジトリ同梱の CSV を仮想 DB とする。詳細は [docs/render.md](docs/render.md)。

### 技術スタック

| 領域 | 採用技術 |
| --- | --- |
| ホスティング | Render.com（Web Service）。デプロイ接続は準備中 |
| Backend | Go（`net/http`）、Docker、CSV 仮想 DB |
| Frontend | 準備中（`frontend/`） |
| バージョン管理・CI/CD | GitHub / GitHub Actions（backend 検証は導入済み） |
| セキュリティツール | golangci-lint / gosec、Trivy、SBOM（CycloneDX）。SCAP は今後 |

## セキュリティ検証項目

### Trivy による多層脆弱性スキャン

- **依存・ファイルシステム:** PR 時に `backend/` をスキャン（Critical / High で失敗）。
- **コンテナスキャン:** 手動 Pre-Deploy でイメージをスキャン。
- **CI/CD 統合:** 重要度に応じたビルドブロックを検証中。

### SBOM（Software Bill of Materials）

- **Backend Verify（PR / push）:** `backend/` のファイルシステムから CycloneDX SBOM を生成し Artifact 保存する。
- **Backend Pre-Deploy（手動）:** コンテナイメージから CycloneDX SBOM を生成し Artifact 保存する。

### SCAP（Security Content Automation Protocol）

- 今後の拡張項目。セキュリティ設定や CIS Benchmark 等との突合実験を予定。

## GitHub Actions（現状）

| ワークフロー | 起動 | 内容 |
| --- | --- | --- |
| `Backend Verify` | PR / push（`backend/**`） | Lint、Test、Trivy FS、SBOM（CycloneDX）、coverage Artifact |
| `Backend Pre-Deploy` | 手動（`workflow_dispatch`） | イメージビルド・non-root・Trivy image・SBOM（**Render デプロイは未接続**） |

デプロイ方針:

- `main` へのマージだけでは Render へデプロイしない。
- 将来は `main` から切ったリリースタグ（`v0.00.00` 形式）を手動 Actions でデプロイする。

テスト結果は Actions のログおよび Artifacts に残す（外部ダッシュボードは初期導入しない）。詳細は [docs/security-test-plan.md](docs/security-test-plan.md)。

## 初期リリース仕様（PoC）

### 機能要件

- **エンドポイント:** `GET /`
- **動作仕様:**
  1. 同梱 CSV（`backend/data/messages.csv`）からメッセージを取得する。
  2. `"Hello World"` を JSON（`{"message":"Hello World"}`）で返す。

### ローカル起動（Docker）

```bash
cd backend
docker compose up --build
curl http://localhost:8080/
```

環境変数（任意）: `ADDR`（既定 `:8080`）、`DATA_CSV_PATH`（コンテナ内既定 `/app/data/messages.csv`）。

## 検証ロードマップ

- **Phase 1:** Docker 化、backend の Verify / Pre-Deploy Actions、Trivy 導入（進行中。Render デプロイ接続が残作業）
- **Phase 2:** SBOM 運用の定着、依存スキャンの継続監視
- **Phase 3:** SCAP / Policy as Code、監査ログ収集の実験
