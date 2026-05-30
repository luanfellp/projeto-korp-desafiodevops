# ─────────────────────────────────────────────────────────────
# Stage 1 — Build
# ─────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Dependências necessárias para download de módulos via HTTPS
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copia apenas go.mod primeiro para aproveitar cache de layers
COPY go.mod ./

# Baixa dependências e resolve go.sum
RUN go mod download || true

# Copia o restante do código-fonte
COPY . .

# go mod tidy garante go.sum consistente; build gera binário estático
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags="-w -s" \
    -o http-server-projeto-korp .

# ─────────────────────────────────────────────────────────────
# Stage 2 — Runtime (imagem mínima)
# ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser  -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/http-server-projeto-korp .
RUN chown appuser:appgroup http-server-projeto-korp

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./http-server-projeto-korp"]
