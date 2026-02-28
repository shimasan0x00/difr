# コードレビュー v4 (2026-02-28)

## 概要

バックエンド・フロントエンド・テスト/設定の全ファイルを対象に包括的レビューを実施。
v1-v3 で 66 件の修正が適用済みの状態に対し、新たに **49 件** の指摘事項を検出。

| 重要度 | 件数 | 対応状況 |
|--------|------|---------|
| Critical | 1 | ✅ 修正済み |
| High | 5 | ✅ 修正済み |
| Medium | 25 | 未対応 |
| Low | 18 | 未対応 |

---

## Critical (1件) — ✅ 修正済み

### C-01: `openBrowser` でゾンビプロセスが発生する
- **ファイル:** `internal/cli/cli.go:158-159`
- **問題:** `exec.Command().Start()` 後に `Wait()` を呼んでいないため、ブラウザ起動のたびに OS レベルでゾンビプロセスが残る
- **修正:** goroutine 内で `cmd.Wait()` を呼び出してプロセスを回収するように変更

---

## High (5件) — ✅ 修正済み

### H-01: WebSocket `writeWS` のエラーが無視され、切断後もストリーミング継続
- **ファイル:** `internal/server/ws_claude.go:102-146`
- **問題:** `writeWS` はエラーをログ出力するのみで戻り値がない。WebSocket 切断後も Claude CLI 出力を読み続け、無駄なリソースを消費する
- **修正:** `writeWS` がエラーを返すように変更。`streamClaudeEvents` 内で write エラー時に早期 return するように修正

### H-02: Claude CLI 実行に個別タイムアウトがない
- **ファイル:** `internal/server/ws_claude.go:61`
- **問題:** Claude CLI 実行に対するタイムアウトがなく、CLI がハングすると無期限にブロックする
- **修正:** Claude CLI 呼び出しに 5 分のタイムアウト付き context を設定

### H-03: `jsonArrayRegex` が誤マッチする可能性
- **ファイル:** `internal/claude/review.go:15`
- **問題:** 正規表現 `(?s)\[[\s\S]*?\{[\s\S]*?"filePath"[\s\S]*?\]` が最初の `]` にマッチしてしまい、JSON 配列が途中で切れる恐れがある
- **修正:** 正規表現ベースの抽出を廃止し、`[` を起点としたブラケットカウントによる JSON 配列境界検出に変更。抽出した候補を `json.Unmarshal` で検証する堅牢なアプローチに切り替え

### H-04: Comment Store の `Save()` が RLock で filesystem 書き込み
- **ファイル:** `internal/comment/store.go:139-143`
- **問題:** `Save()` は `RLock` のみ取得するが `saveLocked()` で filesystem に書き込む。複数の `Save()` が同時実行されるとファイル競合が発生する
- **修正:** `Save()` のロックを `RLock` → `Lock` に変更。コメントを明確化

### H-05: `golang.org/x/sys` が古い (v0.13.0)
- **ファイル:** `go.mod:19`
- **問題:** 2023 年 10 月頃のバージョンでセキュリティ修正を含む更新が存在
- **修正:** `go get golang.org/x/sys@latest && go mod tidy` で最新版に更新

---

## Medium (25件) — 未対応

### バックエンド (8件)

| ID | ファイル | 内容 |
|----|---------|------|
| M-01 | `ws_claude.go:76-88` | `SessionID` が `-` 始まりの場合、Claude CLI フラグとして誤認される可能性 |
| M-02 | `ws_claude.go:30-32` | WebSocket Origin が `localhost:*` と `127.0.0.1:*` のみ。`--host 0.0.0.0` 使用時に外部接続が拒否される |
| M-03 | `handler_diff.go:52-58` | `writeJSON` で status 書き込み後に JSON encode するため、encode 失敗時に 200 + 空 body が返る |
| M-04 | `handler_comment.go:111-128` | `handleExportComments` の `w.Write()` 戻り値未チェック |
| M-05 | `diff/parser.go:47-50` | 空 diff で `"files": null` が返る（フロントは配列を期待） |
| M-06 | `comment/store.go:44-51` | コメント ID が連番 `c1, c2...` で予測可能（列挙攻撃に弱い） |
| M-07 | `watcher/watcher.go:66-69` | `Close()` 二重呼び出しで panic（`sync.Once` 未使用） |
| M-08 | `watcher/watcher.go:121-132` | `Close()` 後にタイマー callback が発火し、`events` チャネルが閉じられないため consumer goroutine がリーク |

### フロントエンド (11件)

