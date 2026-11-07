# セキュリティテスト計画書

「しろかぎ」のセキュリティ実験場における、デザイン・バイ・セキュリティ（Design-by-Security）に沿ったテスト自動化の導入計画を定義する。  
機能正しさだけでなく、**脆弱性・情報漏洩・サプライチェーンリスク・コンテナ構成**を継続的に検証できる状態を目標とする。

関連ドキュメント:

- デプロイ方針: [render.md](./render.md)
- Go 構成（DDD）: [go-architecture.md](./go-architecture.md)

## 目的

1. セキュアでない実装を **PR の段階で検出・ブロック**する。
2. ドメイン／ユースケースの入力検証・エラー応答を **テストコードで担保**する。
3. 依存関係とコンテナイメージの既知脆弱性を **ビルド／デプロイ前に可視化**する。
4. 低コスト（GitHub Actions / 無料枠）でも再現可能なセキュリティ検証パイプラインを実験する。

## テストの位置づけ

| 区分 | 実行タイミング | 主な手段 | 失敗時の扱い |
| --- | --- | --- | --- |
| 静的解析（SAST） | PR / `main` push | `golangci-lint` + `gosec` | Critical 相当はブロック |
| ユニット／ハンドラテスト | PR / `main` push | `go test` | 失敗でブロック |
| 依存関係スキャン | PR / リリース前 | Trivy（fs / go.mod） | Critical / High はブロック（方針確定後） |
| SBOM | PR / リリース前 | Syft（CycloneDX） | 生成失敗はブロック |
| SBOM 監査 | PR / リリース前 | Grype（SBOM 入力） | High 以上でブロック |
| コンテナスキャン | リリースデプロイ前（手動 Actions） | Trivy（image） | Critical / High はデプロイブロック |
| 構成チェック | リリースデプロイ前 | Dockerfile / イメージ検査 | non-root 未設定はブロック |

デプロイ自体は人間主導のリリースタグ＋手動 Actions（[render.md](./render.md)）とし、**検証パイプラインとデプロイパイプラインを分離**する。

## 対象範囲

### In Scope

- Go アプリケーション（`cmd/` / `internal/`）
- 同梱 CSV（仮想 DB）とその読み取り処理
- Dockerfile / コンテナイメージ
- GitHub Actions 上のセキュリティジョブ
- 依存モジュール（`go.mod` / `go.sum`）と SBOM

### Out of Scope（初期）

- 本格的な DAST（動的攻撃スキャナの常時稼働）
- ペネトレーションテストの外部委託
- Render 無料枠を超える有料セキュリティ製品の必須化
- 認証・認可の本番相当実装（将来拡張としてモック／基礎テストのみ）
- **外部のテスト／セキュリティダッシュボード**（Codecov、有料 SAST SaaS、GitHub Code scanning の本格運用など）

## テスト結果の保存・確認方針

GitHub Actions 上で実行したテストやスキャンの結果は、初期は次の **2 系統のみ**で残す。外部ダッシュボード連携は初期導入しない。

### 1. GitHub Actions のログ（コンソール出力）

- **仕組み:** `go test -v ./...` や lint / Trivy の標準出力が、Actions の Jobs 画面にリアルタイムおよび履歴として記録される。
- **用途:** 失敗理由の一次確認、どのケース／ルールで落ちたかの追跡。
- **特徴:** 追加設定なしで最も手軽。長期間の傾向分析やファイル単位の後追いには不向き。
- **運用:** verbose なテスト出力を基本とし、秘密情報はログに出さない。

### 2. GitHub Actions Artifacts（成果物保存）

- **仕組み:** テストレポートやカバレッジ、セキュリティスキャン結果を `actions/upload-artifact` でランごとに保存する。
- **用途:** 後からのダウンロード確認、SBOM / Trivy 結果とのセット保管、リリース監査の証跡。
- **保存対象（初期案）:**

