# Message MCP server（実験）

認証付き（共有シークレット）stdio MCP サーバー。ツール `get_message` で backend の `GET /` からメッセージを取得する。

Auth0 や会員セッションは使わない。stdio ではクライアントが Bearer を送れないため、**起動時に `MCP_API_KEY` を渡したプロセスだけがツールを提供する**形でゲートする。

## 前提

- Node.js 20+
- backend が起動していること（例: `http://localhost:8080`）

## セットアップ

```bash
cd mcp/message-server
cp .env.example .env   # 任意。Cursor 設定で env を渡す場合は不要
npm ci
```

## 起動（手動）

```bash
# PowerShell
$env:MCP_API_KEY = "dev-secret"
$env:API_BASE_URL = "http://localhost:8080"
npm start
```

`MCP_API_KEY` が空／未設定だとプロセスは即終了する。stdio サーバーのため、対話確認は Cursor 接続が本命。

## Cursor 設定例

Cursor の MCP 設定（例: `~/.cursor/mcp.json`）:

```json
{
  "mcpServers": {
    "whitekey-message": {
      "command": "npm",
      "args": ["start"],
      "cwd": "C:/Users/YOU/path/to/WhiteKey-Experimentation-Site/mcp/message-server",
      "env": {
        "MCP_API_KEY": "dev-secret",
        "API_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

`cwd` はリポジトリ内の `mcp/message-server` の絶対パスに置き換える。

接続後、ツール `get_message` を呼ぶとシード値（通常 `Hello World`）が返る。

## ツール

| 名前 | 引数 | 動作 |
| --- | --- | --- |
| `get_message` | なし | `GET {API_BASE_URL}/` の `message` を text で返す |

## 環境変数

| 変数 | 必須 | 既定 | 説明 |
| --- | --- | --- | --- |
| `MCP_API_KEY` | はい | — | 空なら起動失敗 |
| `API_BASE_URL` | いいえ | `http://localhost:8080` | backend のルート URL |
