FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /incidencias-tui .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /incidencias-tui /usr/local/bin/incidencias-tui
ENTRYPOINT ["incidencias-tui"]
