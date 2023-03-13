# Build stage
FROM golang:1.18.3-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o gotybench main.go


# Run stage
FROM alpine:3.16
RUN apk add --no-cache bash
WORKDIR /app
COPY "args.sh" .
RUN chmod 755 /app/args.sh
COPY --from=builder /app/gotybench /bin
ENTRYPOINT ["/app/args.sh"]