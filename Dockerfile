# Stage 1: Build the frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY server/ ./server/
RUN rm -rf server/public && mkdir -p server/public
COPY --from=frontend-builder /web/dist/ server/public/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tavily-proxy ./server

# Stage 3: Final image
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /out/tavily-proxy ./tavily-proxy

VOLUME /app/data
ENV DATABASE_PATH=/app/data/proxy.db

EXPOSE 8080

CMD ["./tavily-proxy"]
