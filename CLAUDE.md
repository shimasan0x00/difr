# diffff - プロジェクトドキュメント

## 概要

GitHub/Azure DevOps 等のプラットフォームに依存しない、ローカルで動作するコードレビュー支援ツール。Git の差分をブラウザ上で GitHub 風に可視化し、Claude Code との連携（Chat + 自動レビュー）機能を提供する。

## 技術スタック

| 領域 | 技術 | バージョン |
|------|------|-----------|
| バックエンド | Go + Chi router | Go 1.25, Chi v5 |
| フロントエンド | React + TypeScript + Vite | React 19, Vite 7, TS 5.9 |
| スタイリング | Tailwind CSS v4 | @tailwindcss/vite |
| 状態管理 | Zustand | v5 |
| 構文ハイライト | Shiki | github-dark テーマ |
| Diffパーサー | go-gitdiff | v0.8 |
| テスト (Go) | testing + testify | testify v1.11 |
| テスト (FE) | Vitest + React Testing Library | Vitest v4 |
| CLI | cobra | v1.10 |
| 配布 | Go embed による単一バイナリ | - |

## ディレクトリ構成

```
diffff/
├── cmd/diffff/
│   └── main.go                     # エントリポイント
├── internal/
│   ├── cli/
│   │   ├── cli.go                  # cobra コマンド定義 + サーバー起動
│   │   ├── cli_test.go             # フラグデフォルト値、バリデーション (3件)
│   │   ├── parse.go                # ParseDiffRequest() - 引数/stdin解析
│   │   ├── parse_test.go           # 全DiffModeのパーステスト (10件)
│   │   └── types.go                # Config 型
│   ├── diff/
│   │   ├── types.go                # DiffMode, DiffRequest, DiffFile, DiffResult 型
│   │   ├── parser.go               # Parse() - go-gitdiff ラッパー + 言語検出
│   │   ├── parser_test.go          # testdataベースのパーステスト (7件)
│   │   └── testdata/               # simple_add.diff, multi_file.diff, etc.
│   ├── git/
│   │   ├── git.go                  # Client 構造体
│   │   ├── diff.go                 # GetDiff() - git diffコマンド実行
│   │   └── diff_test.go            # 実gitリポジトリでのテスト (7件)
│   ├── server/
│   │   ├── server.go               # Chi router + ルーティング設定
│   │   ├── handler_diff.go         # diff系APIハンドラ + writeJSON
│   │   ├── handler_diff_test.go    # httptest によるdiff APIテスト (6件)
│   │   ├── handler_comment.go      # コメントCRUD + エクスポートハンドラ + バリデーション
│   │   ├── handler_comment_test.go # コメントAPIテスト (14件)
│   │   ├── ws_claude.go            # WebSocket /ws/claude ハンドラ + リアルタイムストリーミング
│   │   └── ws_claude_test.go       # WebSocket Claude通信テスト (6件)
│   ├── comment/
│   │   ├── store.go                # JSONファイルベース Store (CRUD, 自動永続化, 並行安全)
│   │   ├── store_test.go           # Store CRUD/永続化/並行テスト (14件)
│   │   ├── export.go               # ExportMarkdown(), ExportJSON()
│   │   └── export_test.go          # エクスポートフォーマットテスト (4件)
│   ├── embed/
│   │   ├── embed_dev.go            # 開発: Vite (localhost:5173) へリバースプロキシ
│   │   └── embed_prod.go           # 本番: //go:embed all:dist + SPA fallback + Cache-Control
│   └── claude/
│       ├── client.go               # NewClient() - CLI可用性チェック
│       ├── client_test.go          # LookPathモックテスト (2件)
│       ├── stream.go               # ParseStreamEvents() - NDJSON パーサー
│       ├── stream_test.go          # stream-jsonパーステスト (6件)
│       ├── runner.go               # Runner インターフェース
│       ├── review.go               # ParseReviewComments() - レビュー結果抽出
│       ├── review_test.go          # レビューパーサーテスト (4件)
│       └── testdata/               # stream_chat.jsonl, stream_error.jsonl, review_response.json
├── web/
│   ├── src/
│   │   ├── main.tsx                # React エントリ
│   │   ├── App.tsx                 # メインコンポーネント (fetch → DiffViewer表示)
│   │   ├── App.test.tsx            # ローディング/エラー/表示テスト (5件)
│   │   ├── api/
│   │   │   ├── client.ts           # fetchDiff(), コメントCRUD API関数
│   │   │   └── types.ts            # DiffResult, DiffFile, Comment 等の型定義
│   │   ├── stores/
│   │   │   ├── diffStore.ts        # Zustand: files, viewMode, selectedFile
│   │   │   ├── commentStore.ts     # Zustand: comments, addComment, removeComment
│   │   │   └── claudeStore.ts      # Zustand: messages, sessionId, loading
│   │   ├── components/
│   │   │   ├── diff/
│   │   │   │   ├── DiffViewer.tsx          # メインdiffビューア (Split/Unified)
│   │   │   │   ├── DiffViewer.test.tsx     # diffビューアテスト (10件)
│   │   │   │   ├── SyntaxHighlight.tsx     # Shiki構文ハイライト (遅延ロード)
│   │   │   │   └── SyntaxHighlight.test.tsx # ハイライトテスト (4件)
│   │   │   ├── comment/
│   │   │   │   ├── CommentForm.tsx         # コメント入力フォーム (二重送信防止対応)
│   │   │   │   ├── CommentForm.test.tsx    # フォームテスト (5件)
│   │   │   │   ├── InlineComment.tsx       # インラインコメント表示
│   │   │   │   └── InlineComment.test.tsx  # コメント表示テスト (3件)
│   │   │   └── claude/
│   │   │       ├── ChatPanel.tsx           # チャットUI (入力+メッセージ表示)
│   │   │       ├── ChatPanel.test.tsx      # チャットテスト (6件)
│   │   │       ├── ReviewButton.tsx        # 自動レビューボタン
│   │   │       └── ReviewButton.test.tsx   # レビューボタンテスト (3件)
│   │   └── test/
│   │       └── setup.ts            # @testing-library/jest-dom
│   ├── package.json
│   ├── vite.config.ts              # React + Tailwind + proxy設定
│   └── vitest.config.ts            # jsdom環境, setupFiles
├── docs/
│   ├── code-review-2026-02-28.md         # コードレビュー v1 (25件修正)
│   ├── code-review-2026-02-28-v2.md      # コードレビュー v2 (22件修正)
│   ├── code-review-2026-02-28-v3.md      # コードレビュー v3 (19件修正)
│   ├── code-review-2026-02-28-v4.md      # コードレビュー v4 (13件修正)
│   ├── code-review-2026-02-28-v6.md      # コードレビュー v6 (12件修正)
│   └── code-review-2026-02-28-v7.md      # コードレビュー v7 (25件修正)
├── plan/
│   └── graceful-tumbling-stonebraker.md  # 実装計画書
├── Taskfile.yml                    # dev/build/test/clean タスク
├── .air.toml                       # Go ホットリロード設定
├── .gitignore
├── go.mod
└── go.sum
```