| 成果物 | 例 | 主なタイミング |
| --- | --- | --- |
| カバレッジ | `coverage.out` | PR / `main` の verify |
| テスト詳細（任意） | JSON / XML レポート | 必要になった段階で追加 |
| SBOM | Syft の CycloneDX（`backend-sbom.cdx.json`） | PR / リリース前 |
| SBOM 監査レポート | Grype の出力 | PR / リリース前 |
| 脆弱性レポート | Trivy の出力ファイル | PR（依存）／リリース前（image） |

- **設定イメージ:**

```yaml
- name: Run Tests
  run: go test -v -coverprofile=coverage.out ./...

- name: Upload Test Coverage
  uses: actions/upload-artifact@v4
  with:
    name: test-coverage
    path: coverage.out
```

- **保持:** GitHub 側の Artifact 保持期間に従う（必要なら workflow で `retention-days` を明示）。無料枠の容量に留意する。

### 初期にやらないこと（外部ダッシュボード）

- Codecov 等へのカバレッジ自動送信
- 外部サービス上での推移グラフ必須化
- GitHub Security タブへの本格的な結果集約（将来検討可）

カバレッジ閾値の厳格化やダッシュボード化は、ログ＋Artifacts 運用が安定してから判断する。

## 合格基準（Definition of Done）

セキュリティテスト導入の各 Phase 完了時に、次を満たすこと。

- PR で SAST・ユニットテストが自動実行される。
- 既知の危険パターン（ハードコード秘密情報、不安全な暗号利用など）がマージ前に検知できる。
- リリースデプロイ前に、依存関係およびコンテナイメージの Critical / High がポリシーどおり扱われる。
- コンテナが非特権ユーザーで動作する方針が検査可能である。
- 失敗理由が **Actions ログ**から追跡できる。
- カバレッジおよび SBOM / スキャンレポートが **Artifacts** に残る（該当 Phase 導入後）。

---

## Phase 1: セキュアコーディングと静的解析（SAST）

コードレベルでの脆弱性や、セキュアでない実装パターンを早期検出する。

### 実装内容

- Go 静的解析に `golangci-lint` を導入する。
- セキュリティ系リントとして `gosec`（および必要に応じて関連ルール）を有効化する。
- `gosec` により、少なくとも次を静的検出する。
  - ハードコードされた認証情報・秘密情報
  - 不安全な暗号化アルゴリズム・乱数利用
  - インジェクションの兆候（SQL / コマンド等）
  - 不安全なファイルパス結合や権限設定の疑い

### CI/CD 統合

- **トリガー:** Pull Request、および `main` への push
- **動作:** 警告／エラーを検知したらジョブ失敗（マージブロック）
- **設定例:** `.golangci.yml` で `gosec` を enable、重大度に応じた fail 閾値を定義

### 成果物

| 成果物 | 説明 |
| --- | --- |
| `.golangci.yml` | リント／gosec 設定 |
| GitHub Actions（verify 系） | PR 時の `golangci-lint` 実行 |
| ドキュメント更新 | ローカル実行手順（任意） |

### 受け入れ条件

- [ ] PR で `golangci-lint` が必須チェックになる
- [ ] 意図的に埋め込んだ insecure サンプルが検知されること（検知確認用の一時検証可）
- [ ] 正当なコードで誤検知が過大でないこと（必要なら除外を最小化して管理）

---

## Phase 2: ユニットテストにおけるセキュリティ検証

ビジネスロジックや入力値検証がセキュアに機能することを、テストコードで担保する。  
DDD 構成では、主に `domain` / `usecase` / `handler` にテストを配置する（[go-architecture.md](./go-architecture.md)）。

### 実装内容

#### インジェクション対策テスト

- 不正入力（SQL 断片、特殊文字、過長文字列、制御文字など）を与えても、期待どおり拒否または無害化されることを検証する。
- 本プロジェクトは CSV 仮想 DB のため、RDB のプリペアドステートメントに加え、**CSV パース／パス操作／式展開に相当する不安全な扱いが無いこと**を重点確認する。
- 将来 RDB を再導入する場合に備え、リポジトリ IF 経由のテストで「文字列連結クエリを書かない」方針を維持できる形にしておく。

