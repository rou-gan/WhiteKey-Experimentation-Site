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
[リリース準備] main の先端（または特定コミット）にタグ付与
                例: v0.00.01
                ↓
[人間] GitHub Actions を手動実行（workflow_dispatch）
       ・入力: デプロイ対象タグ（例: v0.00.01）
                ↓
[GitHub Actions]
       ・タグの存在 / main 由来であることの確認
       ・（必要に応じて）Lint / Test / Security Scan / SBOM
       ・Render Deploy Hook または Render API でデプロイ要求
                ↓
[Render.com]
       ・指定リビジョン（タグ）で Web Service をビルド・起動
       ・CSV（仮想 DB）からデータを読み出し API を提供
```

## Render.com 上の構成

### リソース

| リソース | 用途 |
| --- | --- |
| Web Service | Go API Server（Docker コンテナ） |
| CSV ファイル（仮想 DB） | `"Hello World"` メッセージなどのデータ永続化 |

PostgreSQL などのマネージド DB は **採用しない**（理由は後述）。

### Web Service（想定設定）

| 項目 | 方針 |
| --- | --- |
| Runtime | Docker |
| Region | 実験コスト・レイテンシを考慮して選択（後で固定） |
| Branch / Commit | 通常の「main 追従の自動デプロイ」は **無効** |
| デプロイ契機 | GitHub Actions からの Deploy Hook / API 呼び出し |
| データストア | コンテナ内（またはリポジトリ同梱）の CSV ファイル |
| 環境変数 | CSV パスなど必要最小限（例: `DATA_CSV_PATH`）。DB 接続文字列は不要 |

### データ永続化（CSV 仮想 DB）

PoC では RDB の代わりに CSV を「仮想 DB」として扱う。

- 初期データ（例: `message = "Hello World"`）は CSV としてリポジトリまたはコンテナイメージに同梱する。
- API（`GET /`）は起動時またはリクエスト時に CSV を読み取り、値を返す。
- スキーマは最小構成とする（例: `id,message`）。
- 書き込みが必要になった場合も、まずは同一 CSV への更新を想定する。  
  ※ Render 無料 Web Service のローカルディスクは再デプロイで消えるため、**正本はリポジトリ／イメージ内の CSV** とし、ランタイム書き込みだけに依存しない。

### PostgreSQL を採用しない理由

| 理由 | 内容 |
| --- | --- |
| 無料枠の寿命 | Render.com の PostgreSQL 無料プランは **約 30 日でインスタンスが削除**される。実験場として継続利用しづらい。 |
| コスト方針 | 本プロジェクトは無料枠・低コストでの検証が前提。有料 DB を常時維持する段階ではない。 |
| PoC のデータ量 | 初期要件は `"Hello World"` の保持・返却のみ。RDB の機能（トランザクション、同時書き込み等）は不要。 |
| 構成の単純化 | DB インスタンス・接続文字列・マイグレーション・ネットワーク設定を省き、Web Service 単体で完結させる。 |
| 再現性 | データ定義を CSV としてリポジトリ管理でき、タグ単位で「どのデータ付きビルドか」を追いやすい。 |

将来、本格的な永続化や複数インスタンス共有が必要になった段階で、有料 PostgreSQL や外部ストレージの再検討を行う。

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

1. **Web Service** を作成する（Docker、リポジトリ連携）※ PostgreSQL は作成しない
2. **Auto-Deploy を Off** にする（push 連動デプロイを使わない）
3. 必要なら CSV パスなどの環境変数を設定する（例: `DATA_CSV_PATH`）
4. **Deploy Hook URL** を発行する
5. Deploy Hook URL を GitHub リポジトリの Secrets に登録する  
   - 例: `RENDER_DEPLOY_HOOK_URL`

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

3. 必要なら GitHub Releases で同タグの Release を作成する（リリースノート用。デプロイ自体はタグ + Actions が本体）
4. GitHub Actions のデプロイ用ワークフローを **手動実行** し、入力にタグ（例: `v0.00.01`）を指定する
5. Actions の成功と、Render ダッシュボード上のデプロイ完了を確認する
6. `GET /` が期待どおり `"Hello World"`（または JSON）を返すことを確認する

## セキュリティ上の注意

- Render の API Key / Deploy Hook は GitHub Actions の Secrets / Environments のみに置く。
- 本番相当 Environment を使う場合は、手動承認（Required reviewers）の導入を検討する。
- タグ付けとデプロイ実行の権限は、リポジトリ管理者またはリリース担当に限定する。
- スキャン（Trivy 等）で Critical / High がある場合のデプロイ可否は、別途ポリシーとしてワークフローに組み込む。

## 今後の実装メモ

本ドキュメントは設計メモである。実装時には次をリポジトリへ追加する想定。

- [ ] Render Web Service の作成手順の確定（リージョン・プラン）。PostgreSQL は使わない
- [x] `Dockerfile` および CSV 読み取りによる API 実装
- [x] 初期データ用 CSV（`backend/data/messages.csv`）の追加
- [x] GitHub Actions: PR 用の検証ワークフロー（`backend-verify.yml`）
- [x] GitHub Actions: デプロイ前のビルド／スキャン（`backend-pre-deploy.yml`、Render デプロイは未接続）
- [ ] GitHub Actions: Render へのデプロイステップ追加（Deploy Hook）
- [ ] `RENDER_DEPLOY_HOOK_URL`（および必要なら Render API 用 Secret）の設定
- [x] README を「手動リリースタグデプロイ」「CSV 仮想 DB」に合わせて更新
