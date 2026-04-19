FROM golang:1.26-alpine AS builder

ARG SERVICE=gateway

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOFLAGS="-trimpath" \
    go build -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /src/migrations /app/migrations

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/app"]
