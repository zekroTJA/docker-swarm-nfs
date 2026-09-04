# syntax=docker/dockerfile:1

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/guestbook ./cmd/guestbook

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/guestbook /app/guestbook
ENV GUESTBOOK_ADDRESS=:8080
ENV GUESTBOOK_STORAGE_DIR=/data
VOLUME ["/data"]
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/guestbook"]