## CLI使用方法

```bash
diffff                        # 最新コミットのdiff (HEAD~1..HEAD)
diffff <commit>               # 特定コミットのdiff
diffff <from> <to>            # 2コミット間のdiff
diffff staged                 # ステージング済み変更
diffff working                # 未ステージング変更
git diff | diffff             # stdin パイプ入力

# オプション
--port, -p 3333               # サーバーポート
--host 127.0.0.1              # バインドアドレス
--mode, -m split|unified      # 表示モード
--no-open                     # ブラウザ自動起動抑制
--no-claude                   # Claude連携無効化
--watch, -w                   # ファイル変更監視 (default: false, experimental, log only)
```

## 開発コマンド

```bash
# 開発サーバー起動 (Go :3333 + Vite :5173)
task dev

# テスト
task test                     # 全テスト (Go + Frontend)
task test:backend             # Go テストのみ
task test:frontend            # Vitest のみ
task test:coverage            # カバレッジ付き

# ビルド
task build                    # 本番バイナリ生成 (単一バイナリ)

# その他
task clean                    # ビルド成果物削除
task lint                     # ESLint
task install                  # 全依存インストール
```

## アーキテクチャ

```
[ユーザー] → diffff CLI (cobra)
                ├── 引数/stdin解析 → DiffRequest
                ├── git diff実行 → rawDiff文字列
                └── HTTPサーバー起動 (Chi)
                    ├── /api/diff → DiffResult JSON (パース済み)
                    ├── /api/diff/files → ファイル一覧
                    ├── /api/diff/files/{path} → 個別ファイルdiff
                    ├── /api/diff/stats → 統計情報
                    ├── /api/comments → コメントCRUD
                    ├── /api/comments/export → Markdown/JSONエクスポート
                    ├── /ws/claude → (Phase 4)
                    └── /* → フロントエンド
                        ├── 開発: Vite dev serverへリバースプロキシ
                        └── 本番: embed.FS から配信

[ブラウザ] → React App
              ├── fetchDiff() → /api/diff → DiffResult
              ├── Zustand stores (diffStore, commentStore)
              ├── DiffViewer (Split/Unified + Shiki構文ハイライト)
              └── CommentForm / InlineComment
```

