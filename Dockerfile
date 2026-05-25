# OpenList_Chunk Dockerfile
# Build (local): docker build -t username/openlist-chunk .
# Build (CI):    docker build --build-arg SKIP_FRONTEND=true -t username/openlist-chunk .
# Usage: docker run -d -p 5244:5244 -v /path/to/data:/opt/openlist/data username/openlist-chunk

ARG SKIP_FRONTEND=false

# ---- Frontend build stage (skipped when SKIP_FRONTEND=true) ----
FROM node:22-alpine AS frontend-builder
ARG SKIP_FRONTEND
WORKDIR /build

RUN if [ "$SKIP_FRONTEND" != "true" ]; then \
      apk add --no-cache git jq curl && \
      git clone --depth 1 https://github.com/OpenListTeam/OpenList-Frontend.git . && \
      corepack enable pnpm; \
    else \
      mkdir -p dist && touch dist/.skip; \
    fi

# Copy chunk upload files over upstream (harmless when skipping)
COPY frontend/src/pages/home/uploads/ ./src/pages/home/uploads/

# Download i18n, install deps, and build (all in one layer with pnpm cache)
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    if [ "$SKIP_FRONTEND" != "true" ]; then \
      FRONTEND_RELEASE=$(curl -fsSL \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/OpenListTeam/OpenList-Frontend/releases/latest") && \
      I18N_URL=$(echo "$FRONTEND_RELEASE" | jq -r '.assets[].browser_download_url' | grep "i18n.tar.gz" | head -1) && \
      if [ -n "$I18N_URL" ]; then \
        curl -fsSL "$I18N_URL" -o i18n.tar.gz && \
        tar -xzf i18n.tar.gz -C src/lang && \
        rm -f i18n.tar.gz && \
        echo "i18n translations downloaded"; \
      else \
        echo "Warning: i18n.tar.gz not found, building with English only"; \
      fi && \
      pnpm install && pnpm add crc-32 && pnpm build; \
    fi

# ---- Go build stage ----
FROM golang:1.24-alpine AS builder
ARG SKIP_FRONTEND

RUN apk add --no-cache git bash curl jq
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

# Copy built frontend from builder stage to temp location
COPY --from=frontend-builder /build/dist /tmp/frontend-dist/

# If frontend was built (has index.html), use it; otherwise keep pre-built from context
RUN if [ -f /tmp/frontend-dist/index.html ]; then \
      rm -rf public/dist && cp -r /tmp/frontend-dist/* public/dist/; \
    fi && \
    rm -rf /tmp/frontend-dist

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
