# ドキュメント索引

設計・テスト計画の置き場と番号規約の入口。アーキテクチャやデプロイ方針など横断ドキュメントもここから辿る。

## 種別と置き場

| 種別 | 置き場 | ひな形 | ルール |
| --- | --- | --- | --- |
| 外部設計（EDS） | [eds/](./eds/) | [eds/_TEMPLATE.md](./eds/_TEMPLATE.md) | [.cursor/rules/docs-numbering.mdc](../.cursor/rules/docs-numbering.mdc) |
| ユニットテスト（UT） | [ut/](./ut/) | [ut/plans/_TEMPLATE.md](./ut/plans/_TEMPLATE.md) | [.cursor/rules/ut-plan-go-api.mdc](../.cursor/rules/ut-plan-go-api.mdc) |
| 結合テスト（IT） | [it/](./it/) | [it/plans/_TEMPLATE.md](./it/plans/_TEMPLATE.md) | [.cursor/rules/it-plan.mdc](../.cursor/rules/it-plan.mdc) |

## ナンバリング

形式: `<種別>-<機能コード>-N`（`N` は `1` から始まる連番。ゼロ埋めしない）

| 種別 | プレフィックス | ファイル例 |
| --- | --- | --- |
| 外部設計 | `EDS-<機能>-N` | `docs/eds/EDS-MEM-1.md` |
| UT 計画 | `UT-<機能>-N` | `docs/ut/plans/UT-MEM-1.md` |
| UT 実行記録 | 計画番号＋日付 | `docs/ut/runs/YYYY-MM-DD-UT-MEM-1.md` |
| 結合計画 | `IT-<機能>-N` | `docs/it/plans/IT-MEM-1.md` |
| 結合実行記録 | 計画番号＋日付 | `docs/it/runs/YYYY-MM-DD-IT-MEM-1.md` |

### 機能コード

| コード | 機能 |
| --- | --- |
| `MEM` | 会員（member） |
| `MSG` | message（PoC） |

新規機能を追加するときは、先に上表へコードを追記してから採番する。

### 採番ルール

- **ファイル名＝文書ID**（拡張子 `.md` を除く）。
- 同一種別・同一機能内で `1` から連番。欠番は埋めない。ゼロ埋め（`001` など）は使わない。
- 外部設計は **機能ごと原則1冊**。会員（`MEM`）は登録・認証・削除を `EDS-MEM-1` にまとめる。
- 改訂は **同一文書IDのまま版（版フィールド）を上げる**。内容が別機能になった場合のみ新番号。
- 文書ヘッダに必ず書く: `文書ID` / `機能` / `版` / `状態`（`draft` / `approved`） / `関連文書`。
- ブランチ番号（`feature/1-member`）と文書番号は独立してよい。

### ブランチ番号の予約

ブランチの `<番号>`（`feature/<番号>-…`）のうち、次は **取り消し専用** とし、通常機能のトピックに使わない。詳細は [.cursor/rules/branch-workflow.mdc](../.cursor/rules/branch-workflow.mdc)。

| 番号 | 用途 | 例 |
| --- | --- | --- |
| `1` | 会員（member）トピック | `feature/1-member` |
| `2` | **取り消し（revert）専用** | `feature/2-revert-main-eds` |
| `3` 以降 | 次の通常機能から採番 | |

## 新しい文書の作り方

1. 対応する `_TEMPLATE.md` をコピーする。
2. 未使用の次番号でリネームする（例: `EDS-MEM-1.md`）。
3. ヘッダを埋め、関連する EDS / UT / IT を相互リンクする。
4. 計画類は人が `approved` するまで実装・テストコードを書かない（UT / IT ルールに従う）。

## その他のドキュメント

| 文書 | 内容 |
| --- | --- |
| [go-architecture.md](./go-architecture.md) | Go API の層構成（DDD） |
| [frontend-architecture.md](./frontend-architecture.md) | Frontend 構成 |
| [render.md](./render.md) | デプロイ方針 |
| [security-test-plan.md](./security-test-plan.md) | セキュリティテスト計画 |
| [security-agent-threat-model.md](./security-agent-threat-model.md) | ローカルセキュリティ Agent 脅威モデル |
