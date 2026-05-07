# OpenList_Chunk Dockerfile
# Build: docker build -t username/openlist-chunk .
# Usage: docker run -d -p 5244:5244 -v /path/to/data:/opt/openlist/data username/openlist-chunk

# ---- Frontend build stage ----
FROM node:22-alpine AS frontend-builder

RUN apk add --no-cache git
WORKDIR /build

# Clone upstream frontend source
RUN git clone --depth 1 https://github.com/OpenListTeam/OpenList-Frontend.git .

# Enable pnpm
RUN corepack enable pnpm

# Copy chunk-enabled upload files over upstream
COPY frontend/src/pages/home/uploads/ ./src/pages/home/uploads/

# Download i18n translations from upstream release (only English is in repo)
RUN FRONTEND_RELEASE=$(curl -fsSL \
      -H "Accept: application/vnd.github.v3+json" \
      "https://api.github.com/repos/OpenListTeam/OpenList-Frontend/releases/latest") && \
    I18N_URL=$(echo "$FRONTEND_RELEASE" | grep -oP '"browser_download_url":\s*"\K[^"]*' | grep "i18n.tar.gz") && \
    if [ -n "$I18N_URL" ]; then \
      curl -fsSL "$I18N_URL" -o i18n.tar.gz && \
      tar -xzf i18n.tar.gz -C src/lang && \
      rm -f i18n.tar.gz && \
      echo "i18n translations downloaded"; \
    else \
      echo "Warning: i18n.tar.gz not found, building with English only"; \
    fi

# Install dependencies, add crc-32 (needed by chunk upload), and build
RUN pnpm install && pnpm add crc-32 && pnpm build

# ---- Go build stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git bash curl jq
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
# Use the chunk-enabled frontend built from source
COPY --from=frontend-builder /build/dist ./public/dist/

RUN go mod tidy && CGO_ENABLED=0 go build \
    -ldflags="-w -s" \
    -tags=jsoniter \
    -o /build/openlist .

# ---- Runtime stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1001 openlist && \
    adduser -D -u 1001 -G openlist openlist && \
    mkdir -p /opt/openlist/data

WORKDIR /opt/openlist/

COPY --from=builder --chmod=755 --chown=1001:1001 /build/openlist ./
COPY --from=builder --chmod=755 --chown=1001:1001 /build/public ./public/
COPY --chmod=755 entrypoint.sh /entrypoint.sh

USER openlist

VOLUME /opt/openlist/data/
EXPOSE 5244

ENV UMASK=022
CMD ["/entrypoint.sh"]