## APIエンドポイント

### 実装済み
| Method | Path | レスポンス |
|--------|------|-----------|
| GET | `/api/diff` | `DiffResult { files: DiffFile[], stats: FileStats }` |
| GET | `/api/diff/files` | `DiffFile[]` |
| GET | `/api/diff/files/{path}` | `DiffFile` |
| GET | `/api/diff/stats` | `{ totalFiles, stats: FileStats }` |
| POST | `/api/comments` | コメント作成 → `Comment` |
| GET | `/api/comments` | コメント一覧 (`?file=` フィルタ対応) |
| PUT | `/api/comments/{id}` | コメント更新 → `Comment` |
| DELETE | `/api/comments/{id}` | コメント削除 (204) |
| GET | `/api/comments/export` | エクスポート (`?format=json\|markdown`) |
| GET | `/api/claude/status` | `{ available: boolean }` |

| WS | `/ws/claude` | Chat + 自動レビュー WebSocket |

### 未実装
| Method | Path | 用途 | Phase |
|--------|------|------|-------|
| WS | `/ws/diff` | diff更新リアルタイム通知 | 5 |

## テスト現況

### Go (95テスト関数 + 32サブテスト = 127テストケース, 全Pass, -race有効)

| パッケージ | テスト数 | 内容 |
|-----------|---------|------|
| `internal/cli` | 5+20sub=25 | ParseDiffRequest全モード(9sub+1+空白stdin) + cobraフラグ(1+6sub)/バリデーション(2) + Portバリデーション(1+5sub) |
| `internal/git` | 8+3sub=11 | 実gitリポジトリで全DiffModeのdiff取得 + stdin + 空diff + refインジェクション防止(3sub) |
| `internal/diff` | 7 | testdataベースのdiffパース |
| `internal/server` | 36+9sub=45 | diff API(6) + Claude status(2) + コメントCRUD API(17) + WebSocket Claude(8) + sessionIDバリデーション(1+6sub) + sessionID境界値(1+3sub) |
| `internal/comment` | 21 | Store CRUD(9) + 永続化(2) + 並行アクセス(2) + ポインタ不変性(1) + SaveCreatesFile(1) + エクスポート(6) |
| `internal/claude` | 12 | CLI可用性(2) + stream-jsonパーサー(6) + レビューパーサー(4) |
| `internal/watcher` | 5 | ファイル変更検知(2) + デバウンス(1) + Close(1) + events チャネル close(1) |

### Frontend (91テスト, 全Pass)

| ファイル | テスト数 | 内容 |
|---------|---------|------|
| `App.test.tsx` | 5 | ローディング状態、diff表示、エラー表示、ヘッダー、空diff |
| `DiffViewer.test.tsx` | 11 | ファイルヘッダー、stats、add/delete/context行、行番号、split view、split左右検証、新規ファイル、バイナリ、key一意性 |
| `SyntaxHighlight.test.tsx` | 4 | コード表示、トークン化、未知言語、空文字列 |
| `CommentForm.test.tsx` | 6 | textarea/ボタン表示、submit、空body拒否、空白body拒否、cancel、saving無効化 |
| `InlineComment.test.tsx` | 4 | body表示、行番号、delete確認済み、deleteキャンセル |
| `ChatPanel.test.tsx` | 6 | 入力/送信、空メッセージ拒否、メッセージ表示、ローディング |
| `ReviewButton.test.tsx` | 3 | ボタン表示、クリック、ローディング状態 |
| `claudeStore.test.ts` | 22 | connect/disconnect、最大再接続回数、再接続、4種メッセージ処理、sendChat/sendReview、状態変更 |
| `commentStore.test.ts` | 11 | CRUD操作、エラーハンドリング、loading/savingフラグ、二重送信防止 |
| `diffStore.test.ts` | 5 | setFiles、setViewMode、setError、selectFile |
| `client.test.ts` | 14 | fetchDiff/fetchComments/createComment/updateComment/deleteComment + エラー + extractErrorMessage |

