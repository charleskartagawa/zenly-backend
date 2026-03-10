# Pakai versi golang terbaru biar sinkron sama laptop kamu
FROM golang:latest

RUN apt-get update && apt-get install -y git

WORKDIR /app

# Copy semua file
COPY . .

# Paksa update toolchain kalau versinya beda
RUN go mod download
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
