package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var db *sql.DB
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func initDB() {
	var err error
	connStr := "host=localhost port=5432 user=user_zenly password=password_zenly dbname=zenly_db sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil || db.Ping() != nil {
		log.Fatal("❌ Cek Docker Postgres kamu!")
	}
	fmt.Println("✅ Database Connected!")
}

// Fitur HISTORY: Menampilkan 20 lokasi terakhir semua user
func handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	rows, _ := db.Query("SELECT user_id, ST_Y(geom::geometry), ST_X(geom::geometry), recorded_at FROM user_locations ORDER BY recorded_at DESC LIMIT 20")
	var history []map[string]interface{}
	for rows.Next() {
		var u string; var lat, lon float64; var t string
		rows.Scan(&u, &lat, &lon, &t)
		history = append(history, map[string]interface{}{"user": u, "lat": lat, "lon": lon, "time": t})
	}
	json.NewEncoder(w).Encode(history)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	userID := r.URL.Query().Get("user_id")
	fmt.Printf("🌐 [%s] Online\n", userID)
	for {
		var msg map[string]interface{}
		if err := ws.ReadJSON(&msg); err != nil { break }
		db.Exec("INSERT INTO user_locations (user_id, geom) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326))", 
			msg["user_id"], msg["longitude"], msg["latitude"])
		fmt.Printf("📍 [%s]: %.4f, %.4f\n", msg["user_id"], msg["latitude"], msg["longitude"])
	}
}

func main() {
	initDB()
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/history", handleHistory) // Endpoint baru
	fmt.Println("🚀 Server: Port 9000"); log.Fatal(http.ListenAndServe(":9000", nil))
}