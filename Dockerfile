FROM golang

RUN apt update && apt install -y tesseract-ocr-all

COPY . /app
WORKDIR /app
RUN go build bin/server.go

ENTRYPOINT ["./server"]