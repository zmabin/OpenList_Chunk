<div align="center">
  <img src="https://raw.githubusercontent.com/OpenListTeam/Logo/main/logo.svg" width="128" height="128" alt="logo" />

  <h1>OpenList_Chunk</h1>

  <p><em>Enhanced fork of OpenList — Bypass CDN upload size limits with chunked upload support</em></p>

  <a href="https://github.com/zmabin/OpenList_Chunk/actions?query=workflow%3ABuild"><img src="https://img.shields.io/github/actions/workflow/status/zmabin/OpenList_Chunk/build.yml?branch=main" alt="Build status" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/releases"><img src="https://img.shields.io/github/v/release/zmabin/OpenList_Chunk" alt="latest version" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/blob/main/LICENSE"><img src="https://img.shields.io/github/license/zmabin/OpenList_Chunk" alt="License" /></a>
</div>

---

**English** | [中文](./README_cn.md) | [日本語](./README_ja.md) | [Nederlands](./README_nl.md)

---

## Overview

**OpenList_Chunk** is an enhanced fork of [OpenList](https://github.com/OpenListTeam/OpenList) that refactors the upload logic while keeping all original data structures intact.

**Core goal: Bypass upload size limits imposed by CDN reverse proxies (e.g., Cloudflare Free plan limits single requests to 100MB).**

**Drop-in replacement — no hassle.**

---

## Core Modifications: Bypassing CDN Limits

This project implements two distinct mechanisms to bypass CDN upload body limits.

### 1. Form Chunked Upload

A traditional high-compatibility chunking mechanism based on **"session management + disk cache + streaming merge"**.

- **Workflow**:
  1. **Init session**: Frontend calls `POST /api/fs/put/chunk/init`, backend generates a unique `upload_id` and creates a session file.
  2. **Upload chunks**: Each chunk is sent as `multipart/form-data` to `PUT /api/fs/put/chunk` with `upload_id` and `index`.
  3. **CRC32 verification**: Server computes CRC32 for each chunk and compares against the `X-Chunk-CRC32` header from the client.
  4. **Virtual merge**: After all chunks are uploaded, frontend calls `POST /api/fs/put/chunk/merge`. Backend uses `io.MultiReader` to read all temp files sequentially with zero disk copy, streaming directly to the storage backend.
  5. **Auto cleanup**: Temp chunk directory is deleted after merge.

- **Advantages**: High compatibility, CRC32 integrity verification.
- **Security**: Each session is bound to the uploading user's identity.

### 2. Stream Chunking

Designed for maximum performance and minimal resource usage. Core concept: **"zero-copy pipe"**.

- **Workflow**:
  1. **Frontend streaming**: Frontend logically splits the file and sends `Raw Binary` via `PUT` with `Content-Range` headers.
  2. **io.Pipe bridge**: On the first chunk, the backend creates a zero-buffer pipe (`io.Pipe`) and immediately starts the storage driver upload task reading from the pipe.
  3. **Zero-copy flow**: Subsequent chunks write to the same pipe. Data flows directly from "frontend request" through "server memory" to "cloud storage".
  4. **Auto complete**: After the last chunk, the pipe is closed and upload finishes.

- **Advantages**:
  - **Zero disk usage**: No temp chunks, no disk merge.
  - **Minimal memory**: Through pipe back-pressure, memory stays at KB-level.
  - **High performance**: Direct streaming with no I/O bottleneck.
- **Note**: Server acts as a sync pipe; slow cloud speeds will back-pressure the client via TCP.

---

## Route Changes

| Route | Method | Function | Auth |
|-------|--------|----------|------|
| `/api/fs/put/chunk/init` | POST | Initialize chunk session | `FsUp` middleware |
| `/api/fs/put/chunk` | PUT | Upload a single chunk | `FsUp` + rate limit |
| `/api/fs/put/chunk/merge` | POST | Merge chunks and upload | `FsUp` + rate limit |
| `/api/fs/put` | PUT | Stream upload (supports Content-Range) | `FsUp` + rate limit |

---

## Deployment Guide

### Direct Replacement (Fully Compatible with OpenList Data)

1. Stop your OpenList service
2. Backup the original `openlist` binary
3. Replace with the compiled `openlist` binary
4. Start the service

```bash
systemctl stop openlist
cp openlist /opt/openlist/openlist
chmod +x /opt/openlist/openlist
systemctl start openlist
```

### Build from Source

```bash
git clone https://github.com/zmabin/OpenList_Chunk.git
cd OpenList_Chunk

# Download frontend assets
bash build.sh dev web

# Build (Linux)
export CGO_ENABLED=0
go build -o openlist -tags=jsoniter -ldflags="-s -w" .

# Build (Windows)
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

### Nginx Proxy Config

Refer to `conf.d/openlist.conf` for the full config. Key settings:

```nginx
client_max_body_size 102400m;       # 100GB max upload
proxy_request_buffering off;         # Disable request buffering (required for streaming)
proxy_send_timeout 86400s;           # 24-hour timeout
```

---

## Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `chunked_upload_mode` | Select | `auto` | Chunk mode: `auto` / `disabled` |
| `chunked_upload_chunk_size` | Number | `95` | Chunk threshold (MB), files larger than this will be auto-chunked |

---

## Roadmap

- [x] **Form Chunked Upload**: Session-based multipart chunk + streaming merge
- [x] **Stream Chunking**: Content-Range based zero-copy pipe chunking
- [x] **Scheduled Storage Re-login**: Per-storage keep-alive via forced password re-authentication

---

## Acknowledgments

This project references and builds upon the work of the following excellent projects:

- Thanks to [LusiyAvA/openlist-chunk](https://github.com/LusiyAvA/openlist-chunk) for the core ideas and implementation reference for chunked upload
- Thanks to [OpenListTeam/OpenList](https://github.com/OpenListTeam/OpenList) for providing a stable and reliable foundation framework

---

## Support

If this project helps you, please consider giving it a Star!

Found a bug or have a suggestion? Feel free to open an [Issue](https://github.com/zmabin/OpenList_Chunk/issues).
