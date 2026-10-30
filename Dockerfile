# syntax=docker/dockerfile:1

# Multi-stage build for the GoDo API (./cmd/api).
#
#   web      builds the frontend (web/dist)
#   build    compiles a static binary with the frontend embedded
#   dev      Go toolchain + air, for live reload (used by compose.dev.yaml)
#   runtime  small Alpine image that runs the binary (the default target)

ARG NODE_IMAGE=node:24.22.1-alpine3.24
ARG GO_IMAGE=golang:1.27.3-alpine3.24
ARG RUNTIME_IMAGE=alpine:3.24.3

# ------------------------------------------------------------
# web
# ------------------------------------------------------------
FROM ${NODE_IMAGE} AS web
WORKDIR /src/web

# Install packages in their own layer so code changes don't reinstall them.
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY web/ ./
RUN npm run build

# ------------------------------------------------------------
# build
# ------------------------------------------------------------
FROM ${GO_IMAGE} AS build
WORKDIR /src

# Download modules in their own layer so code changes don't re-download them.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
COPY --from=web /src/web/dist ./web/dist
# -buildvcs=false: .git is not copied into the build context.
# -tags embedweb: embed web/dist so the API serves the frontend.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags='-s -w' \
    -tags embedweb -o /out/api ./cmd/api

# ------------------------------------------------------------
# dev (live reload with air; source is bind-mounted at /app)
# ------------------------------------------------------------
FROM ${GO_IMAGE} AS dev
RUN go install github.com/air-verse/air@v1.67.4
# The bind-mounted checkout may have a .git the container can't read (for
# example a git worktree), so don't try to stamp VCS info.
ENV GOFLAGS=-buildvcs=false
WORKDIR /app
EXPOSE 4000
# Polling is more reliable than fsnotify on bind mounts (macOS/Windows hosts).
# Arguments after "--" (the compose `command`) are passed to the API binary.
ENTRYPOINT ["air", "-build.poll=true", "--"]

# ------------------------------------------------------------
# runtime
# ------------------------------------------------------------
FROM ${RUNTIME_IMAGE} AS runtime
RUN addgroup -S godo && adduser -S -G godo godo
COPY --from=build /out/api /usr/local/bin/godo-api
USER godo
EXPOSE 4000
ENTRYPOINT ["/usr/local/bin/godo-api"]
