# UT 計画: 会員フロント（member-web）

| 項目 | 値 |
| --- | --- |
| 文書ID | `UT-MEM-3` |
| 機能 | `MEM`（会員フロント） |
| 版 | `0.1` |
| 状態 | `approved` |
| 関連文書 | EDS: [EDS-MEM-3](../../eds/EDS-MEM-3.md) / API 正本: [EDS-MEM-1](../../eds/EDS-MEM-1.md) / IT: （初期は未作成） |

`.cursor/rules/ut-plan-go-api.mdc` の計画形式に沿う（層名はフロントの `lib` ユニットに読み替える）。  
本計画は `approved`。実装・`*_test.ts` はこの表に従う。  
画面 E2E・React コンポーネントの描画テストは対象外。backend 契約の再検証は `UT-MEM-1` / `IT-MEM-1` に任せ、本計画に焼かない。

関連: [docs/ut/README.md](../README.md) / [docs/README.md](../../README.md)

---

## 対象ユニット一覧

| 層（相当） | ユニット | パス（想定） |
| --- | --- | --- |
| lib | 入力制約（loginId / password / nickname） | `frontend/lib/member/validation.ts` |
| lib | チャレンジ応答計算（Argon2id + HMAC） | `frontend/lib/member/auth.ts` |
| lib | 会員 API クライアント | `frontend/lib/member/api.ts` |
| lib | クライアントセッション（sessionStorage） | `frontend/lib/member/session.ts` |
| lib | 画面向け汎用エラー文言 | `frontend/lib/member/errors.ts` |

実装ファイル名は上記想定。分割・統合してもよいが、本表の責務は残す。

---

## 1. lib — 入力制約 `validation`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| V-P1 | 制約を満たす loginId・password・nickname を受理する |
| V-P2 | ログインIDが半角英数字 8〜15 でないときは拒否する |
| V-P3 | パスワードが長さ・英字・数字・許可記号の要件を満たさないときは拒否する |
| V-P4 | ニックネームが空（前後空白のみ含む）または 51 ルーン以上のときは拒否する |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| V-01 | V-P1 | 妥当な値を受理すること | loginId=`member01`, password=`Abcd123!`, nickname=`しろかぎ` | （なし） | ok=true |
| V-02 | V-P2 | 7 文字のログインIDを拒否すること | loginId=`member0`, 他は妥当 | （なし） | ok=false |
| V-03 | V-P3 | 記号を含まないパスワードを拒否すること | password=`Abcd1234`, 他は妥当 | （なし） | ok=false |
| V-04 | V-P4 | 51 ルーンのニックネームを拒否すること | nickname=（51ルーン）, 他は妥当 | （なし） | ok=false |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| V-P1 | V-01 |
| V-P2 | V-02 |
| V-P3 | V-03 |
| V-P4 | V-04 |

制約の正本は EDS-MEM-1 §6。本ユニットはクライアント側の事前チェックであり、最終判定は API。

---

## 2. lib — チャレンジ応答計算 `auth`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| A-P1 | 同一の password・salt・challenge から EDS-MEM-1 付録 A 準拠の固定応答（Base64）を返す |
| A-P2 | salt または challenge が不正な Base64 のときはエラーを返す（応答を捏造しない） |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| A-01 | A-P1 | 固定ベクトルで応答が期待 Base64 と一致すること | password・salt・challenge＝計画承認時に固定した既知値（付録 A パラメータ） | （なし） | response が期待 Base64 と同一 |
| A-02 | A-P2 | 不正 Base64 の salt では応答を返さないこと | salt=`@@@`, challenge=妥当 Base64, password=妥当 | （なし） | err≠null, response なし |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| A-P1 | A-01 |
| A-P2 | A-02 |

固定ベクトルは実装時に backend と同じパラメータで一度生成し、期待値を仕様としてテストに埋め込む（実行結果のコピーのみに依存しない）。Argon2id を `m=19456 KiB, t=2, p=1, 出力32B` 未満に弱めない。

---

