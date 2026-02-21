ARG GO_VERSION=1.25.6

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-bookworm-slim AS build

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

COPY . .

ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/anek-bot \
    ./cmd/bot

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/goose \
    ./cmd/migrator

################################################################################

FROM gcr.io/distroless/static:nonroot AS final

WORKDIR /app

COPY --from=build /app/anek-bot ./
COPY --from=build /app/goose ./
COPY migrations/ ./migrations/

USER nonroot:nonroot

ENTRYPOINT ["/app/anek-bot"]
