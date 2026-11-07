# Go アプリケーションアーキテクチャ（DDD）

「しろかぎ」のセキュリティ実験場における Go API の構成方針をまとめる。  
ドメイン駆動設計（DDD）を取り入れつつ、**パッケージのネストを深くしすぎず、ビジネス機能（コンテキスト）ごとにグループ化**する。

`internal/` の中にレイヤー（`domain` / `usecase` / `infrastructure` / `handler`）を閉じ込め、利用側（`cmd/`）からシンプルに組み立てられる構成とする。

## 方針

- **コンテキスト単位:** 機能ごとに `user` / `message` などのパッケージを切る（レイヤー横断で同じ名前空間を揃える）。
- **依存の方向:** `handler` → `usecase` → `domain` ← `infrastructure`（infrastructure が domain のインターフェースを実装）。
- **domain は外部依存ゼロ:** DB・HTTP・CSV ライブラリなどに依存しない。
- **DI は起動時のみ:** 依存注入は `backend/cmd/server/main.go` に集約する。
- **モノレポ:** Go コードは `backend/` 配下に置き、将来の `frontend/` と同居できる構成にする。
- **公開範囲:** アプリ固有コードは原則 `internal/` 配下に置く。汎用ライブラリ化する予定がなければ `pkg/` は使わない。

## ディレクトリ構成

```text
WhiteKey-Experimentation-Site/
├── backend/                        # Go API（将来の frontend/ と並べる）
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # 依存注入（DI）とサーバー起動
│   ├── internal/
│   │   ├── domain/                 # ビジネスルール（外部依存ゼロ）
│   │   │   └── user/
│   │   │       ├── user.go         # エンティティ・値オブジェクト
│   │   │       └── repository.go   # リポジトリのインターフェース
│   │   ├── usecase/                # アプリケーションロジック
│   │   │   └── user/
│   │   │       └── usecase.go      # ユースケース実装
│   │   ├── infrastructure/         # 技術的詳細・外部依存
│   │   │   ├── persistence/
│   │   │   │   └── user_repo.go    # 永続化の実装（CSV / RDB 等）
│   │   │   └── config/
│   │   └── handler/                # HTTP 等のエントリポイント
│   │       └── http/
│   │           └── user_handler.go
│   ├── data/                       # 同梱データ（CSV 仮想 DB 等）
│   ├── openapi/                    # OpenAPI 仕様（openapi.yaml）
│   ├── Dockerfile
│   ├── docker-compose.yml          # ローカル起動（backend）
│   ├── go.mod
│   └── go.sum
├── frontend/                       # （将来）フロントエンド
└── docs/
```

PoC（`GET /` で `"Hello World"` を返す）では、上記の `user` コンテキストに相当するものとして `message` コンテキストを置く想定。レイヤー分割の型は同じである。

## レイヤー責務

| レイヤー | 配置 | 責務 |
| --- | --- | --- |
| Domain | `internal/domain/<context>/` | エンティティ・値オブジェクト・ドメインサービス・リポジトリ IF |
| Usecase | `internal/usecase/<context>/` | ユースケースの進行、入力 DTO、ドメイン／リポジトリの組み立て |
| Infrastructure | `internal/infrastructure/` | CSV・設定・外部 API など技術詳細。domain の IF を実装 |
| Handler | `internal/handler/http/` | HTTP 入出力、ステータスコード、JSON / テキスト変換 |
| Cmd | `cmd/server/` | DI とプロセス起動のみ。ビジネスロジックを書かない |

## 主要レイヤーのサンプルコード

以下は構成理解のための **参考実装**（`user` コンテキスト）。モジュールパスは実装時の `go.mod` に合わせる。

### 1. ドメイン層（`internal/domain/user/`）

ビジネスルールとエンティティ、リポジトリのインターフェースのみを定義する。

`user.go`（エンティティ）

```go
package user

import "errors"

type ID string

type User struct {
	id   ID
	name string
}

func NewUser(id ID, name string) (*User, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	return &User{id: id, name: name}, nil
}

func (u *User) ID() ID       { return u.id }
func (u *User) Name() string { return u.name }
```

`repository.go`（インターフェース）

