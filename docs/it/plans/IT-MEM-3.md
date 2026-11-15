# 結合テスト計画: 会員フロント（ブラウザ E2E）

| 項目 | 値 |
| --- | --- |
| 文書ID | `IT-MEM-3` |
| 機能 | `MEM`（会員フロント） |
| 版 | `0.1` |
| 状態 | `approved` |
| 関連文書 | EDS: [EDS-MEM-3](../../eds/EDS-MEM-3.md) / API: [EDS-MEM-1](../../eds/EDS-MEM-1.md) / UT: [UT-MEM-3](../../ut/plans/UT-MEM-3.md) / API IT: [IT-MEM-1](./IT-MEM-1.md) |

`.cursor/rules/it-plan.mdc` に沿った計画。  
本計画は `approved`。Playwright E2E（`frontend/e2e`）はこのシナリオ表に従う。  
関連: [docs/it/README.md](../README.md) / [docs/README.md](../../README.md)

---

## 範囲

### In Scope

- ブラウザから登録 → ログイン → 会員ホーム表示の一連
- ログアウト後に会員ホームへ戻れないこと
- 退会後に同じ資格情報でログインできないこと

### Out of Scope

- 入力制約・KDF/HMAC・API ヘルパ単体 → `UT-MEM-3`
- HTTP のみの API 一連・DB マスク詳細 → `IT-MEM-1`
- ビジュアルリグレッション・複数ブラウザ網羅（初期は Chromium のみ）
- CI 必須化（初期はローカル実行。必要なら後続で Frontend Verify に追加）

---

## 前提・環境

| 項目 | 内容 |
| --- | --- |
| ツール | Playwright（`frontend/`） |
| 起動方法 | ローカル: backend `docker compose up`（`:8080`）＋ frontend `npm run dev`（`:3000`、`NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`） |
| データ初期状態 | 既存会員と衝突しない一意の loginId / nickname をシナリオごとに生成 |
| 観測手段 | 画面遷移・表示テキスト・固定エラー文言。必要時のみ DB は人が確認（自動化の主観測は UI） |

---

## シナリオ表

| シナリオID | 目的（「〜すること」） | 前提データ | 手順（要約） | 期待 | EDS | UT |
| --- | --- | --- | --- | --- | --- | --- |
| IT-MEM-3-S01 | 登録からログインしてニックネームが見えること | 会員なし（一意 ID） | 1. `/members/register` で登録 2. `/members/login` でログイン | ログイン後 `/members/me` に nickname 表示 | EDS-MEM-3 F-01〜F-03 | V-* / A-* / C-* |
| IT-MEM-3-S02 | ログアウト後に会員ホームを使えないこと | S01 相当でログイン済み | 1. ログアウト 2. `/members/me` を開く | `/members/login` へ誘導（未認証） | F-04 | S-02 |
| IT-MEM-3-S03 | 退会後に同じ資格情報でログインできないこと | S01 相当でログイン済み | 1. 退会（確認 OK）2. 旧 loginId/password でログイン | `/members/login` に留まり `ログインできませんでした。` | F-05 / EDS-MEM-1 F-04 | C-06 / E-02 |

### シナリオ詳細

#### S01 — 登録からホーム表示

1. `/members/register` で loginId / password / nickname を送信
2. `/members/login` へ遷移後、同じ資格情報でログイン

**期待**

- 登録成功後にログイン画面
- ログイン成功後に `/members/me` で nickname がテキスト表示される

#### S02 — ログアウト

1. S01 後に「ログアウト」
2. `/members/me` を直接開く

**期待**

- URL が `/members/login`（またはログイン画面の見出し）

#### S03 — 退会後のログイン失敗

1. S01 後に「退会」→ 確認ダイアログで OK
2. 同じ loginId / password でログインを試行

**期待**

- 画面文言 `ログインできませんでした。`
- `/members/me` に到達しない

---

## 網羅チェック（EDS-MEM-3 機能）

| EDS 機能ID | シナリオ |
| --- | --- |
| F-01 登録画面 | S01 |
| F-02 ログイン画面 | S01, S03 |
| F-03 会員ホーム | S01 |
| F-04 ログアウト | S02 |
| F-05 退会 | S03 |
