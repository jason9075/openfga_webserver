# Build Stage
FROM golang:1.22 AS builder
WORKDIR /app

# 複製 go.mod 與 go.sum 並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼（包含 cmd/app.go、middleware 與 pkg 等目錄）
COPY . .

# 編譯產生可執行檔
# 請確認你的 app.go 內有依據環境變數 WEB_APP_PORT 來設定監聽 port，或自行調整
RUN CGO_ENABLED=0 go build -o web-server ./cmd/app.go

# Final Stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/
COPY --from=builder /app/.env .
COPY --from=builder /app/web-server .
COPY --from=builder /app/init ./init

# 設定預設監聽的 port，請與 docker-compose 中的 WEB_APP_PORT 一致
ARG WEB_APP_PORT=8000
EXPOSE ${WEB_APP_PORT}

CMD ["./web-server"]
