# Render.com デプロイ設計

「しろかぎ」のセキュリティ実験場における、Render.com 上のホスティング構成と、GitHub Actions 経由のデプロイ方針をまとめる。

## 方針

- **人間主導のデプロイ:** `main` へのマージだけでは Render へデプロイしない。
- **リリースタグ起点:** `main` から作成したリリースタグをデプロイ対象とする。
- **実行手段:** GitHub Actions を人が明示的に起動し、指定したタグを Render.com へデプロイする。

本番相当の「いつ何を出すか」を人が判断し、CI は検証・ビルド・デプロイ実行を担う。

## バージョンタグ規約

タグは次の形式とする。

```text
vMAJOR.MINOR.PATCH
```

例:

| タグ | 意味の例 |
| --- | --- |
| `v0.00.01` | 初回 PoC / 極小変更 |
| `v0.01.00` | 機能追加を含むマイナー更新 |
| `v1.00.00` | 安定版としてのメジャー区切り |

ルール:

- 先頭は必ず `v` を付ける。
- 数値はゼロ埋め 2 桁（`v0.00.00` 形式）を基本とする。
- タグは **必ず `main` ブランチのコミット** から作成する。
- 同一タグの付け直し（移動・削除後の再利用）は原則行わない。

## 全体フロー

```text
[開発] feature ブランチ → PR → main へマージ
                ↓
[リリース準備] main の先端（または特定コミット）にタグ付与して push
                例: v0.00.01
                ↓
[GitHub Actions — Backend / Frontend Release Assets]（同一タグ push で並列）
       ・タグ形式 / main 由来であることの確認
       ・各サービスの test/build / Docker / non-root
       ・Trivy image / Syft SBOM / Grype 監査
       ・同一 GitHub Release に各サービスの成果物を添付
                ↓
[人間] Backend / Frontend Pre-Deploy をそれぞれ手動実行（デプロイ前ゲート）
       ・入力: デプロイ対象タグ（例: v0.00.01）
       ・将来: 各 Deploy Hook / API でデプロイ要求
                ↓
[Render.com]
       ・Web Service（backend）: Go API / PostgreSQL
       ・Web Service（frontend）: Next.js standalone（別コンテナ）
       ・Postgres（Free）: whitekey-db
```

同一タグで backend / frontend の両方をリリースする。サービスは **別コンテナ・別 Web Service・別 Deploy Hook**。

### Release に添付する成果物（タグ向けに再生成）

| 資産名の例 | 内容 |
| --- | --- |
| `v0.00.01-backend-sbom.cdx.json` | Syft による CycloneDX SBOM（backend） |
| `v0.00.01-backend-grype-sbom.txt` | Grype による SBOM 監査結果（backend） |
| `v0.00.01-backend-trivy-image.txt` | Trivy イメージスキャン結果（backend） |
| `v0.00.01-frontend-sbom.cdx.json` | Syft による CycloneDX SBOM（frontend） |
| `v0.00.01-frontend-grype-sbom.txt` | Grype による SBOM 監査結果（frontend） |
| `v0.00.01-frontend-trivy-image.txt` | Trivy イメージスキャン結果（frontend） |

PR の Actions Artifact は短期保管用。Release 資産はタグ時点の正本として長く残す。

## Render.com 上の構成

### リソース

| リソース | 用途 |
| --- | --- |
| Web Service（backend） | Go API Server（Docker コンテナ） |
| Web Service（frontend） | Next.js standalone（**別** Docker コンテナ） |
| Render Postgres（Free） | `messages` 等のアプリデータ（リソース名 `whitekey-db`） |

インフラ定義の正本はリポジトリルートの [`render.yaml`](../render.yaml)（Blueprint）。`DATABASE_URL` は手貼りせず `fromDatabase` で注入する。

### 固定名（作り直し時も変えない）

| 項目 | 値 |
| --- | --- |
| Postgres リソース名 | `whitekey-db` |
| databaseName | `whitekey` |
| user | `whitekey` |
| Region | `oregon`（Web Service と同じ。初回選定後は変更しない） |
| Compute plan | `free`（実験用。期限切れ前提） |

有料化すれば期限切れは避けられる（入口はおおよそ $6/月〜）。本実験場は Free 再作成を許容する。

### Web Service — backend（想定設定）

| 項目 | 方針 |
| --- | --- |
| Runtime | Docker（`backend/Dockerfile`） |
| Region | `oregon`（DB と同一） |
| Branch / Commit | `autoDeployTrigger: off`（Blueprint）。通常の main 追従自動デプロイは無効 |
| デプロイ契機 | GitHub Actions からの Deploy Hook / API 呼び出し |
| データストア | Render Postgres（`DATABASE_URL`） |
| 環境変数 | `DATABASE_URL`（`fromDatabase`）、`CORS_ALLOWED_ORIGINS` 等 |

### Web Service — frontend（想定設定）

| 項目 | 方針 |
| --- | --- |
| Runtime | Docker（Next.js `standalone`） |
| Auto-Deploy | **Off**（backend と同じ） |
| デプロイ契機 | Frontend Pre-Deploy → Deploy Hook（準備中） |
| 環境変数 | `API_BASE_URL`（backend の公開 URL）、`PORT=3000`、`HOSTNAME=0.0.0.0` |
| Secrets | `RENDER_FRONTEND_DEPLOY_HOOK_URL`（backend の Hook とは分離） |

詳細は [frontend-architecture.md](./frontend-architecture.md)。

