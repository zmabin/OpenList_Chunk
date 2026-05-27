<div align="center">
  <img src="https://raw.githubusercontent.com/OpenListTeam/Logo/main/logo.svg" width="128" height="128" alt="logo" />

  <h1>OpenList_Chunk</h1>

  <p><em>OpenList 增强分支 — 通过分块上传突破 CDN 上传大小限制</em></p>

  <a href="https://github.com/zmabin/OpenList_Chunk/actions?query=workflow%3ABuild"><img src="https://img.shields.io/github/actions/workflow/status/zmabin/OpenList_Chunk/build.yml?branch=main" alt="Build status" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/releases"><img src="https://img.shields.io/github/v/release/zmabin/OpenList_Chunk" alt="latest version" /></a>
  <a href="https://github.com/zmabin/OpenList_Chunk/blob/main/LICENSE"><img src="https://img.shields.io/github/license/zmabin/OpenList_Chunk" alt="License" /></a>
</div>

---

[English](./README.md) | **中文** | [日本語](./README_ja.md) | [Nederlands](./README_nl.md)

---

## 概述

**OpenList_Chunk** 是 [OpenList](https://github.com/OpenListTeam/OpenList) 的增强分支，重构了上传逻辑，同时保持所有原始数据结构不变。

**核心目标：绕过 CDN 反向代理对上传大小的限制（例如 Cloudflare 免费计划限制单次请求 100MB）。**

**直接替换，无需额外配置。**

---

## 核心修改：绕过 CDN 限制

本项目实现了两种不同的机制来绕过 CDN 上传请求体限制。

### 1. 表单分块上传

基于 **"会话管理 + 磁盘缓存 + 流式合并"** 的传统高兼容性分块机制。

- **工作流程**：
  1. **初始化会话**：前端调用 `POST /api/fs/put/chunk/init`，后端生成唯一的 `upload_id` 并创建会话文件。
  2. **上传分块**：每个分块以 `multipart/form-data` 格式发送到 `PUT /api/fs/put/chunk`，携带 `upload_id` 和 `index`。
  3. **CRC32 校验**：服务器计算每个分块的 CRC32，并与客户端 `X-Chunk-CRC32` 请求头对比。
  4. **虚拟合并**：所有分块上传完成后，前端调用 `POST /api/fs/put/chunk/merge`。后端使用 `io.MultiReader` 顺序读取所有临时文件，零磁盘拷贝，直接流式传输到存储后端。
  5. **自动清理**：合并后删除临时分块目录。

- **优势**：高兼容性，CRC32 完整性校验。
- **安全性**：每个会话绑定上传用户的身份。

### 2. 流式分块

专为最大性能和最小资源占用设计。核心理念：**"零拷贝管道"**。

- **工作流程**：
  1. **前端流式传输**：前端逻辑上分割文件，通过 `PUT` 以 `Raw Binary` 格式发送，携带 `Content-Range` 请求头。
  2. **io.Pipe 桥接**：第一个分块到达时，后端创建零缓冲管道（`io.Pipe`），并立即启动从管道读取的存储驱动上传任务。
  3. **零拷贝数据流**：后续分块写入同一管道。数据直接从"前端请求"经"服务器内存"流向"云存储"。
  4. **自动完成**：最后一个分块后，管道关闭，上传完成。

- **优势**：
  - **零磁盘占用**：无临时分块，无磁盘合并。
  - **极低内存**：通过管道背压，内存保持在 KB 级别。
  - **高性能**：直接流式传输，无 I/O 瓶颈。
- **注意**：服务器充当同步管道；云存储速度慢时会通过 TCP 反压客户端。

### 3. 定时存储重登录

通过强制密码重认证实现的按存储保活机制。

- **实现方式**：
  1. 通过管理页面为每个存储设置登录间隔（分钟）。
  2. 每次触发时，系统丢弃旧驱动实例并剥离所有缓存令牌。
  3. 创建全新驱动实例 — 找不到缓存令牌，执行用户名/密码登录。
  4. 如果密码登录失败，系统自动回退到使用备份令牌进行令牌刷新。

- **如何启用**：
  1. 打开 `http://<host>:5244/@manage/login-schedule`（也可从管理后台侧边栏"任务"下方进入）。
  2. 为每个存储设置登录间隔（分钟）。`0` = 禁用。
  3. 点击保存。

---

## 路由变更

| 路由 | 方法 | 功能 | 认证 |
|------|------|------|------|
| `/api/fs/put/chunk/init` | POST | 初始化分块会话 | `FsUp` 中间件 |
| `/api/fs/put/chunk` | PUT | 上传单个分块 | `FsUp` + 速率限制 |
| `/api/fs/put/chunk/merge` | POST | 合并分块并上传 | `FsUp` + 速率限制 |
| `/api/fs/put` | PUT | 流式上传（支持 Content-Range） | `FsUp` + 速率限制 |
| `/@manage/login-schedule` | GET | 存储定时登录管理页面 | 客户端令牌 |

---

## 部署指南

### 直接替换（完全兼容 OpenList 数据）

1. 停止 OpenList 服务
2. 备份原始 `openlist` 二进制文件
3. 替换为编译后的 `openlist` 二进制文件
4. 启动服务

```bash
systemctl stop openlist
cp openlist /opt/openlist/openlist
chmod +x /opt/openlist/openlist
systemctl start openlist
```

### 从源码构建

```bash
git clone https://github.com/zmabin/OpenList_Chunk.git
cd OpenList_Chunk

# 下载前端资源
bash build.sh dev web

# 构建（Linux）
export CGO_ENABLED=0
go build -o openlist -tags=jsoniter -ldflags="-s -w" .

# 构建（Windows）
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

### Nginx 代理配置

完整配置参考 `conf.d/openlist.conf`。关键设置：

```nginx
client_max_body_size 102400m;       # 最大上传 100GB
proxy_request_buffering off;         # 禁用请求缓冲（流式传输必需）
proxy_send_timeout 86400s;           # 24 小时超时
```

---

## 设置项

| 键名 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `chunked_upload_mode` | 选择 | `auto` | 分块模式：`auto` / `disabled` |
| `chunked_upload_chunk_size` | 数字 | `95` | 分块阈值（MB），超过此大小的文件将自动分块 |

---

## 路线图

- [x] **stream切片上传**：基于 Content-Range 的零拷贝管道分块
- [x] **form切片上传**：基于会话的多部分分块 + 流式合并
- [x] **定时存储重登录**：为每个存储配置定时强制密码重认证保活

---

## 致谢

本项目参考并借鉴了以下优秀项目的工作：

- 感谢 [LusiyAvA/openlist-chunk](https://github.com/LusiyAvA/openlist-chunk) 提供分块上传的核心思路和实现参考
- 感谢 [OpenListTeam/OpenList](https://github.com/OpenListTeam/OpenList) 提供稳定可靠的基础框架

---

## 支持

如果这个项目对你有帮助，请给一个 Star！

发现 Bug 或有建议？欢迎提交 [Issue](https://github.com/zmabin/OpenList_Chunk/issues)。