| ID | ファイル | 内容 |
|----|---------|------|
| M-09 | `claudeStore.ts:125-150` | `sendChat` / `sendReview` が WS 未接続時に無言で失敗（ユーザーフィードバックなし） |
| M-10 | `claudeStore.ts:48,93` | WebSocket `CONNECTING` 状態のガードなし。`connect()` 重複呼び出しで複数接続が発生 |
| M-11 | `claudeStore.ts:21-24,117` | `disconnect()` 後に `reconnectAttempt = MAX` を設定 → 再接続が永久無効に |
| M-12 | `commentStore.ts` / `App.tsx` | コメント操作の `error` / `saving` 状態がどこにも表示されない |
| M-13 | `client.ts:3-16` | API レスポンスのランタイムバリデーションなし |
| M-14 | `ChatPanel.tsx:38` | ストリーミング中にメッセージが in-place 更新されるのに `key={i}` で再レンダリングが不正確 |
| M-15 | `App.tsx:85-98` | Split/Unified トグルに `aria-pressed` がなく、スクリーンリーダーで状態不明 |
| M-16 | `ChatPanel.tsx` | ストリーミングコンテンツに `aria-live` 未設定 |
| M-17 | `SyntaxHighlight.tsx:26-63` | 大規模 diff で数百の concurrent `codeToHtml()` が発生しメインスレッドがブロック |
| M-18 | `DiffViewer.tsx:181-195` | `Record<string, string>` ではなく `Record<LineType, string>` にすべき |
| M-19 | `DiffViewer.tsx:108,265,277` | `!lineNumber` が `0` を falsy 扱い。`=== undefined` が正確 |

### テスト/設定 (6件)

| ID | ファイル | 内容 |
|----|---------|------|
| M-20 | `Taskfile.yml:43-44` | Go テストに `-race` フラグなし。並行安全なはずの comment store の race 検出漏れ |
| M-21 | `watcher_test.go:16,24,44...` | `os.WriteFile` のエラーを無視（t-wada 原則 #4 違反） |
| M-22 | `App.test.tsx` | `loadComments` の fetch 呼び出しが mock されていない |
| M-23 | Zustand stores 全般 | `claudeStore` / `commentStore` / `diffStore` のストア単体テストが皆無 |
| M-24 | `.gitignore` | `.env`, `.diffff/`, `.DS_Store` 等のパターン欠落 |
| M-25 | `store_test.go` | 破損 JSON ファイルの `Load` テストなし |

---

## Low (18件) — 未対応

| ID | ファイル | 内容 |
|----|---------|------|
| L-01 | `comment/store.go:206-212` | `commentSlice()` が map 反復で永続化順序が非決定的 |
| L-02 | `comment/store.go:157-173` | atomic write で `fsync` 未実行（電源障害時にデータ消失の可能性） |
| L-03 | `embed_prod.go:22-39` | SPA handler の特殊 URL パス（二重スラッシュ等）のエッジケース |
| L-04 | `handler_diff.go:24-39` | `handleGetDiffFileByPath` がファイル数分の線形探索 |
| L-05 | `watcher.go` | `events` チャネルが閉じられず consumer goroutine がリーク |
| L-06 | `index.html:8` | `<title>web</title>` → `<title>diffff</title>` にすべき |
| L-07 | `SyntaxHighlight.tsx:36` | `as never` キャストで型安全性が損なわれている |
| L-08 | `InlineComment.tsx:16` | `aria-label="Delete"` が不十分（何を削除するか不明） |
| L-09 | `CommentForm.tsx` | フォーム展開時に textarea へのフォーカス移動なし |
| L-10 | `App.tsx:112` | `key={file.newPath \|\| file.oldPath}` で重複キーの可能性 |
| L-11 | `DiffViewer.tsx:140` | Hunk key `${oldStart}-${newStart}` が衝突する可能性 |
| L-12 | `main.tsx:6` | `!` non-null assertion（標準パターンだが記録） |
| L-13 | `SyntaxHighlight.tsx:4-19` | Shiki highlighter が HMR で再生成される（開発時のみ） |
| L-14 | `cli_test.go:28-43` | type switch に `default` ケースなし |
| L-15 | `parse_test.go:87-92` | `From`/`To` の空値が明示的にアサートされていない |
| L-16 | `.air.toml:8` | `kill_delay = "0s"` で WebSocket 接続が即断 |
| L-17 | `.air.toml:5` | `exclude_dir` に `docs` 未含（機能的影響なし） |
| L-18 | `ws_claude_test.go` | `readWSResponses` がエラー種別を区別しない |

---

## 修正の詳細

### C-01: openBrowser ゾンビプロセス修正

**変更前:**
```go
if err := exec.Command(cmd, args...).Start(); err != nil {
    log.Printf("Failed to open browser: %v", err)
}
```

**変更後:**
```go
c := exec.Command(cmd, args...)
if err := c.Start(); err != nil {
    log.Printf("Failed to open browser: %v", err)
    return
}
go c.Wait()
```

### H-01/H-02: WebSocket エラーハンドリング + タイムアウト

- `writeWS` が `error` を返すように変更
- `streamClaudeEvents` 内で write エラー時に早期 return
- Claude CLI 呼び出しに 5 分のタイムアウト context を設定

### H-03: レビュー JSON パーサー改善

正規表現ベースの JSON 抽出を廃止し、ブラケットカウントによる配列境界検出に変更:
1. テキスト内の `[` を起点に `[]` のネスト深度をカウント
2. 深度 0 に戻った時点で候補文字列を抽出
3. `json.Unmarshal` で候補を検証
4. `filePath` フィールドの存在確認

### H-04: Save() ロック修正

`RLock` → `Lock` に変更し、concurrent な filesystem 書き込みを防止。

### H-05: 依存パッケージ更新

`golang.org/x/sys` を最新版に更新。
