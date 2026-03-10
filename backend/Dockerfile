FROM golang:1.22-alpine

# Install git buat jaga-jaga kalau ada library privat
RUN apk add --no-cache git

WORKDIR /app

# Copy SEMUA isi folder backend sekaligus ke dalam container
COPY . .

# Download dependencies berdasarkan file yang baru saja di-copy
RUN go mod download

# Build binary
RUN go build -o main .

# Port standar
EXPOSE 8080

# Jalankan
CMD ["./main"]