## テスト方針: TDD (Red → Green → Refactor)

1. **Red:** テストを先に書く。コンパイルエラーや失敗は期待通り
2. **Green:** テストを通す最小限のコードを書く
3. **Refactor:** テストが通る状態を維持しつつ品質改善
4. 各ステップで `go test ./...` または `npx vitest run` を実行してサイクル確認

## テスト品質ガイドライン (t-wada原則)

本プロジェクトでは以下の原則に従い、テストの信頼性・可読性・保守性を維持する。

### 1. テスト名は仕様を表す

テスト名は「何を」「どうしたとき」「どうなるか」を表現する。実装の詳細ではなく振る舞いを記述する。

```go
// ✗ 実装の詳細を名前にしている
func TestCreateAndGet(t *testing.T) {}

// ✓ 振る舞いを名前にしている
func TestCreate_AssignsIDAndTimestamp(t *testing.T) {}
func TestGet_ReturnsCreatedComment(t *testing.T) {}
```

```typescript
// ✗ 曖昧
it('renders review button', () => {})

// ✓ 具体的な振る舞い
it('renders with "Auto Review" label', () => {})
it('shows "Reviewing..." and is disabled when loading', () => {})
```

### 2. 1テスト1概念

1つのテスト関数では1つの概念のみを検証する。複数の独立した振る舞いを1テストに詰め込まない。

### 3. AAA (Arrange-Act-Assert) パターン

テストは3つのフェーズを明確に分離する。長いテストではコメントやブランクラインで区切りを示す。

```go
func TestUpdate_ChangesBodyAndSetsUpdatedAt(t *testing.T) {
    // Arrange
    store := newTestStore(t)
    created, _ := store.Create(&Comment{...})

    // Act
    updated, err := store.Update(created.ID, "new body")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "new body", updated.Body)
}
```

### 4. エラーを握りつぶさない

テスト内であっても `json.Unmarshal` や `Write` 等の戻り値エラーは必ず検証する。握りつぶすと偽陽性（本当は壊れているのにテストがパスする）の原因になる。

```go
// ✗ エラーを無視 — レスポンスが壊れていても気づけない
json.Unmarshal(w.Body.Bytes(), &c)

// ✓ エラーを検証
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
```

### 5. アサーションスタイルの統一

- **Go:** testify (`assert` / `require`) に統一。`require` は「これが失敗したら後続のアサーションが無意味」な箇所に使い、`assert` は検証の継続が可能な箇所に使う
- **Frontend:** `@testing-library/react` + `vitest` の `expect` に統一

### 6. テーブル駆動テスト (Go)

入力と期待値のパターンが異なるだけで構造が同じテストはテーブル駆動テストにまとめる。

```go
func TestParseDiffRequest(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantMode diff.DiffMode
    }{
        {"no args defaults to latest commit", []string{}, diff.DiffModeLatestCommit},
        {"staged keyword", []string{"staged"}, diff.DiffModeStaged},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, err := ParseDiffRequest(tt.args, nil)
            require.NoError(t, err)
            assert.Equal(t, tt.wantMode, req.Mode)
        })
    }
}
```

### 7. テストヘルパーには `t.Helper()` を付ける

テストヘルパー関数には必ず `t.Helper()` を呼ぶ。失敗時のスタックトレースがヘルパー内部ではなく呼び出し元を指すようになる。

```go
func newTestStore(t *testing.T) *Store {
    t.Helper()
    dir := t.TempDir()
    return NewStore(filepath.Join(dir, "comments.json"))
}
```

### 8. Frontend: `userEvent` を優先する

ユーザー操作のシミュレーションは `fireEvent` より `userEvent` を使う。`userEvent` はブラウザの実際のイベント発火順序を再現するため、より本物に近いテストになる。

```typescript
// ✗ fireEvent は低レベルすぎる
fireEvent.click(button)

// ✓ userEvent はユーザーの実操作に忠実
const user = userEvent.setup()
await user.click(button)
```

