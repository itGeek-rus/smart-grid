FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/processor ./cmd/processor

FROM gcr.io/distroless/static-debain12:nonroot

WORKDIR /app
COPY --from=builder /out/processor /app/processor

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/processor"]
