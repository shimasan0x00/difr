# コードレビュー v6 - 修正レポート

**日付:** 2026-02-28
**対象:** diffff プロジェクト（Phase 1-5 全完了後）
**前提:** v1-v5 計91件修正済み

---

## 修正一覧 (12件)

### HIGH: リソース管理・並行性

#### 1. WebSocket ハンドラの reader.Close() を defer に変更
**File:** `internal/server/ws_claude.go`
- `streamClaudeEvents` が panic した場合に `reader.Close()` と `claudeCancel()` が呼ばれない問題を修正
- 即時実行される無名関数内で `defer reader.Close()` / `defer claudeCancel()` を使用

#### 2. Watcher の loop() に defer timer.Stop() + close(events) を追加
**File:** `internal/watcher/watcher.go`
- `loop()` 終了時に未発火の `time.AfterFunc` タイマーを停止
- `events` チャネルを close して、`Events()` の読み取り側に終了を通知

#### 3. streamClaudeEvents の引数型を io.Reader に変更
**File:** `internal/server/ws_claude.go`
- `interface{ Read([]byte) (int, error) }` を標準の `io.Reader` に置換

### MEDIUM: エラーハンドリング・堅牢性

#### 4. App.tsx の fetchDiff に AbortController を追加
**Files:** `web/src/App.tsx`, `web/src/api/client.ts`
- React StrictMode での unmount → remount 時に古いリクエストが状態を上書きする問題を防止
- `fetchDiff()` に `signal?: AbortSignal` パラメータを追加
- `useEffect` の cleanup で `controller.abort()` を呼び出し

#### 5. エクスポートに Content-Disposition ヘッダーを追加
**File:** `internal/server/handler_comment.go`
- JSON エクスポート: `attachment; filename="comments.json"`
- Markdown エクスポート: `attachment; filename="comments.md"`

#### 6. DiffViewer の React key を改善
**File:** `web/src/components/diff/DiffViewer.tsx`
- UnifiedView: `key={j}` → `key={\`${line.type}-${line.oldNumber ?? 0}-${line.newNumber ?? 0}-${j}\`}`
- SplitView: `key={i}` → `key={\`${pair.left?.oldNumber ?? 0}-${pair.right?.newNumber ?? 0}-${i}\`}`

#### 7. openBrowser の Wait エラーをログ出力
**File:** `internal/cli/cli.go`
- `_ = c.Wait()` → エラー時に `log.Printf` で出力

### MEDIUM: セキュリティ・データ整合性

#### 9. コメント作成時に fileIndex 存在チェックを追加
**File:** `internal/server/handler_comment.go`
- `filePath` が diff 内に存在しない場合、400 Bad Request を返す

### LOW: コード品質・スタイル

#### 8. embed_prod.go に意図コメントを追加
**File:** `internal/embed/embed_prod.go`
- ファイル存在チェックの `Open → Close` パターンに意図を明記

#### 11. watch フラグの説明を改善
**File:** `internal/cli/cli.go`
- `"Watch for file changes (experimental)"` → `"Watch for file changes (experimental, log only)"`

#### 12. fileIndex の不変性をコメントで文書化
**File:** `internal/server/server.go`
- 初期化後不変でロック不要であることをコメントで明記

#### 14. commentStore に saving ガードを追加
**File:** `web/src/stores/commentStore.ts`
- `addComment` の先頭で `if (get().saving) return` ガードを追加

---

## 対応しなかった項目 (2件、将来対応候補)

| # | 内容 | 理由 |
|---|------|------|
| 10 | SyntaxHighlight の extractInnerTokens を Shiki codeToTokens() API に移行 | 現行実装で十分動作、リスク低 |
| 13 | claudeStore disconnect() のメッセージクリア分離 | UX設計判断、低優先度 |

---

## テスト追加 (6件)

| テスト | ファイル | 内容 |
|--------|---------|------|
| `TestStreamClaudeEvents_ClosesDeferredOnPanic` | `ws_claude_test.go` | defer でリソースリーク防止を検証 |
| `TestWatcher_EventsChannelClosesAfterClose` | `watcher_test.go` | Close 後に events チャネルが閉じることを検証 |
| `TestCreateComment_RejectsFilePathNotInDiff` | `handler_comment_test.go` | diff に存在しないファイルパスで 400 エラー |
| `TestExportComments_IncludesContentDisposition` | `handler_comment_test.go` | Content-Disposition ヘッダー検証 |
| `generates unique keys for lines in unified view` | `DiffViewer.test.tsx` | React key 一意性検証 |
| `prevents duplicate submission while saving` | `commentStore.test.ts` | saving 中の二重送信防止 |

---

## テスト結果

- **Go:** 85テスト関数 + 18サブテスト = 103テストケース, 全Pass
- **Frontend:** 84テスト, 全Pass
- **合計:** 187テスト全Pass (v5比 +6)