### 9. Frontend: グローバルモックは `vi.spyOn` + `afterEach` で安全に

`global.fetch = mockFn` のような直接代入はグローバル汚染の原因。`vi.spyOn` を使い `afterEach` で `mockRestore()` する。

```typescript
let fetchSpy: ReturnType<typeof vi.spyOn>

beforeEach(() => {
  fetchSpy = vi.spyOn(globalThis, 'fetch')
})

afterEach(() => {
  fetchSpy.mockRestore()
})
```

### 10. 具体的なアサーション

「存在する」だけでなく「何が」「どう」存在するかを検証する。常にtrueになるアサーションは意味がない。

```typescript
// ✗ container は常に truthy — テストとして無意味
expect(container).toBeTruthy()

// ✓ 具体的に何を期待しているかが明確
const span = container.querySelector('span')
expect(span).toBeInTheDocument()
expect(span?.textContent).toBe('')
```

## Go依存パッケージ

| パッケージ | 用途 | 状態 |
|-----------|------|------|
| `github.com/spf13/cobra` | CLI | 導入済み |
| `github.com/go-chi/chi/v5` | HTTP router | 導入済み |
| `github.com/stretchr/testify` | テスト | 導入済み |
| `github.com/bluekeyes/go-gitdiff` | Diff パーサー | 導入済み |
| `github.com/coder/websocket` | WebSocket | 導入済み |
| `github.com/fsnotify/fsnotify` | ファイル変更監視 | 導入済み |

## NPM依存パッケージ (主要)

| パッケージ | 用途 |
|-----------|------|
| `react`, `react-dom` | UI |
| `zustand` | 状態管理 |
| `shiki` | 構文ハイライト |
| `vitest`, `@testing-library/react` | テスト |
| `@testing-library/user-event` | ユーザー操作テスト |

## ビルド構成

- **開発モード** (`!production` build tag): `embed_dev.go` が `localhost:5173` へリバースプロキシ
- **本番モード** (`production` build tag): `embed_prod.go` が `//go:embed all:dist` で静的ファイル配信
- **ビルドフロー**: `npm run build` → `web/dist/` 生成 → `internal/embed/dist/` にコピー → `go build -tags production` → コピー削除

## 既知の注意点

- グローバルgit hookがmainブランチへの直接コミットを禁止しているため、Goテストでのテストリポジトリでは`test-branch`を使用
- Shiki highlighterはシングルトンで遅延初期化（初回ロード時のみWASM初期化）
- コメントStoreはsync.RWMutexで並行安全。CUD操作は内部で自動永続化（saveLocked）し、失敗時はロールバック
- サーバー起動時にdiffをパースしてDiffResultとしてメモリに保持
- WebSocket Claude通信はリアルタイムストリーミング（NDJSON行単位で即座にWS送信）
- HTTPサーバーにタイムアウト設定済み（Read:30s, Write:60s, Idle:120s）
- コメント作成/更新にバリデーション（filePath/body必須, line>=1, fileIndex存在チェック）
- 本番embed: Cache-Control設定済み（assets: immutable 1年, index.html: no-cache）
- ファイル監視は再帰的（サブディレクトリ含む、.git/node_modules等スキップ）
- WebSocket再接続は指数バックオフ（最大5回、最大30秒間隔）
- Claude CLIのstderrを捕捉しエラー時に含める
- WebSocket Claude: reader.Close()/claudeCancel() は defer で確実にリソース解放
- エクスポートAPIにContent-Dispositionヘッダー付与（ダウンロードファイル名指定）
- App.tsx: AbortController で unmount 時にfetchをキャンセル
- commentStore: addComment/updateComment/removeComment に saving ガードで二重送信防止
- Watcher: events チャネルの close を `sync.Once` で保護、`trySendEvent()` で closed channel send を recover
- WebSocket sessionId: 正規表現 `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$` でバリデーション
- ws_claude.go: `claude.StreamEvent` を再利用（ローカル `streamEvent` 型廃止）
- SyntaxHighlight: `codeToTokens` API 使用（HTML正規表現パース廃止）
- 全 Go テストに `-race` フラグ有効
- comment Store: `saveLocked` で `tmp.Sync()` 追加（atomic write の安全性向上）
- comment Store: `commentSlice()` が ID でソート（決定論的JSON出力）
- comment Store: `Load` 時に ID/FilePath/Line の妥当性検証
- Claude タイムアウト: `server.WithClaudeTimeout()` で設定可能（デフォルト5分）
- Port バリデーション: 1-65535 の範囲チェック
- API client: エラーレスポンスのボディを解析してサーバーのメッセージを表示
- コメント削除: `window.confirm` による確認ダイアログ
- ChatPanel: ストリーミング中は `scrollIntoView({ behavior: 'instant' })` でスクロール
- ChatPanel: Claude レスポンスのコードブロック表示対応 (`SimpleMarkdown`)
- コメントボタン: `focus:text-blue-400` でキーボードアクセシビリティ改善
- claudeStore: `_resetModuleState()` でテスト間のモジュールレベル変数リセット
- handler_comment: CUD エラーを `slog.Error` で記録
- 全 `log.Printf` を `log/slog` 構造化ログに移行
- `writeJSON` の `interface{}` → `any` に変更 (Go 1.25)
- StatusBadge: `aria-label` を `File status: ${status}` に改善
- Split/Unified トグルボタンに `focus-visible` スタイル追加
- `openBrowser` に `context.WithTimeout(15秒)` 追加
- claudeStore: `isReconnecting` フラグで自動再接続と手動接続を区別（再接続上限バグ修正）
- App.tsx: `commentsByFile` useMemo で不変データパターン (`[...existing, c]`) に修正

