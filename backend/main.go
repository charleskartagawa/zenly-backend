package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var db *sql.DB
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func initDB() {
	// Railway/Render otomatis kasih DATABASE_URL
	connStr := os.Getenv("DATABASE_URL")
	
	// Kalau di laptop (lokal), pakai settingan lama kamu
	if connStr == "" {
		connStr = "host=localhost port=5432 user=user_zenly password=password_zenly dbname=zenly_db sslmode=disable"
		fmt.Println("🏠 Menjalankan Database LOKAL...")
	} else {
		fmt.Println("☁️ Menjalankan Database CLOUD...")
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal buka koneksi DB:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Gagal konek ke DB (Ping Error):", err)
	}
	fmt.Println("✅ Database PostgreSQL Terhubung!")
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT user_id, ST_Y(geom::geometry), ST_X(geom::geometry), recorded_at FROM user_locations ORDER BY recorded_at DESC LIMIT 20")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var u string
		var lat, lon float64
		var t string
		rows.Scan(&u, &lat, &lon, &t)
		history = append(history, map[string]interface{}{"user": u, "lat": lat, "lon": lon, "time": t})
	}
	json.NewEncoder(w).Encode(history)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("⚠️ Upgrade Error:", err)
		return
	}
	defer ws.Close()

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}
	fmt.Printf("🌐 User [%s] Telah Terhubung\n", userID)

	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("❌ User [%s] Putus Koneksi\n", userID)
			break
		}

		// Masukkan data ke PostGIS
		_, err = db.Exec("INSERT INTO user_locations (user_id, geom) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326))",
			userID, msg["longitude"], msg["latitude"])
		
		if err != nil {
			log.Println("⚠️ Gagal Simpan Lokasi:", err)
		} else {
			fmt.Printf("📍 Lokasi [%s] Diterima: Lat %.5f, Lon %.5f\n", userID, msg["latitude"], msg["longitude"])
		}
	}
}

func main() {
	initDB()

	// Railway/Cloud butuh port dinamis via variabel PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default kalau di laptop
	}

	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/history", handleHistory)

	fmt.Printf("🚀 Server Zenly jalan di Port :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Server Error:", err)
	}
}
