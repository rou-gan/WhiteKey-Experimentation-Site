---
name: semgrep-review
description: >-
  Run Semgrep on backend/frontend for Phase A advisory review after push or PR
  creation. Use with /review-security; do not treat as a merge gate (GHA Semgrep
  is the deterministic gate).
---

# Semgrep レビュー（Phase A / advisory）

push または PR 作成**直後**のセキュリティレビューで、変更のあった `backend/` / `frontend/` を Semgrep でスキャンする。

## 前提

- マージゲートは GitHub Actions の `Semgrep` workflow（`semgrep scan --error`）。本 Skill の結果は **advisory**。
- Semgrep Cloud / `SEMGREP_APP_TOKEN` は不要。
- CVE の再列挙（Trivy / Grype と重複）はしない。パターン指摘のみ載せる。

## 実行手順

1. Semgrep MCP の `semgrep_scan` があればそれを使う。なければ CLI:

```bash
semgrep scan --metrics=off \
  --config p/security-audit \
  --config p/owasp-top-ten \
  --config p/golang \
  --config p/javascript \
  backend frontend
```

差分のみ見る場合は、変更パスに絞ってよい（例: `backend/internal/handler`）。

2. ERROR / WARNING 相当を Severity 表に要約する（同一 finding の長文コピペは避ける）。
3. `internal/experiment/` や実験意図が PR に明示されている指摘は Critical 扱いにしない。
4. `/review-security` の PR コメントに Semgrep 節を含め、`gh pr review --comment` で投稿する（`APPROVE` / `REQUEST_CHANGES` は使わない）。

## ローカルセットアップ（人間）

1. Semgrep CLI: `python3 -m pip install semgrep` または Cursor Semgrep plugin + `/setup-semgrep-plugin`
2. （任意）`npx skills add semgrep/skills`
