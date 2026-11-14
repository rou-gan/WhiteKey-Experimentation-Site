# 結合テスト計画・実行記録

複数 API／層をまたぐシナリオの **計画** と **実行記録** を置く場所。  
単一ユニットの責務検証は [docs/ut/](../ut/) に書き、ここへ重複させない。

ルール: `.cursor/rules/it-plan.mdc`  
番号規約の正本: [docs/README.md](../README.md)

## 構成

```text
docs/it/
  README.md
  plans/
    _TEMPLATE.md              # 計画のひな形
    IT-<機能>-NNN.md          # 計画（ファイル名＝文書ID）
  runs/
    _TEMPLATE.md              # 実行記録のひな形
    YYYY-MM-DD-IT-<機能>-NNN.md
```

| 置き場 | 残すもの | 運用 |
| --- | --- | --- |
| `plans/` | シナリオ表 | `IT-<機能>-NNN.md` |
| `runs/` | シナリオ＋結果 | `YYYY-MM-DD-IT-<機能>-NNN.md` |

実行記録は計画の列を残し、末尾に **結果** を付ける。フルログは残さない。

## 流れ

1. `plans/_TEMPLATE.md` をコピーし、`IT-<機能>-NNN.md` にリネームする。
2. 人がシナリオを承認する（状態を `approved` にする）。
3. 承認後に結合テストを実装する。
4. 実行後、`runs/_TEMPLATE.md` をコピーして結果を残す。
