# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache make
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOFLAGS="-buildvcs=false" go mod download
COPY . .
RUN make build

# Runtime stage
FROM gcr.io/distroless/base-debian12
LABEL org.opencontainers.image.source="https://github.com/denchenko/messageflow" \
      org.opencontainers.image.licenses="MIT"

USER nonroot:nonroot
COPY --from=builder /app/bin/messageflow /usr/bin/messageflow
ENTRYPOINT ["/usr/bin/messageflow"]
CMD ["--help"] 
