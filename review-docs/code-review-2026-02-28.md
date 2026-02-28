# コードレビュー結果 — 2026-02-28

## 概要

difr コードベース全体のレビューにより、セキュリティ・エラーハンドリング・並行性・フロントエンド品質に関する29件の指摘を発見。25件を修正、4件を延期。

## 修正済み (25件)

### Group 1: セキュリティ修正 (3件)

| # | 内容 | ファイル |
|---|------|---------|
| 1 | Git引数コマンドインジェクション防止 (`validateRef` + `--` セパレータ) | `internal/git/diff.go` |
| 2 | WebSocket Origin検証 (`OriginPatterns` に変更) | `internal/server/ws_claude.go` |
| 3 | stdin読み込みサイズ制限 (100MB上限) | `internal/git/diff.go` |

### Group 2: エラーハンドリング修正 — Go (6件)

| # | 内容 | ファイル |
|---|------|---------|
| 4 | `server.New()` のエラー伝播 | `internal/server/server.go`, `internal/cli/cli.go`, テストファイル |
| 5 | `embed.Handler()` のエラー伝播 | `internal/embed/embed_dev.go`, `embed_prod.go` |
| 6 | Claude Runner の本番設定 (`Client.Run()` 実装) | `internal/claude/client.go`, `client_test.go` |
| 7 | コメント永続化 (CUD操作後に `Save()`) | `internal/server/handler_comment.go` |
| 8 | writeJSON エラーログ | `internal/server/handler_diff.go` |
| 9 | writeWS エラーログ | `internal/server/ws_claude.go` |

### Group 3: 並行性修正 (2件)

| # | 内容 | ファイル |
|---|------|---------|
| 10 | Watcher データレース修正 (`sync.Mutex` で `pending` 保護) | `internal/watcher/watcher.go` |
| 11 | Comment Store 防御的コピー (返却値をコピー) | `internal/comment/store.go` |

### Group 4: フロントエンド修正 (10件)

| # | 内容 | ファイル |
|---|------|---------|
| 12 | commentStore エラーハンドリング (try-catch + error state) | `web/src/stores/commentStore.ts` |
| 13 | CommentForm 送信後クリア | `web/src/components/comment/CommentForm.tsx` |
| 14 | DiffViewer `\|\|` → `??` (nullish coalescing) | `web/src/components/diff/DiffViewer.tsx` |
| 15 | DiffViewer 型安全性 (`FileStatus` 型使用) | `web/src/components/diff/DiffViewer.tsx` |
| 16 | SyntaxHighlight 安全性コメント (`dangerouslySetInnerHTML`) | `web/src/components/diff/SyntaxHighlight.tsx` |
| 17 | SyntaxHighlight `React.memo` ラップ | `web/src/components/diff/SyntaxHighlight.tsx` |
| 18 | App.tsx Zustand セレクタ (個別セレクタで再レンダリング最適化) | `web/src/App.tsx` |
| 19 | react-markdown 未使用依存の削除 | `web/package.json` |
| 20 | アクセシビリティ属性 (`aria-label`) | `CommentForm.tsx`, `ChatPanel.tsx` |
| 21 | 配列 index key の改善 (hunk に意味のあるkey使用) | `web/src/components/diff/DiffViewer.tsx` |

### Group 5: コード品質改善 (4件)

| # | 内容 | ファイル |
|---|------|---------|
| 22 | no-op `normalizePath` の削除 | `internal/diff/parser.go` |
| 23 | `detectLanguage` マップのパッケージレベル化 | `internal/diff/parser.go` |
| 24 | グレースフルシャットダウン (`signal.Notify` + `http.Server.Shutdown`) | `internal/cli/cli.go` |
| 25 | グローバル `cfg` の除去 (ローカル変数化、`GetConfig()` 削除) | `internal/cli/cli.go` |

## 延期 (4件)

| # | 内容 | 理由 |
|---|------|------|
| 26 | ChatPanel/ReviewButton/CommentのApp統合 | 大規模UI設計変更が必要 |
| 27 | Split viewのadd/deleteペアリング | DiffViewerの大幅な再構築が必要 |
| 28 | 大規模diff用の仮想化 | 新依存追加+レンダリングモデル変更 |
| 29 | Store/APIクライアントのテスト追加 | コード変更なし。専用テスト追加タスクとして実施 |

## テスト結果

### 新規テスト追加

| テスト | ファイル |
|--------|---------|
| `TestBuildDiffArgs_RejectsDashPrefixedRef` (3ケース) | `internal/git/diff_test.go` |
| `TestGet_ReturnedPointerDoesNotMutateStore` | `internal/comment/store_test.go` |
| `var _ Runner = (*Client)(nil)` (コンパイル時検証) | `internal/claude/client_test.go` |
| CommentForm クリア確認アサーション | `web/src/components/comment/CommentForm.test.tsx` |

### 最終テスト結果

```
Go:       全パッケージ Pass (レース検出なし)
Frontend: 34テスト 全Pass
```
