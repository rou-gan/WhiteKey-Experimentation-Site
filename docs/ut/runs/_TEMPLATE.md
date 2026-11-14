# UT 実行記録: \<文書ID\>

| 項目 | 値 |
| --- | --- |
| 計画文書ID | `UT-<機能>-NNN` |
| 日付 | YYYY-MM-DD |
| ブランチ |  |
| commit |  |
| 計画 | `docs/ut/plans/UT-<機能>-NNN.md` |
| コマンド | 下記 |

```text
cd backend
go test ./internal/... -count=1
```

## 結果

計画のケース表をコピーし、末尾に **結果** 列を付ける。FAIL 時のみ備考を短く。

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 | 結果 |
| --- | --- | --- | --- | --- | --- | --- |
|  |  |  |  |  |  | PASS / FAIL |

## まとめ

- 全体: PASS / FAIL
- メモ: （FAIL 時のみ。不要なら空）
