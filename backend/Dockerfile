FROM golang:1.22-alpine

WORKDIR /app

# Gunakan wildcard biar kalau go.sum belum ada, build-nya nggak error
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy semua file dari folder tempat Dockerfile ini berada
COPY . .

# Build
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
