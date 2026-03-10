# Gunakan image Go resmi versi terbaru
FROM golang:1.22-alpine

# Set folder kerja di dalam container
WORKDIR /app

# Copy file go.mod dan go.sum duluan (biar build lebih cepat)
COPY go.mod go.sum ./
RUN go mod download

# Copy semua kode backend kamu
COPY . .

# Build aplikasinya jadi file binary bernama 'main'
RUN go build -o main .

# Ekspos port (nanti Railway bakal otomatis atur ini)
EXPOSE 8080

# Jalankan aplikasinya
CMD ["./main"]