FROM golang:1.18-alpine AS builder
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o app .

FROM alpine:3.16
COPY --from=builder /go/src/app/app /
EXPOSE 8088
CMD ["/app"]
