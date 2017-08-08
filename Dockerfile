FROM golang:alpine as builder

COPY . .
RUN go build -o /gosimpleproxy;

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /gosimpleproxy /

ENTRYPOINT ["/gosimpleproxy"]
