######################## build stage ########################
FROM golang:1.22-alpine AS builder
WORKDIR /src

# ── Go modules first (better cache) ────────────────────────
COPY go.mod go.sum ./
RUN go mod download

# ── Copy source and compile static binary ──────────────────
COPY . .

# The mock-server code listens on 9090 (constant in main.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o /bin/mock-server ./main.go

####################### runtime stage #######################
FROM gcr.io/distroless/static-debian11

COPY --from=builder /bin/mock-server /mock-server

EXPOSE 9090
ENTRYPOINT ["/mock-server"]
