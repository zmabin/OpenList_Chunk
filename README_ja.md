<div align="center">
  <img src="https://raw.githubusercontent.com/OpenListTeam/Logo/main/logo.svg" width="128" height="128" alt="logo" />

  <h1>OpenList_Chunk</h1>

  <p><em>OpenList の拡張フォーク — チャンクアップロードで CDN アップロードサイズ制限を回避</em></p>

  <a href="https://github.com/zmabin/OpenList_Chunk/actions?query=workflow%3ABuild"><img src="https://img.shields.io/github/actions/workflow/status/zmabin/OpenList_Chunk/build.yml?branch=main" alt="Build status" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/releases"><img src="https://img.shields.io/github/v/release/zmabin/OpenList_Chunk" alt="latest version" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/blob/main/LICENSE"><img src="https://img.shields.io/github/license/zmabin/OpenList_Chunk" alt="License" /></a>
</div>

---

[English](./README.md) | [中文](./README_cn.md) | **日本語** | [Nederlands](./README_nl.md)

---

## 概要

**OpenList_Chunk** は [OpenList](https://github.com/OpenListTeam/OpenList) の拡張フォークで、アップロードロジックをリファクタリングしつつ、すべての元のデータ構造を維持しています。

**核心目標：CDN リバースプロキシによるアップロードサイズ制限を回避（例：Cloudflare 無料プランは単一リクエストを 100MB に制限）。**

**そのまま置き換えるだけ — 手間なし。**

---

## コア変更：CDN 制限の回避

本プロジェクトは、CDN アップロードボディ制限を回避する2つの異なるメカニズムを実装しています。

### 1. フォームチャンクアップロード

**「セッション管理 + ディスクキャッシュ + ストリーミングマージ」** に基づく従来の高互換性チャンク機構。

- **ワークフロー**：
  1. **セッション初期化**：フロントエンドが `POST /api/fs/put/chunk/init` を呼び出し、バックエンドが一意の `upload_id` を生成してセッションファイルを作成。
  2. **チャンクアップロード**：各チャンクは `multipart/form-data` として `PUT /api/fs/put/chunk` に `upload_id` と `index` を付けて送信。
  3. **CRC32 検証**：サーバーは各チャンクの CRC32 を計算し、クライアントの `X-Chunk-CRC32` ヘッダーと比較。
  4. **仮想マージ**：全チャンクのアップロード後、フロントエンドが `POST /api/fs/put/chunk/merge` を呼び出し。バックエンドは `io.MultiReader` を使用して全一時ファイルを順次読み込み、ディスクコピーなしでストレージバックエンドに直接ストリーミング。
  5. **自動クリーンアップ**：マージ後、一時チャンクディレクトリを削除。

- **利点**：高互換性、CRC32 整合性検証。
- **セキュリティ**：各セッションはアップロードユーザーの ID にバインド。

### 2. ストリームチャンキング

最大のパフォーマンスと最小のリソース使用のために設計。コアコンセプト：**「ゼロコピーパイプ」**。

- **ワークフロー**：
  1. **フロントエンドストリーミング**：フロントエンドが論理的にファイルを分割し、`Content-Range` ヘッダー付きで `Raw Binary` を `PUT` で送信。
  2. **io.Pipe ブリッジ**：最初のチャンクで、バックエンドがゼロバッファパイプ（`io.Pipe`）を作成し、すぐにパイプから読み取るストレージドライバのアップロードタスクを開始。
  3. **ゼロコピーフロー**：後続のチャンクは同じパイプに書き込まれ、データは「フロントエンドリクエスト」から「サーバーメモリ」を経て「クラウドストレージ」へ直接流れる。
  4. **自動完了**：最後のチャンク後、パイプが閉じアップロード完了。

- **利点**：
  - **ディスク使用量ゼロ**：一時チャンクなし、ディスクマージなし。
  - **最小メモリ**：パイプのバックプレッシャーにより、メモリは KB レベルを維持。
  - **高性能**：I/O ボトルネックのないダイレクトストリーミング。
- **注意**：サーバーは同期パイプとして機能。クラウド速度が遅い場合、TCP を介してクライアントにバックプレッシャーがかかる。

---

## ルート変更

| ルート | メソッド | 機能 | 認証 |
|--------|----------|------|------|
| `/api/fs/put/chunk/init` | POST | チャンクセッション初期化 | `FsUp` ミドルウェア |
| `/api/fs/put/chunk` | PUT | チャンクをアップロード | `FsUp` + レート制限 |
| `/api/fs/put/chunk/merge` | POST | チャンクを結合してアップロード | `FsUp` + レート制限 |
| `/api/fs/put` | PUT | ストリームアップロード（Content-Range 対応） | `FsUp` + レート制限 |

---

## デプロイガイド

### 直接置換（OpenList データと完全互換）

1. OpenList サービスを停止
2. 元の `openlist` バイナリをバックアップ
3. コンパイル済みの `openlist` バイナリに置換
4. サービスを開始

```bash
systemctl stop openlist
cp openlist /opt/openlist/openlist
chmod +x /opt/openlist/openlist
systemctl start openlist
```

### ソースからビルド

```bash
git clone https://github.com/zmabin/OpenList_Chunk.git
cd OpenList_Chunk

# フロントエンドアセットをダウンロード
bash build.sh dev web

# ビルド（Linux）
export CGO_ENABLED=0
go build -o openlist -tags=jsoniter -ldflags="-s -w" .

# ビルド（Windows）
set CGO_ENABLED=0
go build -o openlist.exe -tags=jsoniter -ldflags="-s -w" .
```

### Docker

```bash
# ghcr.io
docker pull ghcr.io/zmabin/openlist-chunk:latest
docker run -d -p 5244:5244 -v ./data:/opt/openlist/data ghcr.io/zmabin/openlist-chunk:latest

# Docker Hub
docker pull zmabin/openlist-chunk:latest
docker run -d -p 5244:5244 -v ./data:/opt/openlist/data zmabin/openlist-chunk:latest
```

### Nginx プロキシ設定

完全な設定は `conf.d/openlist.conf` を参照。主要設定：

```nginx
client_max_body_size 102400m;       # 最大アップロード 100GB
proxy_request_buffering off;         # リクエストバッファリング無効（ストリーミングに必須）
proxy_send_timeout 86400s;           # 24時間タイムアウト
```

---

## 設定

| キー | タイプ | デフォルト | 説明 |
|------|--------|-----------|------|
| `chunked_upload_mode` | Select | `auto` | チャンクモード：`auto` / `disabled` |
| `chunked_upload_chunk_size` | Number | `95` | チャンク閾値（MB）、これを超えるファイルは自動的にチャンク化 |

---

## ロードマップ

- [x] **ストリームチャンクアップロード**：Content-Range ベースのゼロコピーパイプチャンキング
- [x] **フォームチャンクアップロード**：セッションベースのマルチパートチャンク + ストリーミングマージ
- [x] **ストレージ定期再ログイン**：ストレージごとのパスワード再認証によるキープアライブ

---

## 謝辞

本プロジェクトは以下の優秀なプロジェクトの成果を参考にしています：

- [LusiyAvA/openlist-chunk](https://github.com/LusiyAvA/openlist-chunk) のチャンクアップロードに関する中核的なアイデアと実装参照に感謝
- [OpenListTeam/OpenList](https://github.com/OpenListTeam/OpenList) の安定した信頼性の高い基盤フレームワークに感謝

---

## サポート

このプロジェクトが役に立ったら、Star をお願いします！

バグを見つけたり提案がある場合は、[Issue](https://github.com/zmabin/OpenList_Chunk/issues) をお気軽にどうぞ。
