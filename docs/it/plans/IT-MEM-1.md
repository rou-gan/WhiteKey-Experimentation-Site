# 結合テスト計画: 会員

| 項目 | 値 |
| --- | --- |
| 文書ID | `IT-MEM-1` |
| 機能 | `MEM`（会員） |
| 版 | `0.1` |
| 状態 | `draft` |
| 関連文書 | EDS: [EDS-MEM-1](../../eds/EDS-MEM-1.md) / OpenAPI: `backend/openapi/openapi.yaml` / UT: [UT-MEM-1](../../ut/plans/UT-MEM-1.md) |

`.cursor/rules/it-plan.mdc` に沿った計画。  
結合テストコードは本計画が `approved` になるまで作らない。

関連: [docs/it/README.md](../README.md) / [docs/README.md](../../README.md)

---

## 範囲

### In Scope

- 登録 → 挑戦発行 → 応答検証 → 削除の一連（HTTP＋永続状態）
- 削除後に同じ資格情報で認証できないこと
- 未登録 loginId でも挑戦は 200 だが検証は失敗すること
- 有効会員間の loginId／ニックネーム一意が永続込みで効くこと

### Out of Scope

- 単層のバリデーション網羅（長さ・文字種の細部）→ `UT-MEM-1`
- ハンドラ単体の status 写像のみ → `UT-MEM-1`
- レート制限・アカウントロック（EDS 初期未実装）
- フロントエンド UI

---

## 前提・環境

| 項目 | 内容 |
| --- | --- |
| 起動方法 | `backend` の Docker Compose（API + Postgres）＋ migrate |
| データ初期状態 | 会員テーブルは空（またはテスト用に truncate） |
| 観測手段 | HTTP status / JSON、DB の status・マスク後 loginId/nickname、セッション無効化後の API 結果 |
| クライアント計算 | OpenAPI / EDS §7 どおり。`secret = Argon2id(password, salt)`、`response = Base64(HMAC-SHA256(secret, challenge_bytes))` |

---

## シナリオ表

| シナリオID | 目的（「〜すること」） | 前提データ | 手順（要約） | 期待 | EDS | UT |
| --- | --- | --- | --- | --- | --- | --- |
| IT-MEM-1-S01 | 登録から認証まで通してセッションを得ること | 会員なし | 1. POST `/members` 2. POST challenge 3. POST verify | 201→200→200、sessionToken 非空 | F-01, F-02, F-03 | U-R-01, U-C-01, U-V-01 |
| IT-MEM-1-S02 | 削除後に旧パスワードで認証できないこと | S01 相当の会員＋セッション | 1. DELETE `/members/me` 2. challenge（旧 loginId）3. verify | 204→200→401、DB は deleted＋マスク | F-02, F-03, F-04 | U-D-01, I-02 |
| IT-MEM-1-S03 | 未登録 loginId でも挑戦は成功形だが検証は失敗すること | 会員なし | 1. challenge（未登録）2. 任意 response で verify | 200（ダミー）→401 | F-02, F-03 / §5.2–5.3 | U-C-02, U-V-02 |
| IT-MEM-1-S04 | 既に使われている loginId で再登録できないこと | 会員 A 登録済み | 1. 別 nickname で同じ loginId を POST `/members` | 409、汎用（どちらが重複か区別なし） | F-01 / §5.1 | U-R-03, I-03 |
| IT-MEM-1-S05 | 既に使われている nickname で再登録できないこと | 会員 A 登録済み | 1. 別 loginId で同じ nickname を POST `/members` | 409、汎用（区別なし） | F-01 / §5.1 | U-R-03 |

### シナリオ詳細

#### S01 — 登録からセッション発行

1. `POST /members`（例: loginId=`member01`, password=`Abcd123!`, nickname=`しろかぎ`）
2. `POST /members/auth/challenge`（loginId=`member01`）
3. salt・challenge から応答を計算し `POST /members/auth/verify`

**期待（観測可能）**

- HTTP: 201（loginId/nickname、password なし）→ 200（challengeId/salt/challenge）→ 200（sessionToken, nickname, expiresAt）
- 永続状態: 会員が active で存在する

#### S02 — 削除後の認証不可

1. S01 相当で得た Bearer で `DELETE /members/me`
2. 旧 loginId で challenge
3. 旧パスワードから計算した response で verify（新しい challenge の値を使う）

**期待（観測可能）**

- HTTP: 204 → 200 → 401
- 永続状態: status=deleted、loginId/nickname が元値と異なるマスク、旧 sessionToken での要認証 API は 401

#### S03 — 未登録でも挑戦は 200

1. 存在しない loginId（形式は妥当、例: `unknown01`）で challenge
2. 返却 salt/challenge と適当な password から response を作り verify

**期待（観測可能）**

- HTTP: 200（フィールド非空）→ 401
- 404 にならないこと

#### S04 / S05 — 一意制約（loginId / nickname）

1. 会員 A を登録済みにする
2. S04: 同一 loginId・別 nickname で登録 / S05: 別 loginId・同一 nickname で登録

**期待（観測可能）**

- HTTP: 409、本文は汎用（loginId か nickname かの区別なし）
- 永続状態: 2件目の active 行は増えない

---

## 網羅チェック（EDS 機能に対して）

| EDS 機能ID | シナリオ |
| --- | --- |
| F-01 | S01, S04, S05 |
| F-02 | S01, S02, S03 |
| F-03 | S01, S02, S03 |
| F-04 | S02 |
