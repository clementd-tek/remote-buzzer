# syntax=docker/dockerfile:1

# --- build stage -----------------------------------------------------------
# go.mod currently pins `go 1.26.5`. If that toolchain image isn't
# available yet in your registry, either bump this tag to whatever's
# current or lower the `go` directive in go.mod to match.
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

# --- runtime stage -----------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/server /usr/local/bin/server

USER app

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