### データ永続化（PostgreSQL）

- スキーマと初期シードの正本はアプリ埋め込みマイグレーション（goose）。配置: `backend/internal/infrastructure/persistence/migrations/`。
- backend 起動時に `DATABASE_URL` へ接続し、migrate up したうえでリッスンする。
- 初回シード: `messages` に `id=1` / `content=Hello World`。
- Free Postgres は約 30 日で期限切れするため、**データ消失を前提**に運用する。スキーマは再デプロイで復元する。

### Free Postgres 作り直しランブック

データが消えてよい前提。接続文字列のコピペやスキーマ手打ちはしない。

1. 期限切れメールを確認する（作成から約 30 日。その後 14 日の猶予あり）。
2. 猶予終了後、または明示削除後、Dashboard で **同じ名前** `whitekey-db` の Free Postgres を新規作成する（region=`oregon`、databaseName / user は Blueprint どおり）。
3. Blueprint（`render.yaml`）を sync し、backend の `DATABASE_URL` が新 DB に再紐付けされることを確認する。
4. backend を再デプロイする（現行の Deploy Hook / Pre-Deploy 運用）。
5. 起動時 migrate により `messages` テーブルと Hello World シードが復元されることを `GET /` で確認する。

手作業の中心は「Free DB の新規作成」のみ。ワークスペースあたり Free Postgres は 1 つ。

## GitHub Actions との役割分担

| フェーズ | トリガー | 目的 |
| --- | --- | --- |
| PR / `main` 更新時 | push / pull_request | Lint・Test・脆弱性スキャンなど **検証**（デプロイしない） |
| リリースデプロイ | `workflow_dispatch`（手動） | 指定タグを Render へ **デプロイ** |

デプロイ用ワークフローの想定入力:

| 入力名 | 例 | 説明 |
| --- | --- | --- |
| `tag` | `v0.00.01` | デプロイするリリースタグ |

ワークフロー側で最低限確認すること:

1. 入力タグが `v[0-9]{1,2}\.[0-9]{2}\.[0-9]{2}` 形式であること
2. そのタグがリポジトリに存在すること
3. タグが指すコミットが `main` の履歴上にあること
4. 確認後に Render の Deploy Hook（または API）を呼び出すこと

## Render 側の準備チェックリスト

ダッシュボードで次を用意する。

1. Blueprint で [`render.yaml`](../render.yaml) を接続し、`whitekey-db` と Web Service（backend / frontend）を作成する
2. 両方の Web Service とも **Auto-Deploy を Off**（`autoDeployTrigger: off`）にする
3. backend: `DATABASE_URL` は `fromDatabase`（手貼りしない）。CORS オリジン等。frontend: `API_BASE_URL` 等
4. 各サービスで **Deploy Hook URL** を発行する
5. GitHub Secrets に登録する  
   - backend: `RENDER_DEPLOY_HOOK_URL`  
   - frontend: `RENDER_FRONTEND_DEPLOY_HOOK_URL`

> Deploy Hook は秘密情報として扱い、Issue・PR・ドキュメント本文には書かない。

## 人間が行うリリース手順

1. `main` がデプロイしたい状態であることを確認する（CI 緑、レビュー済み）
2. ローカルまたは GitHub UI でタグを作成し、リモートへ push する

```bash
git checkout main
git pull origin main
git tag v0.00.01
git push origin v0.00.01
```

3. タグ push により `Backend Release Assets` が走り、GitHub Release とセキュリティ成果物が自動作成されることを確認する
4. GitHub Actions の Pre-Deploy（将来はデプロイ用）ワークフローを **手動実行** し、入力にタグ（例: `v0.00.01`）を指定する
5. Actions の成功と、Render 接続後はダッシュボード上のデプロイ完了を確認する
6. `GET /` が期待どおり `"Hello World"`（または JSON）を返すことを確認する

## セキュリティ上の注意

- Render の API Key / Deploy Hook は GitHub Actions の Secrets / Environments のみに置く。
- 本番相当 Environment を使う場合は、手動承認（Required reviewers）の導入を検討する。
- タグ付けとデプロイ実行の権限は、リポジトリ管理者またはリリース担当に限定する。
- スキャン（Trivy 等）で Critical / High がある場合のデプロイ可否は、別途ポリシーとしてワークフローに組み込む。
- DB 接続文字列（`DATABASE_URL`）はログやエラー応答に出さない。

## 今後の実装メモ

- [ ] Render Blueprint / Web Service（backend / frontend）と Free Postgres（`whitekey-db`）の接続確定
- [x] backend Postgres 永続化・起動時 migrate・ローカル docker compose
- [x] GitHub Actions: PR 用の検証ワークフロー（`backend-verify.yml`、Postgres サービス付き）
- [x] GitHub Actions: デプロイ前のビルド／スキャン（`backend-pre-deploy.yml`、Render デプロイは未接続）
- [x] GitHub Actions: タグ push で Release へセキュリティ成果物を添付（`backend-release.yml`）
- [x] frontend Next.js スキャフォールド / Dockerfile / Verify・Release・Pre-Deploy
- [ ] GitHub Actions: Render へのデプロイステップ追加（backend / frontend Deploy Hook）
- [ ] `RENDER_DEPLOY_HOOK_URL` / `RENDER_FRONTEND_DEPLOY_HOOK_URL` の設定
- [x] README / アーキテクチャを PostgreSQL 方針に合わせて更新
