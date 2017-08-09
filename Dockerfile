FROM golang:alpine as builder

COPY . .
RUN go build -ldflags "-s -w" -o /gosimpleproxy;

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /gosimpleproxy /

ENTRYPOINT ["/gosimpleproxy"]
