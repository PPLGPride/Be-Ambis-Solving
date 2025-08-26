# Dockerfile
FROM golang:1.24.5-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/server /app/server
ENV PORT=8000
EXPOSE 8000
ENTRYPOINT ["/app/server"]
