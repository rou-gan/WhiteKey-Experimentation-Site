# UT 計画: 会員（member）

| 項目 | 値 |
| --- | --- |
| 文書ID | `UT-MEM-1` |
| 機能 | `MEM`（会員） |
| 版 | `0.1` |
| 状態 | `approved` |
| 関連文書 | EDS: [EDS-MEM-1](../../eds/EDS-MEM-1.md)（v0.3・許可記号一覧あり） / OpenAPI: `backend/openapi/openapi.yaml` / IT: `docs/it/plans/IT-MEM-1`（未作成） |

`.cursor/rules/ut-plan-go-api.mdc` に沿った計画。  
実装パッケージ名は [go-architecture.md](../../go-architecture.md) の `member` コンテキスト想定。本計画は `approved`。実装・`_test.go` はこの表に従う。

関連: [docs/ut/README.md](../README.md) / [docs/README.md](../../README.md)

---

## 対象ユニット一覧

| 層 | ユニット | パス（想定） |
| --- | --- | --- |
| domain | `NewActiveMember` | `backend/internal/domain/member` |
| domain | `MaskForLogicalDelete` | `backend/internal/domain/member` |
| usecase | `UseCase.Register` | `backend/internal/usecase/member` |
| usecase | `UseCase.IssueChallenge` | `backend/internal/usecase/member` |
| usecase | `UseCase.VerifyResponse` | `backend/internal/usecase/member` |
| usecase | `UseCase.DeleteMe` | `backend/internal/usecase/member` |
| handler | `MemberHandler`（登録／挑戦／検証／削除） | `backend/internal/handler/http` |
| infra | `memberPostgresRepository`（保存・参照・論理削除） | `backend/internal/infrastructure/persistence` |

---

## 1. domain — `NewActiveMember`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| D-N1 | 制約を満たす loginId・password・nickname なら active な会員を生成する |
| D-N2 | ログインIDが半角英数字 8〜15 でないときは拒否する |
| D-N3 | パスワードが長さ・英字・数字・記号の要件を満たさないときは拒否する |
| D-N4 | ニックネームが空または 51 ルーン以上のときは拒否する |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| D-N-01 | D-N1 | 妥当な値で active 会員を生成できること | loginId=`member01`, password=`Abcd123!`, nickname=`しろかぎ` | （なし） | err=nil, status=active, 各フィールドが入力どおり |
| D-N-02 | D-N2 | 7 文字のログインIDを拒否すること | loginId=`member0`（7文字）, 他は妥当 | （なし） | err≠nil, member=nil |
| D-N-03 | D-N3 | 記号を含まないパスワードを拒否すること | password=`Abcd1234`, 他は妥当 | （なし） | err≠nil, member=nil |
| D-N-04 | D-N4 | 51 ルーンのニックネームを拒否すること | nickname=（51ルーン）, 他は妥当 | （なし） | err≠nil, member=nil |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| D-N1 | D-N-01 |
| D-N2 | D-N-02 |
| D-N3 | D-N-03 |
| D-N4 | D-N-04 |

---

## 2. domain — `MaskForLogicalDelete`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| D-M1 | 論理削除時に loginId・nickname・秘密値相当を復元不能なマスクに置き換える |
| D-M2 | 状態を deleted にし、active のまま残さない |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| D-M-01 | D-M1 | マスク後に元の loginId・nickname・秘密値が残らないこと | active 会員（loginId=`member01` 等） | （なし） | マスク後の loginId/nickname/秘密値 ≠ 元値 |
| D-M-02 | D-M2 | マスク後の状態が deleted であること | 同上 | （なし） | status=deleted |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| D-M1 | D-M-01 |
| D-M2 | D-M-02 |

---

## 3. usecase — `Register`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| U-R1 | 成功時はリポジトリへ保存し loginId・nickname を返す（password は返さない） |
| U-R2 | ドメイン検証失敗を保存せず呼び出し元へ返す |
| U-R3 | リポジトリの重複エラーを衝突として伝播する |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| U-R-01 | U-R1 | 登録成功時に公開フィールドのみ返すこと | 妥当な loginId/password/nickname | save err=nil | err=nil, loginId・nickname 返却, 応答に password なし, save が1回呼ばれる |
| U-R-02 | U-R2 | 無効パスワードではリポジトリを呼ばないこと | password=`short`, 他は妥当 | （呼ばない） | err≠nil, save 呼び出し回数=0 |
| U-R-03 | U-R3 | 重複を衝突エラーとして伝播すること | 妥当な入力 | save err=`duplicate` | err=衝突相当, 応答ボディ用の成功値なし |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| U-R1 | U-R-01 |
| U-R2 | U-R-02 |
| U-R3 | U-R-03 |

---

