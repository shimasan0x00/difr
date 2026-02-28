# コードレビュー結果 v2 — 2026-02-28

## 概要

前回レビュー（v1）で25件修正済み。今回の包括的レビューで追加の22件を発見。
**全22件対応完了。全88テスト Pass。**

## 対応状況

### CRITICAL (1件)

| # | 内容 | ファイル | 状態 |
|---|------|---------|------|
| C1 | プロセスゾンビリーク → `cmdReadCloser` ラッパーで `cmd.Wait()` 保証 | `internal/claude/client.go` | ✅ |

### HIGH (4件)

| # | 内容 | ファイル | 状態 |
|---|------|---------|------|
| H1 | 未使用CLIフラグ → 全4フラグ接続 (`--mode`, `--no-open`, `--no-claude`, `--watch`) | `internal/cli/cli.go`, `internal/server/server.go` | ✅ |
| H2 | コメントStore非アトミック書き込み → tmpfile+rename方式 | `internal/comment/store.go` | ✅ |
| H3 | Save失敗がHTTPレスポンスに反映されない → 500返却 | `internal/server/handler_comment.go` | ✅ |
| H4 | SPA fallbackなし → `spaHandler` でindex.htmlにfallback | `internal/embed/embed_prod.go` | ✅ |

### MEDIUM (11件)

| # | 内容 | ファイル | 状態 |
|---|------|---------|------|
| M1 | Claude Chat/ReviewをApp.tsxに統合 | `web/src/App.tsx` | ✅ |
| M2 | コメントUIをDiffViewerに統合 | `web/src/components/diff/DiffViewer.tsx`, `App.tsx` | ✅ |
| M3 | bufio.Scanner 64KB→1MBに引き上げ | `internal/claude/stream.go` | ✅ |
| M4 | 非イディオマティックタイムアウト → `5*time.Second` | `internal/cli/cli.go` | ✅ |
| M5 | リクエストボディサイズ制限 → `MaxBytesReader` (1MB) | `internal/server/handler_comment.go` | ✅ |
| M6 | Split View行ペアリング → `pairLines()` でGitHub風横並び表示 | `web/src/components/diff/DiffViewer.tsx` | ✅ |
| M7 | ChatPanel `<input>` → `<textarea>` + 自動スクロール | `web/src/components/claude/ChatPanel.tsx` | ✅ |
| M8 | commentStoreに `updateComment` アクション追加 | `web/src/stores/commentStore.ts` | ✅ |
| M9 | claudeStoreにWebSocket接続管理 (connect/disconnect/sendChat/sendReview) | `web/src/stores/claudeStore.ts` | ✅ |
| M10 | `ParseReviewComments` の不要な `error` 戻り値削除 | `internal/claude/review.go` | ✅ |
| M11 | `List()` を `CreatedAt` でソート | `internal/comment/store.go` | ✅ |

### LOW (6件)

| # | 内容 | ファイル | 状態 |
|---|------|---------|------|
| L1 | `ExportJSON` エラー伝播 → `(string, error)` シグネチャ | `internal/comment/export.go` | ✅ |
| L2 | テストの `os.Chdir` 排除 → `WithWorkDir(t.TempDir())` 方式 | `internal/server/*_test.go` | ✅ |
| L3 | `handleGetDiffFileByPath` → `chi.URLParam(r, "*")` | `internal/server/handler_diff.go` | ✅ |
| L4 | `lint:backend` タスク追加 (`go vet`) | `Taskfile.yml` | ✅ |
| L5 | 不要ファイル削除 (`App.css`, `react.svg`) | `web/src/` | ✅ |
| L6 | `.gitignore` に `node_modules/` 追加 | `.gitignore` | ✅ |

## 主な変更詳細

### バックエンド設定リファクタ (H1 + L2)
- `server.New()` に Option パターン導入: `WithWorkDir()`, `WithNoClaude()`, `WithViewMode()`
- `os.Getwd()` 依存を排除し、テストで `os.Chdir` 不要に
- `--no-open`: `openBrowser()` でOS別にブラウザ自動起動
- `--watch`: `watcher.New()` でファイル変更検知を統合
- `--no-claude`: `WithNoClaude(true)` でClaude初期化をスキップ
- `--mode`: `WithViewMode()` + `GET /api/diff/mode` エンドポイント追加

### フロントエンドUI統合 (M1 + M2 + M6)
- App.tsx: ヘッダーにSplit/Unified切替トグル、ReviewButton、サイドバーにChatPanel
- DiffViewer: インラインコメント表示・追加・削除機能を統合。行ホバーで「+」ボタン表示
- Split View: `pairLines()` アルゴリズムで連続delete/addをペアリングしGitHub風横並び表示

### WebSocket接続管理 (M9)
- `claudeStore` に `connect()`, `disconnect()`, `sendChat()`, `sendReview()` アクション追加
- `WSResponse` イベントに応じてメッセージ蓄積・セッション管理・エラーハンドリング
- テキストストリーミング: 同一assistantメッセージに追記

## テスト結果

```
Go:       54テスト 全Pass (8パッケージ)
Frontend: 34テスト 全Pass (7ファイル)
合計:     88テスト 全Pass
```

## 変更ファイル一覧

### Go (11ファイル)
- `internal/cli/cli.go` — CLIフラグ接続, ブラウザ自動起動, watcher統合
- `internal/cli/types.go` — 変更なし (Config型は既に定義済み)
- `internal/server/server.go` — Optionパターン, viewMode, /api/diff/mode
- `internal/server/handler_diff.go` — chi.URLParam使用
- `internal/server/handler_comment.go` — MaxBytesReader, Save失敗500, ExportJSON error
- `internal/server/handler_diff_test.go` — os.Chdir排除
- `internal/server/handler_comment_test.go` — os.Chdir排除
- `internal/server/ws_claude_test.go` — os.Chdir排除
- `internal/claude/client.go` — cmdReadCloser
- `internal/claude/stream.go` — scanner.Buffer 1MB
- `internal/claude/review.go` — error戻り値削除
- `internal/comment/store.go` — atomic write, List()ソート
- `internal/comment/export.go` — ExportJSON error伝播
- `internal/embed/embed_prod.go` — SPA fallback

### Frontend (7ファイル)
- `web/src/App.tsx` — Claude/コメント/ViewMode統合
- `web/src/stores/claudeStore.ts` — WebSocket接続管理
- `web/src/stores/commentStore.ts` — updateComment追加
- `web/src/components/diff/DiffViewer.tsx` — コメント統合, Split Viewペアリング
- `web/src/components/claude/ChatPanel.tsx` — textarea化, 自動スクロール
- `web/src/components/claude/ChatPanel.test.tsx` — textarea対応

### 設定 (3ファイル)
- `Taskfile.yml` — lint:backend追加
- `.gitignore` — node_modules/追加
- 削除: `web/src/App.css`, `web/src/assets/react.svg`