```go
package user

import "context"

type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id ID) (*User, error)
}
```

### 2. ユースケース層（`internal/usecase/user/`）

ドメインオブジェクトとインターフェースを組み立て、アプリケーションの処理を進行させる。

`usecase.go`

```go
package user

import (
	"context"

	"my-project/internal/domain/user"
)

type UseCase struct {
	repo user.Repository
}

func NewUseCase(repo user.Repository) *UseCase {
	return &UseCase{repo: repo}
}

type RegisterInput struct {
	ID   string
	Name string
}

func (uc *UseCase) Register(ctx context.Context, in RegisterInput) error {
	u, err := user.NewUser(user.ID(in.ID), in.Name)
	if err != nil {
		return err
	}
	return uc.repo.Save(ctx, u)
}
```

### 3. インフラストラクチャ層（`internal/infrastructure/persistence/`）

ドメイン層で定義された `Repository` インターフェースを実装する。

`user_repo.go`（RDB 例）

```go
package persistence

import (
	"context"
	"database/sql"

	"my-project/internal/domain/user"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) Save(ctx context.Context, u *user.User) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (id, name) VALUES (?, ?)", u.ID(), u.Name())
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	var name string
	err := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", string(id)).Scan(&name)
	if err != nil {
		return nil, err
	}
	return user.NewUser(id, name)
}
```

本プロジェクトの永続化方針は CSV 仮想 DB である（詳細は [render.md](./render.md)）。実装時は上記と同型の IF 実装を、CSV 読み書きに置き換える。

```go
// 概念例: CSV 実装も同じ Repository を満たす
func NewUserCSVRepository(path string) user.Repository {
	return &userCSVRepository{path: path}
}
```

### 4. 起動処理（`cmd/server/main.go`）

すべての依存関係を組み立てる（Dependency Injection）場所とする。

`main.go`

```go
package main

import (
	"database/sql"
	"log"
	"net/http"

	userhttp "my-project/internal/handler/http"
	"my-project/internal/infrastructure/persistence"
	userUseCase "my-project/internal/usecase/user"
)

func main() {
	db, err := sql.Open("mysql", "user:pass@tcp(127.0.0.1:3306)/dbname")
	if err != nil {
		log.Fatal(err)
	}

	// 依存関係の注入
	userRepo := persistence.NewUserRepository(db)
	userUC := userUseCase.NewUseCase(userRepo)
	userHandler := userhttp.NewUserHandler(userUC)

	// ルーティング設定 & サーバー起動
	http.HandleFunc("/users", userHandler.Register)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

本プロジェクトの PoC では、MySQL ではなく CSV リポジトリを注入し、`GET /` 用ハンドラを登録する。

## 本プロジェクト PoC への当てはめ

| 項目 | PoC での想定 |
| --- | --- |
| コンテキスト | `message`（表示メッセージ） |
| エンドポイント | `GET /` |
| ユースケース例 | CSV からメッセージを取得して返す |
| 永続化 | `internal/infrastructure/persistence` の CSV 実装 |
| データ正本 | リポジトリ同梱 CSV（例: `backend/data/messages.csv`） |

依存の流れ（PoC）:

```text
backend/cmd/server
  → backend/internal/handler/http (GET /)
    → usecase/message
      → domain/message (Entity + Repository IF)
      ← infrastructure/persistence (CSV Repository 実装)
```

## 依存関係ルール（守ること）

1. `domain` は他レイヤーを import しない。
2. `usecase` は `domain` のみに依存する（infrastructure / handler を import しない）。
3. `handler` は `usecase`（必要なら DTO）に依存し、infrastructure を直接呼ばない。
4. `infrastructure` は `domain` のインターフェースを実装する。usecase を import しない。
5. 具象の配線は `cmd/server`（または薄い DI 用パッケージ）に閉じる。

## 今後の実装メモ

- [x] `backend/go.mod` のモジュールパス確定（`.../backend`）
- [x] `backend/cmd/server/main.go` のスケルトン作成
- [x] PoC 用 `message` コンテキスト（domain / usecase / handler / CSV persistence）
- [x] `backend/data/messages.csv` の追加
- [ ] レイヤー境界を崩さないための簡易テスト（domain / usecase 中心）
