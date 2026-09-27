FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-w -s" -o /app/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /app/server /server

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]