## 3. lib — 会員 API クライアント `api`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| C-P1 | 登録成功（201）時に loginId・nickname を返す（password を応答に含めない） |
| C-P2 | 登録の非成功ステータスを成功扱いにせずエラーにする |
| C-P3 | 挑戦発行成功（200）時に challengeId・salt・challenge を返す |
| C-P4 | 応答検証成功（200）時に sessionToken・nickname・expiresAt を返す |
| C-P5 | 応答検証の 401 を成功扱いにせずエラーにする |
| C-P6 | 退会成功（204）をエラーにしない |
| C-P7 | 退会リクエストに `Authorization: Bearer <token>` を付与する |
| C-P8 | ログイン用 verify リクエストの JSON に password を含めない |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| C-01 | C-P1 | 登録 201 で公開フィールドのみ返すこと | loginId/password/nickname 妥当 | fetch 201 + `{loginId,nickname}` | err=null, 返却に password なし |
| C-02 | C-P2 | 登録 409 をエラーにすること | 同上 | fetch 409 | err≠null |
| C-03 | C-P3 | 挑戦 200 の JSON をそのまま返すこと | loginId=`member01` | fetch 200 + challenge フィールド | err=null, 3 フィールド取得 |
| C-04 | C-P4 | verify 200 でセッション情報を返すこと | challengeId/response | fetch 200 + session フィールド | err=null, 3 フィールド取得 |
| C-05 | C-P5 | verify 401 をエラーにすること | 同上 | fetch 401 | err≠null |
| C-06 | C-P6 | 退会 204 を成功とすること | token=`tok` | fetch 204 空 | err=null |
| C-07 | C-P7 | 退会で Bearer を送ること | token=`tok` | fetch 記録 | Authorization=`Bearer tok` |
| C-08 | C-P8 | verify ボディに password が無いこと | challengeId/response | fetch 記録 | JSON キーに password なし |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| C-P1 | C-01 |
| C-P2 | C-02 |
| C-P3 | C-03 |
| C-P4 | C-04 |
| C-P5 | C-05 |
| C-P6 | C-06 |
| C-P7 | C-07 |
| C-P8 | C-08 |

依存は `fetch` の狭い stub。実サーバや backend ハンドラは起動しない。

---

## 4. lib — クライアントセッション `session`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| S-P1 | 保存した sessionToken・nickname・expiresAt を読み出せる |
| S-P2 | 破棄後は未認証として扱う |
| S-P3 | expiresAt が過去のときは未認証として扱う（読み出し時に有効とみなさない） |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| S-01 | S-P1 | 保存直後に同一値を取得できること | token=`abc`, nickname=`しろかぎ`, expiresAt=未来 | sessionStorage stub | get が同一値 |
| S-02 | S-P2 | clear 後にセッションが無いこと | 事前に保存あり | stub | get=null（または未認証） |
| S-03 | S-P3 | 期限切れを未認証にすること | expiresAt=過去 | stub | get=null（または未認証） |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| S-P1 | S-01 |
| S-P2 | S-02 |
| S-P3 | S-03 |

`localStorage` や Cookie へ書かないこと。キー名は実装で固定し、テストはその実装キーを使う。

---

## 5. lib — 汎用エラー文言 `errors`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| E-P1 | 登録失敗を固定文言 `登録できませんでした。` に写す（ステータスを区別しない） |
| E-P2 | ログイン失敗を固定文言 `ログインできませんでした。` に写す |
| E-P3 | 退会失敗を固定文言 `退会できませんでした。` に写す |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| E-01 | E-P1 | 登録の 400 と 409 が同じ文言になること | status=400 と status=409 | （なし） | 両方とも `登録できませんでした。` |
| E-02 | E-P2 | ログイン失敗を固定文言にすること | status=401（または計算失敗相当） | （なし） | `ログインできませんでした。` |
| E-03 | E-P3 | 退会失敗を固定文言にすること | status=404 | （なし） | `退会できませんでした。` |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| E-P1 | E-01 |
| E-P2 | E-02 |
| E-P3 | E-03 |

HTTP 本文・内部例外文字列は返さない。

---

## Out of Scope（本計画）

- `/members/*` ページの React Testing Library / E2E
- CORS・ネットワーク実機・backend 起動
- Argon2 パラメータ弱化の許容
- `UT-MEM-1` と同一のサーバ側ドメイン／ユースケース再テスト
