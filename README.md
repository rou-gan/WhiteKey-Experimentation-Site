# 「しろかぎ」のセキュリティ実験場

デザイン・バイ・セキュリティ（Design-by-Security）の考え方に基づき、コンテナや依存関係の脆弱性管理、サプライチェーンセキュリティ（SBOM）、セキュリティコンプライアンス検証（SCAP）を統合的に実践・検証するための実験場。

本番運用を想定したセキュアな CI/CD パイプラインおよび監視・監査プロセスを、Render.com および GitHub Actions 環境でローコスト（無料枠・低コスト）に検証することを目的とする。

## リポジトリ構成

モノレポ構成。Go API は `backend/`、将来のフロントエンドは `frontend/` に分離する。

```text
WhiteKey-Experimentation-Site/
├── backend/                 # Go API（DDD 風レイヤー、Docker、CSV 仮想 DB）
│   └── openapi/             # OpenAPI 仕様
├── frontend/                # フロントエンド（準備中）
├── docs/                    # 設計ドキュメント
└── .github/workflows/       # backend 用 GitHub Actions
```

詳細設計:

- デプロイ方針: [docs/render.md](docs/render.md)
- Go アーキテクチャ: [docs/go-architecture.md](docs/go-architecture.md)
- セキュリティテスト計画: [docs/security-test-plan.md](docs/security-test-plan.md)
- OpenAPI: [backend/openapi/openapi.yaml](backend/openapi/openapi.yaml)

## システムアーキテクチャ

```text
[GitHub Repository]
       ↓ (PR / push: backend/**)
[GitHub Actions — Backend Verify]
       ├── golangci-lint (+ gosec)
       ├── go test (+ coverage Artifact)
       ├── Trivy filesystem scan
       ├── Syft SBOM (CycloneDX)
       └── Grype による SBOM 突き合わせ監査
                ↓
[人間] main からリリースタグ作成・push (例: v0.00.01)
       ↓
[GitHub Actions — Backend Release Assets]（タグ push で自動）
       ├── タグ形式 / main 由来の確認
       ├── test / image build / non-root
       ├── Trivy image / Syft SBOM / Grype 監査
       └── GitHub Release に成果物を添付
                ↓
[人間] Backend Pre-Deploy を手動実行（デプロイ前ゲート）
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
| セキュリティツール | golangci-lint / gosec、Trivy、Syft（SBOM）、Grype（SBOM 監査）。SCAP は今後 |

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
| `Backend Verify` | PR / push（`backend/**`） | Lint、Test、Trivy FS、Syft SBOM、Grype 監査、coverage Artifact |
| `Backend Release Assets` | タグ push（`v*.*.*`） | タグ検証、image ビルド／スキャン、**GitHub Release へ成果物添付** |
| `Backend Pre-Deploy` | 手動（`workflow_dispatch`） | デプロイ前ゲート（**Render デプロイは未接続**） |

デプロイ方針:

- `main` へのマージだけでは Render へデプロイしない。
- リリースタグ（`v0.00.00` 形式）を `main` から作成して push すると、セキュリティ成果物付き Release が自動作成される。
- Render へのデプロイは、将来 Pre-Deploy（または同等）を手動実行して行う。

テスト／スキャン結果は Actions のログ・Artifacts、およびリリースタグ時は Release 資産に残す（外部ダッシュボードは初期導入しない）。詳細は [docs/security-test-plan.md](docs/security-test-plan.md) / [docs/render.md](docs/render.md)。

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
- **Phase 2:** Syft SBOM + Grype 突き合わせ監査の定着、依存スキャンの継続監視
- **Phase 3:** SCAP / Policy as Code、監査ログ収集の実験