#### アクセス制御・認証テスト（将来拡張の基礎）

- 現 PoC に本格認証は含めない。
- ただしハンドラ／ユースケース境界で、未認証・不正トークンを想定した **モックベースの拒否テスト**を先に置ける構造にする。
- 秘密情報をテストコードへ直書きしない（フィクスチャはダミー値に限定）。

#### エラーハンドリング・情報漏洩防止テスト

- エラー時レスポンスに、次が含まれないことを検証する。
  - スタックトレース
  - 内部ファイルパス
  - 環境変数・秘密情報
  - CSV の生データや実装詳細の過剰な露出
- HTTP ハンドラテストで、クライアント向けメッセージが汎用的であることを確認する。

### CI/CD 統合

- `go test -v -coverprofile=coverage.out ./...` を PR 必須チェックに含める。
- 標準出力は Actions ログに残す。`coverage.out` は Artifact としてアップロードする。
- カバレッジ閾値の厳格化や外部ダッシュボード連携は後続（まずはケース網羅と結果の保存を優先）。

### 成果物

| 成果物 | 配置例 |
| --- | --- |
| ドメイン／バリデーションテスト | `internal/domain/.../*_test.go` |
| ユースケースの不正入力テスト | `internal/usecase/.../*_test.go` |
| ハンドラの漏洩防止テスト | `internal/handler/http/*_test.go` |
| カバレッジ Artifact | Actions 上の `test-coverage`（`coverage.out`） |

### 受け入れ条件

- [ ] 代表的な不正入力ケースが自動化されている
- [ ] エラー応答の情報漏洩チェックがハンドラテストにある
- [ ] 認証まわりはモックで「拒否できる」基礎が残っている（未実装でもスキップ方針を明記可）
- [ ] テスト実行がログに残り、カバレッジが Artifact 保存される

---

## Phase 3: 依存関係・サプライチェーンのスキャン（再確認と拡張）

企画初期で導入する Trivy を活用し、テスト／ビルド工程でサプライチェーンリスクを継続監視する。

### 実装内容

- Go モジュール（`go.mod` / `go.sum`）の既知脆弱性（CVE）を Trivy でスキャンする。
- **Syft** で SBOM（CycloneDX JSON）を生成する（Verify: `backend/`、Pre-Deploy: コンテナイメージ）。
- **Grype** にその SBOM を渡し、脆弱性を突き合わせて監査する（High 以上で失敗）。
- Trivy による直接スキャン（fs / image）も併用し、多層で確認する。
- 生成物（SBOM・Grype・Trivy）は Actions の Artifact として保管する。

### CI/CD 統合

| タイミング | 内容 |
| --- | --- |
| PR | Trivy FS + Syft SBOM + Grype（SBOM 監査）。Critical / High で失敗 |
| リリースデプロイ前 | イメージ向けに Syft SBOM 再生成 + Grype 監査 + Trivy image |
| 保管 | SBOM / Grype / Trivy レポートを Artifact 保存 |

### 合格／失敗ポリシー（初期案）

| 重要度 | PR | リリースデプロイ |
| --- | --- | --- |
| Critical | ブロック | ブロック |
| High | ブロック | ブロック |
| Medium 以下 | 警告（ログ） | 警告（ログ） |

例外許可が必要な場合は、理由・期限・追跡 Issue を残す運用とする（黙殺しない）。

### 受け入れ条件

- [ ] PR で依存関係スキャンが走る
- [x] SBOM が Syft で生成・保存される（Verify: path / Pre-Deploy: image）
- [x] Grype による SBOM 突き合わせ監査が入っている（High 以上で失敗）
- [x] Critical / High 検出時にジョブが失敗する（Trivy / Grype）

---

## Phase 4: コンテナ・インフラ層のセキュリティテスト

Render へデプロイする直前のコンテナイメージに対し、セキュリティ基準を検証する。

