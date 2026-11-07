# 「しろかぎ」のセキュリティ実験場

デザイン・バイ・セキュリティ（Design-by-Security）の考え方に基づき、コンテナや依存関係の脆弱性管理、サプライチェーンセキュリティ（SBOM）、セキュリティコンプライアンス検証（SCAP）を統合的に実践・検証するための実験場。

本番運用を想定したセキュアな CI/CD パイプラインおよび監視・監査プロセスを、Render.com および GitHub Actions 環境でローコスト（無料枠・低コスト）に検証することを目的とする。

## システムアーキテクチャ

```
[GitHub Repository]
       ↓ (Push / Pull Request)
[GitHub Actions CI/CD Pipeline]
       ├── Trivy (コンテナ/依存関係スキャン)
       ├── SBOM生成 (Syft等)
       ├── SCAPベースのコンプライアンスチェック
       ↓ (自動ビルド＆デプロイ)
[Render.com]
       ├── Go API Server (Dockerコンテナ)
       └── PostgreSQL (DB: "Hello World" の保持)
```

### 技術スタック

| 領域 | 採用技術 |
| --- | --- |
| ホスティング | Render.com (Web Service / PostgreSQL) |
| 言語・フレームワーク | Go (`net/http` または軽量ルーター、コンテナデプロイ) |
| バージョン管理・CI/CD | GitHub / GitHub Actions |
| セキュリティツール | Trivy, Syft/Grype (SBOM生成), OpenSCAP / SCAP関連ツール |

## セキュリティ検証項目

本環境では、プロダクト運用に直結するセキュリティプラットフォームの検証を行う。

### Trivy による多層脆弱性スキャン

- **コードスキャン:** ソースコードの静的解析（SAST）およびシークレット検出。
- **コンテナスキャン:** Docker イメージビルド時の OS パッケージおよび Go バイナリの脆弱性検出。
- **CI/CD 統合:** 脆弱性の重要度（Critical / High）に応じたビルドブロックの仕組みの検証。

### SBOM（Software Bill of Materials）の生成と管理

- **生成:** ビルドプロセス内で Syft 等を用いて SBOM（SPDX / CycloneDX 形式）を自動生成。
- **監査:** 依存ライブラリのバージョン追跡と、脆弱性アラート連動のテスト。

### SCAP（Security Content Automation Protocol）の適用検証

- セキュリティ設定やコンプライアンスベースラインの自動評価手法の検証。
- インフラ設定やコンテナのセキュリティベンチマーク（CIS Benchmark 等）との突合実験。

## デプロイメントパイプライン（GitHub Actions）

1. **トリガー:** `main` ブランチへのマージ（または PR 作成時）
2. **ステップ:**
   - **Lint & Test:** Go の静的解析（`golangci-lint`）とユニットテスト
   - **Security Scan:** Trivy によるファイルシステムおよびコンテナイメージのスキャン
   - **SBOM Generation:** SBOM アーティファクトの生成と保存
   - **Deploy to Render:** Render API / Deploy Hook を利用した自動デプロイ

## 初期リリース仕様（PoC スコープ）

### 機能要件

- **エンドポイント:** `GET /`
- **動作仕様:**
  1. Render 上の PostgreSQL に接続する。
  2. 事前登録された `"Hello World"` 文字列を DB から取得する。
  3. JSON 形式またはプレーンテキストとしてレスポンスを返却する。

### データベース初期化

マイグレーションスクリプトまたは初回起動時処理により、テーブル作成と初期データ（`message = "Hello World"`）のインサートを自動化する。

## 検証ロードマップ

- **Phase 1（初期リリース）:** Docker 化、GitHub Actions による Render への自動デプロイ、Trivy スキャンの導入。
- **Phase 2（サプライチェーン強化）:** SBOM の自動生成・アーティファクト保管の仕組み化。
- **Phase 3（コンプライアンス拡張）:** SCAP / セキュリティポリシーのコード化（Policy as Code）および監査ログの収集実験。
