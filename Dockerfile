FROM golang:1.18
WORKDIR /go/src/app
ENV TZ Asia/Tokyo
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN apt-get update && \
  apt-get install -y --no-install-recommends tesseract-ocr libtesseract-dev && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* && \
  go build -o ./server ./cmd/server/main.go && \
  go build -o ./scrape ./cmd/scrape/main.go
CMD ["./server"]
