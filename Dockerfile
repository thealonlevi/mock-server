# ---------- build stage ----------
FROM golang:1.22-alpine AS builder
WORKDIR /src

# copy go modules first for better build-cache reuse
COPY go.mod go.sum ./
RUN go mod download

# copy the rest of the source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/echo-id ./main.go

# ---------- runtime stage ----------
FROM gcr.io/distroless/static-debian11
COPY --from=builder /bin/echo-id /echo-id

EXPOSE 9090
ENTRYPOINT ["/echo-id"]
    