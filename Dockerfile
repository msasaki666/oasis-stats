FROM golang:1.18-alpine AS builder
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o app .

FROM alpine:3.16
ENV TZ Asia/Tokyo
COPY --from=builder /go/src/app/app /
EXPOSE 8088
RUN apk --update add tzdata && \
    rm -rf /var/cache/apk/*
CMD ["/app"]
