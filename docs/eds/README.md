# 外部設計書（EDS）

機能の外部から見える契約（API・データ制約・認証フロー・状態・エラー方針）を書く。  
内部のパッケージ構成や SQL は書かない（必要なら [go-architecture.md](../go-architecture.md) へリンク）。

番号規約の正本: [docs/README.md](../README.md)  
ルール: [.cursor/rules/docs-numbering.mdc](../../.cursor/rules/docs-numbering.mdc)

## 構成

```text
docs/eds/
  README.md
  _TEMPLATE.md          # コピー元
  EDS-<機能>-NNN.md     # 実文書（ファイル名＝文書ID）
```

## 運用

1. `_TEMPLATE.md` をコピーし、`EDS-<機能>-NNN.md` にリネームする。
2. 機能ごと **原則1冊**（会員 `MEM` は登録・認証・削除を1冊にまとめる）。
3. 状態が `approved` になるまで実装の正本にしない。
4. 改訂は同一 ID のまま `版` を上げる。

## 関連テスト

外部設計のトレーサビリティ欄から `UT-<機能>-NNN` / `IT-<機能>-NNN` へリンクする。