---

## 実装ロードマップ

### Phase 1: 基盤 ✅ 完了

| ステップ | 内容 | テスト | 状態 |
|---------|------|--------|------|
| 1-1 | CLI引数パース (cobra + ParseDiffRequest) | 13件 | ✅ |
| 1-2 | Git diff取得 (os/exec, 全6モード) | 7件 | ✅ |
| 1-3 | stdin検出 (os.Stdin.Stat) | parse_test内 | ✅ |
| 1-4 | HTTPサーバー + GET /api/diff (Chi) | 2件 | ✅ |
| 1-5 | フロントエンド初期化 + ビルドシステム | 5件 | ✅ |

### Phase 2: Diff表示 ✅ 完了

| ステップ | 内容 | テスト | 状態 |
|---------|------|--------|------|
| 2-1 | Diffパーサー (go-gitdiff) | 7件 | ✅ |
| 2-2 | ファイル別API | 6件 | ✅ |
| 2-3 | DiffViewerコンポーネント (Split/Unified) | 9件 | ✅ |
| 2-4 | 構文ハイライト (Shiki) | 4件 | ✅ |

### Phase 3: コメント機能 ✅ 完了

| ステップ | 内容 | テスト | 状態 |
|---------|------|--------|------|
| 3-1 | コメントStore (JSON永続化, 並行安全) | 11件 | ✅ |
| 3-2 | コメントAPI (CRUD + エクスポート) | 8件 | ✅ |
| 3-3 | コメントUI (CommentForm, InlineComment) | 7件 | ✅ |
| 3-4 | エクスポート (Markdown/JSON) | 4件 | ✅ |

### Phase 4: Claude Code連携 ✅ 完了

| ステップ | 内容 | テスト | 状態 |
|---------|------|--------|------|
| 4-1 | Claude CLI可用性チェック | 2件 | ✅ |
| 4-2 | stream-jsonパーサー (NDJSON) | 6件 | ✅ |
| 4-3 | 自動レビューレスポンスパーサー | 4件 | ✅ |
| 4-4 | WebSocket Claude通信 | 2件 | ✅ |
| 4-5 | ChatPanel / ReviewButton UI | 9件 | ✅ |

**Claude Code呼び出しパターン:**
```bash
claude -p "<prompt>" --output-format stream-json
claude -r <session-id> -p "<prompt>" --output-format stream-json
claude -p "<review prompt with diff>" --output-format stream-json --max-turns 1
```

### Phase 5: 仕上げ ✅ 完了

| ステップ | 内容 | テスト | 状態 |
|---------|------|--------|------|
| 5-1 | ファイル変更監視 (fsnotify + デバウンス) | 4件 | ✅ |
| 5-2 | 本番ビルド検証 (フロントエンド + Goバイナリ) | - | ✅ |
| 5-3 | 統合テスト + 最終確認 (Go 127 + FE 91 = 218テスト全Pass) | - | ✅ |
