# UT 計画: message コンテキスト

`.cursor/rules/ut-plan-go-api.mdc` に沿った計画。  
対象は PoC の Message API（`GET /` → usecase → CSV リポジトリ）の一般 UT。  
`internal/experiment/` や意図的脆弱エンドポイントは対象外。

関連: [go-architecture.md](../../go-architecture.md) / [docs/ut/README.md](../README.md)

---

## 対象ユニット一覧

| 層 | ユニット | パス |
| --- | --- | --- |
| domain | `NewMessage` | `backend/internal/domain/message` |
| usecase | `UseCase.GetByID` | `backend/internal/usecase/message` |
| handler | `MessageHandler.GetRoot` | `backend/internal/handler/http` |
| infra | `messageCSVRepository.FindByID` | `backend/internal/infrastructure/persistence` |

---

## 1. domain — `NewMessage`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| D-P1 | 空でない ID・本文なら Message を生成する |
| D-P2 | 空の ID または空の本文は拒否する |
| D-P3 | 特殊文字を含む本文もデータとして保持する（解釈・サニタイズしない） |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| D-01 | D-P1 | 妥当な ID・本文で Message を生成できること | id=`1`, content=`Hello World` | （なし） | err=nil, ID=`1`, Content=`Hello World` |
| D-02 | D-P2 | 空の ID を拒否すること | id=``, content=`Hello` | （なし） | err≠nil, msg=nil |
| D-03 | D-P2 | 空の本文を拒否すること | id=`1`, content=`` | （なし） | err≠nil, msg=nil |
| D-04 | D-P3 | SQL/XSS 風の本文をデータとして保持すること | id=`1`, content=`<script>…` 等 | （なし） | err=nil, Content が入力と同一 |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| D-P1 | D-01 |
| D-P2 | D-02, D-03 |
| D-P3 | D-04 |

---

## 2. usecase — `GetByID`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| U-P1 | 成功時はリポジトリのメッセージをそのまま返す |
| U-P2 | リポジトリエラーを握りつぶさず呼び出し元へ伝播する |
| U-P3 | ID を解釈・改変せず不透明値としてリポジトリへ渡す |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| U-01 | U-P1 | 取得成功時はメッセージをそのまま返すこと | id=`1` | msg content=`Hello World`, err=nil | err=nil, content=`Hello World` |
| U-02 | U-P2 | リポジトリエラーを握りつぶさず伝播すること | id=`missing` | err=`not found` | err≠nil, msg=nil |
| U-03 | U-P3 | 悪意のある ID を不透明値として転送すること | id=`1'; DROP TABLE messages;--` | err 任意 | repo が受け取った id が入力と同一 |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| U-P1 | U-01 |
| U-P2 | U-02 |
| U-P3 | U-03 |

---

## 3. handler — `GetRoot`

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| H-P1 | GET 成功時は 200 と JSON の `message` を返す |
| H-P2 | GET 以外は 405 で拒否する |
| H-P3 | usecase 失敗時は 500 と汎用メッセージにし、内部詳細を本文へ出さない |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| H-01 | H-P1 | 正常系で JSON メッセージと Content-Type が返ること | Method=GET, path=`/` | usecase が content=`Hello World` | status=200, Content-Type に `application/json`, body.message=`Hello World` |
| H-02 | H-P2 | GET 以外のメソッドを 405 で拒否すること | Method=POST, path=`/` | （呼ばない） | status=405 |
| H-03 | H-P3 | 内部エラー詳細がレスポンスへ漏洩しないこと | Method=GET | usecase err に path/secret/stack 相当の文言 | status=500, body は汎用（例: `internal server error`）, 内部断片を含まない |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| H-P1 | H-01 |
| H-P2 | H-02 |
| H-P3 | H-03 |

---

## 4. infra — `FindByID`（CSV）

### 責務リスト

| 責務ID | 責務 |
| --- | --- |
| I-P1 | CSV 上で ID が一致する行のメッセージを返す |
| I-P2 | 一致する行が無ければ not found エラーにする |
| I-P3 | 一致しない特殊 ID でもルックアップを安全に終える（解釈しない） |
| I-P4 | キャンセル済み context では処理を始めず context エラーを返す |
| I-P5 | CSV ファイルを開けなければエラーを返す |

### ケース表

| ケースID | 責務ID | 目的 | 入力 | 依存の戻り | 期待 |
| --- | --- | --- | --- | --- | --- |
| I-01 | I-P1 | CSV から ID 一致のメッセージを取得できること | id=`1`, CSV=`id,message\n1,Hello World\n` | （実ファイル） | err=nil, content=`Hello World` |
| I-02 | I-P2 | 存在しない ID は未発見エラーになること | id=`999`, 同上 CSV | （実ファイル） | err≠nil, msg=nil |
| I-03 | I-P3 | 特殊な ID を解釈せず一致しなければ not found で終えること | id=`1 OR 1=1` 等（CSV に無し） | （実ファイル） | err≠nil（not found）, msg=nil |
| I-04 | I-P4 | キャンセル済み context を検出して中断すること | ctx canceled, id=`1` | （実ファイル可） | err≠nil（context） |
| I-05 | I-P5 | CSV ファイル不在時にエラーを返すこと | 存在しない path | — | err≠nil |

### 網羅チェック

| 責務ID | 行 |
| --- | --- |
| I-P1 | I-01 |
| I-P2 | I-02 |
| I-P3 | I-03 |
| I-P4 | I-04 |
| I-P5 | I-05 |
