# difr

ローカル動作のコードレビュー支援ツール。Git diff をブラウザで GitHub 風に可視化 + Claude Code 連携。

## 技術スタック

- **バックエンド:** Go 1.25 / Chi v5 / cobra
- **フロントエンド:** React 19 / TypeScript 5.9 / Vite 7 / Tailwind CSS v4 / Zustand v5 / Shiki
- **テスト:** testify (Go) / Vitest + React Testing Library (FE)
- **配布:** Go embed による単一バイナリ

## 開発ルール

### Git ブランチ

- **main への直接コミット・push は禁止**（グローバル hook で強制）
- 開発は必ず `feature/<作業名>` ブランチを作成して行う
- テスト用 git リポジトリでは `test-branch` を使用（hook 回避）

### テスト方針: TDD + t-wada 原則

1. **TDD:** Red → Green → Refactor。各ステップで `go test -race ./...` / `npx vitest run` 確認
2. テスト名は仕様（振る舞い）を表す。1テスト1概念。AAA パターン
3. エラーを握りつぶさない（`require.NoError` で必ず検証）
4. Go: testify (`assert`/`require`) + テーブル駆動テスト + `t.Helper()`
5. FE: `userEvent` 優先、`vi.spyOn` + `afterEach` でモック管理、具体的アサーション

## 開発コマンド

```bash
task dev              # Go :3333 + Vite :5173
task test             # 全テスト (Go + FE)
task test:backend     # Go テストのみ (-race 有効)
task test:frontend    # Vitest のみ
task build            # 本番バイナリ (単一バイナリ)
task clean            # ビルド成果物削除
```

## アーキテクチャ

```
CLI (cobra) → git diff → HTTPサーバー (Chi)
  ├── /api/diff          DiffResult JSON
  ├── /api/diff/files    ファイル一覧 / 個別取得
  ├── /api/diff/stats    統計情報
  ├── /api/files/*       ファイルコンテンツ配信 (tracked のみ)
  ├── /api/comments      コメント CRUD + エクスポート
  ├── /api/claude/status Claude CLI 可用性
  ├── /ws/claude         Chat + 自動レビュー (WebSocket)
  └── /*                 フロントエンド (dev: proxy, prod: embed)

React App → Zustand stores
  ├── DiffViewer (Split/Unified + Shiki) + CommentForm
  ├── FileViewer (未変更ファイル表示 + Shiki)
  ├── FileListPanel (Changed / All Files タブ)
  │   └── DirectoryTree (IDE 風ディレクトリツリー)
  └── ChatPanel
```

## ビルド構成

- **開発:** `embed_dev.go` (`dev` tag) → Vite dev server へリバースプロキシ
- **本番:** `embed_prod.go` (`!dev` tag, デフォルト) → `//go:embed all:dist` で配信
- **フロー:** `npm run build` → `web/dist/` → `internal/embed/dist/` にコピー・コミット → `go build` (タグ不要)

## CI/CD

- **CI (`ci.yml`):** PR / main push で実行。Lint → Test Backend → Test Frontend → E2E (PR時のみ)
- **Release (`release.yml`):** `v*` タグ push で実行。test → build-frontend → build (5プラットフォーム) → publish

### E2E テスト作成時の注意点

- **ロケーター:** `getByText()` はサイドバーと本体で重複マッチしやすい。ファイル名には `locator('[id="diff-file-xxx"]')` を使う
- **CSS hover 依存の UI:** `group-hover` で表示されるボタンは headless Chrome では hover なしだとクリックできない。必ず `element.hover()` してから操作する
- **コンポーネントの確認 UI:** `InlineComment` の削除は `confirm()` ダイアログではなくインライン確認ボタン（"Confirm delete"）。テストではコンポーネントの実装を確認してから操作手順を書く
- **モック NDJSON フォーマット:** `StreamEvent.ContentBlocks()` は `message.content` を参照する。モックの JSON は `{"type":"assistant","message":{"content":[...]}}` の形式にする（`"content"` を直接トップレベルに置かない）
- **CI 環境のタイミング:** WebSocket 接続確立を待つ（`connection-indicator`）。レスポンス待ちは余裕を持ったタイムアウト (20s) を設定する

## 主要な設計判断

- 型定義は `internal/diff/types.go` に集約（循環 import 回避）
- コメント Store: `sync.RWMutex` 並行安全、CUD で自動永続化 + ロールバック
- WebSocket Claude: NDJSON 行単位リアルタイムストリーミング、指数バックオフ再接続（最大5回）
- HTTP タイムアウト: Read 30s / Write 60s / Idle 120s
- Claude タイムアウト: `WithClaudeTimeout()` でカスタマイズ可（デフォルト5分）
- `log/slog` 構造化ログ統一
- ファイルブラウザ: `trackedIndex` (git 管理下ファイルのホワイトリスト) + `filepath.IsLocal()` + symlink 解決でセキュリティ確保。5MB 上限、バイナリ判定 (null バイト + UTF-8 検証)
- サイドバー: Changed / All Files タブ切り替え。All Files は `buildFileTree()` でフラットパスからツリー構造を構築、`DirectoryTree` で再帰表示
- FileViewer: 未変更ファイルの全文表示。`fileContentCache` (Zustand Map) でキャッシュ。言語検出は拡張子マッピング (`langMap`)
