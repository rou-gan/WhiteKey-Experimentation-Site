# ユニットテスト計画・実行記録

一般 UT（セキュリティ専用計画ではない）の **計画** と **実行記録** を置く場所。

ルール: `.cursor/rules/ut-plan-go-api.mdc`  
番号規約の正本: [docs/README.md](../README.md)

## 構成

```text
docs/ut/
  README.md
  plans/
    _TEMPLATE.md              # 計画のひな形（コピーして使う）
    UT-<機能>-NNN.md          # 計画（ファイル名＝文書ID）
  runs/
    _TEMPLATE.md              # 実行記録のひな形
    YYYY-MM-DD-UT-<機能>-NNN.md
```

| 置き場 | 残すもの | 運用 |
| --- | --- | --- |
| `plans/` | 責務リスト＋ケース表 | `UT-<機能>-NNN.md`。ひな形は `_TEMPLATE.md` |
| `runs/` | 実行記録（ケース表＋結果） | `YYYY-MM-DD-UT-<機能>-NNN.md` |

実行記録の結果表は、計画の列（目的・入力・依存の戻り・期待）を残し、末尾に **結果** を付ける。差分メモや解説は書かない。

レガシー: `plans/message.md` など文書ID形式以前のファイルはそのまま残してよい。新規・大きな改訂時に `UT-MSG-001` 形式へ揃える。

## 流れ

1. `plans/_TEMPLATE.md` をコピーし、`UT-<機能>-NNN.md` にリネームする（コード禁止）。
2. 人が表を承認する（状態を `approved` にする）。
3. 承認後に `_test.go` を書く／更新する。
4. テスト実行後、`runs/_TEMPLATE.md` をコピーして結果を残す。

## 実行記録に書かないもの

- `go test -v` の全文ログ
- `coverage.out` / HTML（リポジトリでは ignore）

CI の合否はゲート、ここは計画どおり回した記録。