### 実装内容

- Trivy 等によるコンテナイメージの OS パッケージ／ライブラリ脆弱性スキャン。
- 非特権ユーザー（Non-root）でコンテナが動作することの構成チェック。
  - Dockerfile で `USER` 指定があること
  - イメージ／実行設定で uid=0 でないこと（検査コマンドまたはポリシーチェック）
- 可能なら、不要なパッケージ削減・最小ベースイメージ採用をスキャン結果と合わせて改善する。

### CI/CD 統合

- **トリガー:** リリースタグ指定の手動デプロイ Actions（`workflow_dispatch`）内、Deploy Hook 呼び出し **前**
- イメージビルド → スキャン／構成チェック合格 → Render デプロイ、の順序を守る
- スキャン結果は Jobs ログに出し、レポートファイルは Artifact に保存してからデプロイ可否を判定する

### 受け入れ条件

- [ ] デプロイ前にイメージスキャンが必須である
- [ ] non-root で起動する構成が検査される
- [ ] Critical / High があるイメージはデプロイに進まない
- [ ] イメージスキャンレポートが Artifact に残る

---

## ロードマップと優先度

```text
Phase 1  SAST（golangci-lint / gosec）     … 最初に導入
    ↓
Phase 2  セキュリティ観点のユニットテスト … アプリ実装と並行
    ↓
Phase 3  Trivy + Syft SBOM + Grype 監査    … PR / リリースに拡張
    ↓
Phase 4  コンテナスキャン + non-root 検証  … デプロイ直前ゲート
```

| Phase | 目安 | 依存 |
| --- | --- | --- |
| 1 | Go リポジトリ初期化直後 | Actions の verify ワークフロー |
| 2 | PoC API（`GET /`）実装と同時 | domain / usecase / handler |
| 3 | Phase 1 安定後 | Trivy / Syft の Actions 導入 |
| 4 | Dockerfile 作成後 | 手動デプロイワークフロー |

## ツール一覧（予定）

| 目的 | ツール |
| --- | --- |
| Lint / SAST | `golangci-lint`, `gosec` |
| ユニットテスト | `go test`, `testing` / `net/http/httptest` |
| 結果の一次確認 | GitHub Actions Jobs ログ |
| 結果ファイルの保管 | `actions/upload-artifact`（coverage / SBOM / Trivy 等） |
| 依存・イメージスキャン | Trivy |
| SBOM 生成 | Syft（`anchore/sbom-action`） |
| SBOM 監査 | Grype（`anchore/scan-action`、SBOM 入力） |
| CI | GitHub Actions |
| デプロイ先 | Render.com（Web Service、CSV 仮想 DB） |

初期は Codecov 等の外部ダッシュボードは使わない。

## リスクと留意事項

- 無料枠・実行時間制限により、重いスキャンはリリースパイプライン側へ寄せる場合がある。
- Artifact の容量・保持期間にも無料枠制約があるため、巨大な中間成果物は置かない。
- gosec の誤検知は、安易な全面 disable ではなくルール単位・行単位で最小化して管理する。
- CSV 正本はリポジトリ管理のため、**機微データを CSV に置かない**（テスト用もダミーデータ）。
- Secrets（Deploy Hook 等）はテストログおよび Artifact へ出力しない。

## 今後の実装メモ

- [ ] `.golangci.yml` と PR 用 verify workflow（Phase 1）
- [ ] セキュリティ観点の `*_test.go`（Phase 2）
- [ ] `go test -v -coverprofile` のログ出力と `coverage.out` の Artifact アップロード
- [x] Trivy + Syft SBOM + Grype 監査ジョブと Artifact 保存（Phase 3）
- [ ] Dockerfile の non-root 化とイメージスキャン（Phase 4）
- [ ] Critical / High ブロックポリシーの Actions への実装
- [ ] README へのテスト／セキュリティゲート概要の追記
- [ ] （将来）外部ダッシュボードやカバレッジ閾値の要否判断
