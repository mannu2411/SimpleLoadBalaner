FROM golang:1.13 AS builder
LABEL authors="manubhav"
WORKDIR  /app
COPY main.go go.mod ./
COPY handler.go loadBalancer.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o lb .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/lb .
ENTRYPOINT ["/root/lb"]