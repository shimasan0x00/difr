# difr — インストール & ビルドガイド

## 前提条件

| ツール | バージョン | 用途 |
|--------|-----------|------|
| [Go](https://go.dev/dl/) | 1.25+ | バックエンドビルド |
| [Node.js](https://nodejs.org/) | 20+ | フロントエンドビルド |
| [Task](https://taskfile.dev/installation/) | 3.x | タスクランナー（ソースビルド時） |

> **`go install` でインストールする場合は Go のみで OK です。**

## インストール方法

### 方法 1: `go install`（推奨）

Go がインストール済みなら、1コマンドでインストールできます:

```bash
go install github.com/shimasan0x00/difr/cmd/difr@latest
```

`$GOPATH/bin/difr` にバイナリがインストールされます。

### 方法 2: ソースからビルド

```bash
git clone https://github.com/shimasan0x00/difr.git
cd difr

# 依存関係のインストール
task install

# 本番バイナリをビルド（フロントエンド込み単一バイナリ）
task build
```

カレントディレクトリに `difr` バイナリが生成されます。
任意の場所に配置してください:

```bash
# 例: /usr/local/bin に配置
sudo cp difr /usr/local/bin/

# または ~/go/bin に配置
cp difr ~/go/bin/
```

### 方法 3: 手動ビルド（Task なし）

Task をインストールしたくない場合は手動で実行できます:

```bash
git clone https://github.com/shimasan0x00/difr.git
cd difr

# Go 依存関係
go mod tidy

# フロントエンドビルド（dist を更新したい場合のみ）
cd web
npm ci
npm run build
cd ..
rm -rf internal/embed/dist
cp -r web/dist internal/embed/dist

# 本番バイナリビルド
go build -ldflags "-s -w" -o difr ./cmd/difr
```

## 動作確認

### 基本的な使い方

任意の Git リポジトリ内で実行すると、ブラウザが自動で開きます:

```bash
# 最新コミットの diff を表示
difr

# 特定コミットの diff
difr abc1234

# 2コミット間の diff
difr main feature/my-branch

# ステージング済み変更
difr staged

# 未ステージング（作業ツリー）の変更
difr working

# パイプ入力
git diff HEAD~3..HEAD | difr
```

### オプション

```bash
# ポートを変更（デフォルト: 3333）
difr -p 8080

# Unified 表示モード（デフォルト: split）
difr -m unified

# ブラウザ自動起動を抑制
difr --no-open

# Claude Code 連携を無効化
difr --no-claude

# ファイル変更を監視（experimental）
difr -w
```

### 試してみる（クイックスタート）

```bash
# 1. 適当な Git リポジトリに移動
cd ~/your-project

# 2. 最新コミットの diff をブラウザで表示
difr

# → http://127.0.0.1:3333 が自動で開く
# → Ctrl+C でサーバー停止
```

## 開発モード

開発に参加する場合:

```bash
cd difr

# 依存関係インストール
task install

# 開発サーバー起動（Go :3333 + Vite :5173 のホットリロード）
task dev

# テスト実行
task test              # 全テスト
task test:backend      # Go テストのみ
task test:frontend     # フロントエンドのみ

# リント
task lint
```

## トラブルシューティング

### `difr: command not found`

`$GOPATH/bin` が PATH に含まれているか確認:

```bash
echo $PATH | tr ':' '\n' | grep go
# ~/go/bin が表示されなければ PATH に追加
```

### ポートが使用中

別のポートを指定:

```bash
difr -p 8080
```

### Git リポジトリ外で実行した

difr は Git リポジトリ内で実行する必要があります（パイプ入力を除く）:

```bash
# パイプ入力なら任意のディレクトリで OK
git -C /path/to/repo diff | difr
```
