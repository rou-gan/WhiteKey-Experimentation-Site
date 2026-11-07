# セキュリティ Agent 脅威モデル（ローカル `/review-security`）

「しろかぎ」実験場向けに、ローカル Cursor の `/review-security`（`security-review` サブエージェント）が優先して見る観点を定義する。  
確定的ゲート（`Backend Verify` の lint / test / Trivy / Syft / Grype）は置き換えない。Agent レビューは **advisory** とする。

関連:

- セキュリティテスト計画: [security-test-plan.md](./security-test-plan.md)
- Go 構成: [go-architecture.md](./go-architecture.md)

## 運用タイミング

- **実行する:** push または Pull Request 作成の**直後**（手元の Cursor で `/review-security`）
- **実行しない:** PR 作成前
- **結果の残し方:** 該当 PR に `gh pr review --comment` でレビューコメント（`APPROVE` / `REQUEST_CHANGES` は使わない）

## 優先して見るもの

| 優先度 | 観点 | 典型例 |
| --- | --- | --- |
| High | 反射型 XSS | HTML 埋め込み前のエスケープ欠落（`/experiment/greet` や handler の HTML 応答） |
| High | 情報漏洩 | エラー応答へのスタックトレース、内部パス、秘密情報、CSV 生データの露出 |
| High | 秘密情報のコミット | API キー、トークン、Deploy Hook URL などのハードコード |
| Medium | CSV / パス操作 | パストラバーサル、不安全なパス結合、CSV パースの不正入力放置 |
| Medium | 新規 HTTP エンドポイント | 入力検証欠落、過剰な情報を返すレスポンス |
| Medium | コンテナ権限 | Dockerfile で root 実行のまま、不要な権限の追加 |
| Low | サプライチェーン設計コメント | 依存追加の意図・更新方針（CVE の再列挙はしない） |

## 見なくてよいもの（重複回避）

- gosec / Trivy / Grype が既にブロックする既知 CVE の再列挙
- スタイル・命名・パフォーマンス微修正
- テストのカバレッジ数値そのもの

## 実験コードと本番相当パス

| 区分 | パス例 | 扱い |
| --- | --- | --- |
| 本番相当 | `internal/handler/`、`internal/usecase/`、`internal/domain/`、`internal/infrastructure/`、`Dockerfile`、`.github/workflows/` | 厳しく見る。回帰は High 以上で報告 |
| 実験 | `internal/experiment/`、意図的脆弱サンプル、実験用ブランチの明示的ドリル | 検知してよい。PR 本文やコメントに実験意図が明示されていれば Critical 扱いにしない |

## 出力形式

指摘は次の列で簡潔にまとめる。

| Severity | Location | Finding |
| --- | --- | --- |
| High / Medium / Low | `path:line` | なぜ危険かと推奨修正を一文程度 |

- 指摘が無い場合はその旨を書く。
- ユーザーが明示するまでコード修正はしない。
- マージ可否の最終判断は人間と `Backend Verify` に任せる。
