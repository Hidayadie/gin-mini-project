package latihanapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestMain menyiapkan koneksi database sebelum semua test dijalankan.
// Jika DATABASE_URL belum diatur, test dilewati.
func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Println("DATABASE_URL belum diatur, test dilewati")
		os.Exit(0)
	}
	if err := initDB(); err != nil {
		fmt.Println("Gagal terhubung ke database:", err)
		os.Exit(1)
	}
	code := m.Run()
	db.Close()
	os.Exit(code)
}

func TestTodoAPI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", getTodos)
	mux.HandleFunc("GET /todos/{id}", getTodoByID)
	mux.HandleFunc("POST /todos", createTodo)
	mux.HandleFunc("PUT /todos/{id}", updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", deleteTodo)

	// Test 1: GET /todos
	t.Run("GET /todos", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status code salah: dapat %d, mau %d", rec.Code, http.StatusOK)
		}
	})

	// Test 2: POST /todos (todo yang dibuat dipakai oleh test berikutnya)
	var createdTodo Todo
	t.Run("POST /todos", func(t *testing.T) {
		body := `{"title": "Belajar Unit Test Go", "is_done": false}`
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status code salah: dapat %d, mau %d", rec.Code, http.StatusCreated)
		}
		if err := json.NewDecoder(rec.Body).Decode(&createdTodo); err != nil {
			t.Fatalf("gagal decode response json: %v", err)
		}
		if createdTodo.ID == 0 || createdTodo.Title != "Belajar Unit Test Go" {
			t.Fatalf("data todo baru tidak sesuai: %+v", createdTodo)
		}
	})

	url := fmt.Sprintf("/todos/%d", createdTodo.ID)

	// Test 3: GET /todos/{id}
	t.Run("GET /todos/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status code salah: dapat %d, mau %d", rec.Code, http.StatusOK)
		}
	})

	// Test 4: PUT /todos/{id}
	t.Run("PUT /todos/{id}", func(t *testing.T) {
		body := `{"title": "Belajar Unit Test Go (Updated)", "is_done": true}`
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status code salah: dapat %d, mau %d", rec.Code, http.StatusOK)
		}

		var updated Todo
		if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
			t.Fatalf("gagal decode response json: %v", err)
		}
		if updated.Title != "Belajar Unit Test Go (Updated)" || !updated.IsDone {
			t.Errorf("data tidak sesuai update: %+v", updated)
		}
	})

	// Test 5: DELETE /todos/{id}
	t.Run("DELETE /todos/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status code salah: dapat %d, mau %d", rec.Code, http.StatusOK)
		}
	})

	// Test 6: GET todo yang sudah dihapus harus 404
	t.Run("GET non-existing /todos/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status code salah: dapat %d, mau %d", rec.Code, http.StatusNotFound)
		}
	})
}