## 4. usecase — `IssueChallenge`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| U-C1 | 形式妥当な loginId なら challengeId・salt・challenge を返す |
| U-C2 | 未登録でも登録済みと同じ形の成功結果を返す（ダミー可） |
| U-C3 | 形式不正な loginId は挑戦を発行せずエラーにする |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| U-C-01 | U-C1 | 登録済み会員へ挑戦を発行できること | loginId=`member01` | find=active 会員 | err=nil, challengeId/salt/challenge が非空 |
| U-C-02 | U-C2 | 未登録でも成功形で返すこと | loginId=`unknown01` | find=not found | err=nil, 各フィールド非空（ダミー可） |
| U-C-03 | U-C3 | 短すぎる loginId を拒否すること | loginId=`short` | （find 呼ばないまたは呼ばない想定） | err≠nil |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| U-C1 | U-C-01 |
| U-C2 | U-C-02 |
| U-C3 | U-C-03 |

---

## 5. usecase — `VerifyResponse`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| U-V1 | 正当な応答なら sessionToken・nickname・expiresAt を返す |
| U-V2 | 応答不一致・挑戦不明などを認証失敗として返す（理由を区別しない） |
| U-V3 | 成功後は当該挑戦を再利用不可にする |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| U-V-01 | U-V1 | 正しい応答でセッションを発行すること | 有効 challengeId + 正当 response | 挑戦有効・秘密一致 | err=nil, sessionToken 非空, nickname 非空, expiresAt 設定 |
| U-V-02 | U-V2 | 誤った応答を認証失敗にすること | 有効 challengeId + 不正 response | 挑戦有効 | err=認証失敗相当（不一致用の別型にしない） |
| U-V-03 | U-V3 | 一度成功した挑戦の再検証を失敗させること | 成功済み challengeId + 正当 response | 挑戦消費済み | err=認証失敗相当 |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| U-V1 | U-V-01 |
| U-V2 | U-V-02 |
| U-V3 | U-V-03 |

---

## 6. usecase — `DeleteMe`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| U-D1 | 認証済み会員を論理削除（マスク）しセッションを無効化する |
| U-D2 | 対象が既に削除済みなどのときは失敗を伝播する |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| U-D-01 | U-D1 | 削除成功時にマスク保存とセッション無効化が行われること | 有効 session → member | find=active, save/mask err=nil | err=nil, mask 保存が呼ばれる, 当該 session 無効 |
| U-D-02 | U-D2 | 削除済み会員の削除を失敗させること | 有効 session → 既に deleted | find=deleted または not found | err≠nil |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| U-D1 | U-D-01 |
| U-D2 | U-D-02 |

---

## 7. handler — `MemberHandler`

HTTP 契約の写像のみ。ドメイン／ユースケースの詳細ルールは証明しない。

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| H-P1 | 登録成功を 201 と JSON（loginId, nickname）で返す |
| H-P2 | 登録の検証失敗／衝突を 400／409 の汎用 JSON にし、項目列挙しない |
| H-P3 | 挑戦発行成功を 200 で返す（usecase 成功をそのまま写像） |
| H-P4 | 検証失敗を 401 の汎用 JSON にする |
| H-P5 | 削除成功を 204、未認証を 401、対象なしを 404 にする |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| H-01 | H-P1 | POST `/members` 成功が 201 になること | 妥当 JSON | usecase 成功 | status=201, body.loginId/nickname あり, password なし |
| H-02 | H-P2 | 登録衝突が 409 汎用になること | 妥当 JSON | usecase 衝突 | status=409, body に loginId/nickname のどちらが重複かの区別なし |
| H-03 | H-P3 | POST `/members/auth/challenge` 成功が 200 になること | loginId 付き JSON | usecase 成功（ダミー含む） | status=200, challengeId/salt/challenge あり |
| H-04 | H-P4 | 検証失敗が 401 汎用になること | challengeId+response | usecase 認証失敗 | status=401, 失敗理由の区別なし |
| H-05 | H-P5 | DELETE `/members/me` 成功が 204 になること | Authorization Bearer 有効 | usecase 成功 | status=204, ボディ空 |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| H-P1 | H-01 |
| H-P2 | H-02 |
| H-P3 | H-03 |
| H-P4 | H-04 |
| H-P5 | H-05 |

---

## 8. infra — `memberPostgresRepository`

`TEST_DATABASE_URL` があるときだけ実行（無いときは skip）。migrate 適用後に実行する。

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| I-P1 | active 会員を loginId で取得できる |
| I-P2 | 論理削除後は旧 loginId では active として取得できない |
| I-P3 | 同一 active の loginId 重複保存を拒否する |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| I-01 | I-P1 | 保存した会員を loginId で取得できること | 保存 loginId=`member01` | （実 DB） | err=nil, 取得行が active |
| I-02 | I-P2 | マスク削除後に旧 loginId で active 取得できないこと | 削除前 loginId=`member01` | （実 DB） | 削除後 Find active は not found |
| I-03 | I-P3 | 同じ loginId の二重保存が失敗すること | 同一 loginId で2回 Save | （実 DB） | 2回目 err≠nil |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| I-P1 | I-01 |
| I-P2 | I-02 |
| I-P3 | I-03 |
