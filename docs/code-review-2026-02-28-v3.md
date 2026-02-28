# コードレビュー結果 v3 — 2026-02-28

## 概要

v1 で 25件、v2 で 22件を修正済み。今回の包括的レビュー（Go vet + 全テスト Pass 確認済み）で追加 19件の指摘を発見。
**全19件対応完了。Go 98テスト (サブテスト含む) + Frontend 34テスト 全Pass。**

## 重要度: 高 (5件) — 全修正済み

| # | 箇所 | 問題 | 状態 |
|---|------|------|------|
| 1 | `internal/embed/embed_prod.go` | `spaHandler.ServeHTTP` で `fs.Open` したファイルを `Close()` していない（リソースリーク） | ✅ |
| 2 | `web/src/stores/claudeStore.ts` | WebSocket `onmessage` で `JSON.parse` が `try-catch` なし（不正データでクラッシュ） | ✅ |
| 3 | `internal/comment/store.go` | CUD操作とSaveが非アトミック → CUD内で自動永続化 + ロールバック | ✅ |
| 4 | `internal/server/ws_claude.go` | バッチ処理 → `streamClaudeEvents` で行単位リアルタイムストリーミング | ✅ |
| 5 | `go.mod` | `coder/websocket` と `fsnotify` が `// indirect` → `go mod tidy` で修正 | ✅ |

## 重要度: 中 — バックエンド (7件) — 全修正済み

| # | 箇所 | 問題 | 状態 |
|---|------|------|------|
| 6 | `internal/cli/cli.go` | `http.Server` タイムアウト設定 (Read:30s, Write:60s, Idle:120s) | ✅ |
| 7 | `internal/git/diff.go` | `validateRef` に空文字列チェック追加 | ✅ |
| 8 | `internal/server/handler_comment.go` | コメント作成: `filePath`/`body`必須 + `line>=1` / 更新: `body`必須 | ✅ |
| 9 | `internal/server/ws_claude.go` | `SetReadLimit(1MB)` でメッセージサイズ制限 | ✅ |
| 10 | `internal/claude/client.go` | `cmdReadCloser` に `stderr` バッファ追加、エラー時にメッセージ含む | ✅ |
| 11 | `internal/embed/embed_prod.go` | Cache-Control: assets → `immutable, 1年`, index.html → `no-cache` | ✅ |
| 12 | `internal/watcher/watcher.go` | `addRecursive` で再帰的ディレクトリ監視 (hidden/node_modules等スキップ) | ✅ |

## 重要度: 中 — フロントエンド (4件) — 全修正済み

| # | 箇所 | 問題 | 状態 |
|---|------|------|------|
| 13 | `web/src/App.tsx` | `useMemo` で `commentsByFile` Map を事前計算 | ✅ |
| 14 | `web/src/components/diff/DiffViewer.tsx` | `useMemo(() => pairLines(hunks), [hunks])` | ✅ |
| 15 | `web/src/stores/claudeStore.ts` | `onclose` で指数バックオフ再接続 (最大5回, 最大30秒) | ✅ |
| 16 | `web/src/stores/commentStore.ts` | `saving` 状態追加、CUD操作で `saving: true/false` 管理 | ✅ |

## 重要度: 中 — テスト (3件) — 全修正済み

| # | 箇所 | 問題 | 状態 |
|---|------|------|------|
| 17 | `internal/comment/store_test.go` | `store.Create` の全6箇所でエラーを `require.NoError` で検証 | ✅ |
| 18 | `internal/server/handler_comment_test.go` | +7テスト: Export(Markdown/JSON), Update不正JSON, Update空body, Create空filePath, Create line=0 | ✅ |
| 19 | `internal/server/ws_claude_test.go` | +4テスト: buildClaudeArgs (chat/session/review) + RunnerError | ✅ |

## 主な変更詳細

### Store 自動永続化 (#3)
- `Create`/`Update`/`Delete` 内でロック保持中に `saveLocked()` を呼び出し
- 永続化失敗時はメモリ上の変更をロールバック
- ハンドラ側の明示的 `Save()` 呼び出しを削除

### リアルタイムストリーミング (#4)
- `ParseStreamEvents` (バッチ) → `streamClaudeEvents` (逐次) に置き換え
- NDJSON を `bufio.Scanner` で1行ずつ読み、即座に WebSocket 送信
- `claude` パッケージへの依存を削除（inline の `streamEvent` 型使用）

### 再帰的ディレクトリ監視 (#12)
- `filepath.WalkDir` でサブディレクトリも自動追加
- `.git`, `node_modules`, `vendor`, `dist`, `.diffff` 等をスキップ

## テスト結果

```
Go:       全パッケージ Pass (98テスト, サブテスト含む)
Frontend: 34テスト 全Pass (7ファイル)
go vet:   警告なし
```

## 重要度: 低 (今回対応しない)

- `diff/parser.go`: 拡張子なしファイル（Makefile等）の言語検出
- `handler_diff.go`: `interface{}` → `any` への更新
- `handler_comment.go`: `w.Write` のエラー無視
- `cli.go`: `exec.Command.Start()` 後の `Wait()` 未呼出（ゾンビプロセス）
- `DiffViewer.tsx`: hunk.lines の key にインデックス使用、定数オブジェクトがコンポーネント内定義
- `CommentForm.tsx`: `<form>` 要素未使用（アクセシビリティ）
- フロントエンド Store / API クライアントの単体テストなし
- `claude/review.go`: 正規表現JSON抽出がネストされた `]` で誤動作の可能性

## 変更ファイル一覧

### Go (10ファイル)
- `internal/embed/embed_prod.go` — fs.Open Close, Cache-Control
- `internal/comment/store.go` — CUD自動永続化 + saveLocked + ロールバック
- `internal/server/ws_claude.go` — streamClaudeEvents, SetReadLimit
- `internal/server/handler_comment.go` — バリデーション追加, Save削除
- `internal/cli/cli.go` — HTTP タイムアウト設定
- `internal/git/diff.go` — validateRef 空文字列チェック
- `internal/claude/client.go` — stderr 捕捉
- `internal/watcher/watcher.go` — 再帰的ディレクトリ監視
- `go.mod` / `go.sum` — go mod tidy

### Frontend (4ファイル)
- `web/src/stores/claudeStore.ts` — JSON.parse try-catch, 再接続ロジック
- `web/src/stores/commentStore.ts` — saving 状態
- `web/src/App.tsx` — commentsByFile useMemo
- `web/src/components/diff/DiffViewer.tsx` — pairLines useMemo

### テスト (3ファイル)
- `internal/comment/store_test.go` — Create エラー検証
- `internal/server/handler_comment_test.go` — +7テスト
- `internal/server/ws_claude_test.go` — +4テスト, mockRunner改善
