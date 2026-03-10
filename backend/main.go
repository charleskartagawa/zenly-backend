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
	// Railway akan memberikan variabel lingkungan DATABASE_URL secara otomatis
	// Jika tidak ada, kita pakai manual URL yang kamu berikan tadi
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Paste URL Railway kamu di sini sebagai fallback jika dijalankan lokal
		connStr = "postgresql://postgres:nCnQZcKFeUuJrKurnuVenlmOvnOlwsCJ@postgres.railway.internal:5432/railway"
		fmt.Println("🏠 Menjalankan Database LOKAL/Internal...")
	} else {
		fmt.Println("☁️ Menjalankan Database CLOUD (Railway Environment)...")
	}

	var err error
	// Driver 'postgres' bisa langsung menerima format postgresql://...
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal buka koneksi DB:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Gagal konek ke DB (Ping Error). Pastikan Database di Railway sudah 'Active':", err)
	}
	fmt.Println("✅ Database PostgreSQL Terhubung!")
}

// Endpoint untuk mengambil riwayat lokasi (GET /history)
func handleHistory(w http.ResponseWriter, r *http.Request) {
	// Biar bisa diakses dari mana saja (CORS)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Query mengambil 20 lokasi terbaru
	rows, err := db.Query(`
		SELECT user_id, ST_Y(geom::geometry), ST_X(geom::geometry), recorded_at 
		FROM user_locations 
		ORDER BY recorded_at DESC 
		LIMIT 20
	`)
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
		history = append(history, map[string]interface{}{
			"user": u, 
			"lat": lat, 
			"lon": lon, 
			"time": t,
		})
	}
	
	if history == nil {
		history = []map[string]interface{}{}
	}
	
	json.NewEncoder(w).Encode(history)
}

// Endpoint WebSocket untuk menerima lokasi (WS /ws)
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

		// Validasi data latitude & longitude
		lat, latOk := msg["latitude"].(float64)
		lon, lonOk := msg["longitude"].(float64)

		if latOk && lonOk {
			// Simpan ke PostGIS
			_, err = db.Exec(`
				INSERT INTO user_locations (user_id, geom) 
				VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326))
			`, userID, lon, lat)
			
			if err != nil {
				log.Println("⚠️ Gagal Simpan ke Database:", err)
			} else {
				fmt.Printf("📍 Lokasi [%s] Diterima: Lat %.5f, Lon %.5f\n", userID, lat, lon)
			}
		}
	}
}

func main() {
	initDB()

	// Port dinamis dari Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/history", handleHistory)

	fmt.Printf("🚀 Server Zenly Online di Port :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Server Error:", err)
	}
}
