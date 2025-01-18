# Build the app
FROM golang:1.22.3 as builder
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o ./main

# Only copy the build
FROM scratch
WORKDIR /app
COPY --from=builder /build/main ./main
ENTRYPOINT ["./main"]