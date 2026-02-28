# コードレビュー v7 - difr (2026-02-28)

Phase 1-5 全完了後、v1-v6 で計103件修正済みの状態に対する包括的コードレビュー。

## 修正サマリ

| 重要度 | ID | 内容 | 状態 |
|--------|-----|------|------|
| **High** | H-1 | Watcher `AfterFunc` コールバックの closed channel panic 防止 (`sync.Once` + `trySendEvent` recover) | ✅ |
| **High** | H-2 | claudeStore `connect()` で `reconnectAttempt` をリセット | ✅ |
| **High** | H-3 | `sendChat`/`sendReview` が WS 切断時にエラーメッセージを表示 | ✅ |
| **High** | H-4 | WebSocket `sessionId` のバリデーション (正規表現チェック) | ✅ |
| **High** | H-5 | `.gitignore` に `.env`, `*.pem`, `*.key`, `.difr/` 追加 | ✅ |
| **High** | H-6 | Go テストに `-race` フラグ追加 (Taskfile.yml) | ✅ |
| **Medium** | M-1 | ハンドラ内でエラーをログに記録 (`log.Printf`) | ✅ |
| **Medium** | M-2 | `streamEvent` 型の重複定義を `claude.StreamEvent` に統一 | ✅ |
| **Medium** | M-3 | ChatPanel スクロールをストリーミング中は `instant` に | ✅ |
| **Medium** | M-4 | ストリーミング時の配列コピー最適化 (`slice(0, -1)` + 新オブジェクト) | ✅ |
| **Medium** | M-5 | Split view で削除行にもコメントボタン追加 | ✅ |
| **Medium** | M-6 | `updateComment`/`removeComment` に二重送信防止ガード追加 | ✅ |
| **Medium** | M-7 | API エラーレスポンスのボディ解析 (`extractErrorMessage`) | ✅ |
| **Medium** | M-8 | コメント読み込みエラーを UI に表示 | ✅ |
| **Medium** | M-9 | Port バリデーション (1-65535) 追加 | ✅ |
| **Medium** | M-10 | `saveLocked` で rename 前に `tmp.Sync()` 追加 | ✅ |
| **Medium** | M-11 | SyntaxHighlight を `codeToTokens` API に移行 (HTML正規表現パース廃止) | ✅ |
| **Low** | L-1 | `commentSlice()` の出力を ID でソート (決定論的JSON出力) | ✅ |
| **Low** | L-2 | `Load` 時のコメントデータ検証 (ID/FilePath/Line) | ✅ |
| **Low** | L-3 | `selectedFile` 未使用 — 将来使用のため保留 | — |
| **Low** | L-4 | Claude タイムアウトを設定可能に (`WithClaudeTimeout` Option) | ✅ |
| **Low** | L-5 | `saveLocked` のコメントを「write lock必須」に修正 | ✅ |
| **Low** | L-6 | `DiffModeCommit` 初回コミット — 保留 (稀なケース) | — |
| **Low** | L-7 | モジュールレベル可変状態のリセット関数 (`_resetModuleState`) 追加 | ✅ |
| **Low** | L-8 | コメント削除確認ダイアログ (`window.confirm`) 追加 | ✅ |
| **Low** | L-9 | ChatPanel の Claude レスポンスにコードブロック表示対応 | ✅ |
| **Low** | L-10 | キーボードアクセシビリティ (`focus:text-blue-400`) 追加 | ✅ |

## 修正統計

- **修正済み:** 25件 (High: 6, Medium: 11, Low: 8)
- **保留:** 2件 (L-3: 将来使用, L-6: 稀なケース)
- **テスト追加:** 新規テスト 3件 (H-4: sessionID validation 6 subtests, L-8: confirm dialog 1 test, H-3: error assertion update 2 tests)

## テスト結果

- **Go:** 109テストケース (103 → 109, +6 sessionID validation subtests), 全Pass, race condition 0件検出
- **Frontend:** 85テスト (84 → 85, +1 delete cancel test), 全Pass

## 主な設計変更

### H-1: Watcher race condition 修正
- `sync.Once` で events チャネルの close を保護
- `trySendEvent()` で closed channel への send を recover で安全に処理

### H-4: sessionId バリデーション
- `buildClaudeArgs` が error を返すように変更 (シグネチャ変更)
- 正規表現 `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$` でフラグインジェクション防止

### M-2: streamEvent 型統一
- `ws_claude.go` のローカル `streamEvent` を廃止、`claude.StreamEvent` を再利用

### M-11: SyntaxHighlight API 移行
- `codeToHtml` → `codeToTokens` に変更
- HTML 正規表現パース (`extractInnerTokens`) を完全廃止
- React 要素としてトークンを直接レンダリング

### L-4: Claude タイムアウト設定
- `server.WithClaudeTimeout(duration)` Option 追加
- デフォルト 5分、server.go で `defaultClaudeTimeout` として定義
