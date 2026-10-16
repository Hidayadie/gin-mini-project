package latihanapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // driver PostgreSQL untuk database/sql
	"github.com/joho/godotenv"

)

// Todo merepresentasikan struktur data untuk satu item todo
type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	IsDone bool   `json:"is_done"`
}

// db adalah pool koneksi ke PostgreSQL (aman dipakai banyak request sekaligus)
var db *sql.DB

// initDB membuka koneksi ke database memakai environment variable DATABASE_URL
func initDB() error {

	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: Tidak dapat memuat file .env:", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("environment variable DATABASE_URL belum diatur")
	}

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	// Ping untuk memastikan database benar-benar bisa dihubungi
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// writeJSON mengirim response JSON dengan status code tertentu
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError mengirim pesan error dalam format JSON
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// GET /todos - Mengambil semua todo
func getTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), "SELECT id, title, is_done FROM todos ORDER BY id")
	if err != nil {
		log.Println("getTodos:", err)
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data")
		return
	}
	defer rows.Close()

	todos := []Todo{} // slice kosong agar hasilnya [] (bukan null)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.IsDone); err != nil {
			log.Println("getTodos scan:", err)
			writeError(w, http.StatusInternalServerError, "Gagal membaca data")
			return
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		log.Println("getTodos rows:", err)
		writeError(w, http.StatusInternalServerError, "Gagal membaca data")
		return
	}

	writeJSON(w, http.StatusOK, todos)
}

// GET /todos/{id} - Mengambil detail satu todo berdasarkan ID
func getTodoByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	var t Todo
	err = db.QueryRowContext(r.Context(),
		"SELECT id, title, is_done FROM todos WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.IsDone)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Todo tidak ditemukan")
		return
	}
	if err != nil {
		log.Println("getTodoByID:", err)
		writeError(w, http.StatusInternalServerError, "Gagal mengambil data")
		return
	}

	writeJSON(w, http.StatusOK, t)
}

// POST /todos - Menambahkan todo baru
func createTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	if newTodo.Title == "" {
		writeError(w, http.StatusBadRequest, "Title tidak boleh kosong")
		return
	}

	// ID dibuat oleh database (SERIAL); RETURNING mengembalikan ID baru
	err := db.QueryRowContext(r.Context(),
		"INSERT INTO todos (title, is_done) VALUES ($1, $2) RETURNING id",
		newTodo.Title, newTodo.IsDone).Scan(&newTodo.ID)
	if err != nil {
		log.Println("createTodo:", err)
		writeError(w, http.StatusInternalServerError, "Gagal menyimpan data")
		return
	}

	writeJSON(w, http.StatusCreated, newTodo)
}

// PUT /todos/{id} - Mengubah/memperbarui todo berdasarkan ID
func updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	var updatedData Todo
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	if updatedData.Title == "" {
		writeError(w, http.StatusBadRequest, "Title tidak boleh kosong")
		return
	}

	err = db.QueryRowContext(r.Context(),
		"UPDATE todos SET title = $1, is_done = $2 WHERE id = $3 RETURNING id",
		updatedData.Title, updatedData.IsDone, id).Scan(&updatedData.ID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Todo tidak ditemukan")
		return
	}
	if err != nil {
		log.Println("updateTodo:", err)
		writeError(w, http.StatusInternalServerError, "Gagal mengubah data")
		return
	}

	writeJSON(w, http.StatusOK, updatedData)
}

// DELETE /todos/{id} - Menghapus todo berdasarkan ID
func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	result, err := db.ExecContext(r.Context(), "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		log.Println("deleteTodo:", err)
		writeError(w, http.StatusInternalServerError, "Gagal menghapus data")
		return
	}

	// RowsAffected = jumlah baris yang terhapus; 0 berarti ID tidak ada
	if n, _ := result.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "Todo tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Todo berhasil dihapus"})
}

// StartServer menghubungkan database, mendaftarkan rute, lalu menjalankan HTTP server
func StartServer() {
	if err := initDB(); err != nil {
		fmt.Println("Gagal terhubung ke database:", err)
		return
	}
	defer db.Close()
	fmt.Println("✅ Terhubung ke database PostgreSQL")

	mux := http.NewServeMux()

	// Routing bawaan Go standard library
	mux.HandleFunc("GET /todos", getTodos)
	mux.HandleFunc("GET /todos/{id}", getTodoByID)
	mux.HandleFunc("POST /todos", createTodo)
	mux.HandleFunc("PUT /todos/{id}", updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", deleteTodo)

	port := ":8080"
	fmt.Println("🚀 Server TODO API berjalan di http://localhost" + port)
	fmt.Println("Tekan Ctrl+C untuk menghentikan server.")

	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Printf("Gagal menjalankan server: %s\n", err)
	}
}
