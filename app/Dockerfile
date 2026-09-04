# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/guestbook ./cmd/guestbook

FROM alpine
WORKDIR /app
COPY --from=build /out/guestbook /app/guestbook
ENV GUESTBOOK_ADDRESS=:8080
ENV GUESTBOOK_STORAGE_DIR=/data
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["/app/guestbook"]
