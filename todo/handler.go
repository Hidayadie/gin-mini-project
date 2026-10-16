package todo

import (
  "database/sql"
  "errors"
  "fmt"
  "log"
  "net/http"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
)

// Handler menyimpan dependensi yang dibutuhkan semua handler todo
type Handler struct {
  DB *sql.DB
}

// respondError mengirim pesan error dengan format seragam {"error": "..."}
func respondError(c *gin.Context, status int, message string) {
  c.JSON(status, gin.H{"error": message})
}

// validationMessage mengubah error binding menjadi pesan yang ramah
func validationMessage(err error) string {
  var ve validator.ValidationErrors
  if errors.As(err, &ve) && len(ve) > 0 {
    fe := ve[0]
    switch fe.Tag() {
    case "required":
      return fe.Field() + " wajib diisi"
    case "min":
      return fe.Field() + " minimal " + fe.Param() + " karakter"
    }
    return fe.Field() + " tidak valid"
  }
  return "Format JSON tidak valid"
}

// parseID mengambil path param :id dan mengubahnya menjadi angka
func parseID(c *gin.Context) (int, bool) {
  id, err := strconv.Atoi(c.Param("id"))
  if err != nil {
    respondError(c, http.StatusBadRequest, "ID harus berupa angka")
    return 0, false
  }
  return id, true
}

func (h *Handler) List(c *gin.Context) {
  query := "SELECT id, title, is_done FROM todos"
  args := []any{}

  fmt.Println("query =>", query)

  if done := c.Query("done"); done != "" {
    isDone, err := strconv.ParseBool(done)
    if err != nil {
      respondError(c, http.StatusBadRequest, "Parameter done harus true atau false")
      return
    }
    query += " WHERE is_done = $1"
    args = append(args, isDone)
  }

  rows, err := h.DB.QueryContext(c.Request.Context(), query+" ORDER BY id", args...)
  if err != nil {
    log.Println("List:", err)
    respondError(c, http.StatusInternalServerError, "Gagal mengambil data")
    return
  }
  defer rows.Close()

  todos := []Todo{} // slice kosong agar hasilnya [] (bukan null)
  for rows.Next() {
    var t Todo
    if err := rows.Scan(&t.ID, &t.Title, &t.IsDone); err != nil {
      log.Println("List scan:", err)
      respondError(c, http.StatusInternalServerError, "Gagal membaca data")
      return
    }
    todos = append(todos, t)
  }
  if err := rows.Err(); err != nil {
    log.Println("List rows:", err)
    respondError(c, http.StatusInternalServerError, "Gagal membaca data")
    return
  }

  c.JSON(http.StatusOK, todos)
}

// GetByID - GET /todos/:id
func (h *Handler) GetByID(c *gin.Context) {
  id, ok := parseID(c)
  if !ok {
    return
  }

  var t Todo
  err := h.DB.QueryRowContext(c.Request.Context(),
    "SELECT id, title, is_done FROM todos WHERE id = $1", id).
    Scan(&t.ID, &t.Title, &t.IsDone)
  if errors.Is(err, sql.ErrNoRows) {
    respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
    return
  }
  if err != nil {
    log.Println("GetByID:", err)
    respondError(c, http.StatusInternalServerError, "Gagal mengambil data")
    return
  }

  c.JSON(http.StatusOK, t)
}

// Create - POST /todos
func (h *Handler) Create(c *gin.Context) {
  var req CreateTodoRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    respondError(c, http.StatusBadRequest, validationMessage(err))
    return
  }

  t := Todo{Title: req.Title, IsDone: req.IsDone}
  err := h.DB.QueryRowContext(c.Request.Context(),
    "INSERT INTO todos (title, is_done) VALUES ($1, $2) RETURNING id",
    t.Title, t.IsDone).Scan(&t.ID)
  if err != nil {
    log.Println("Create:", err)
    respondError(c, http.StatusInternalServerError, "Gagal menyimpan data")
    return
  }

  c.JSON(http.StatusCreated, t)
}

// Update - PUT /todos/:id
func (h *Handler) Update(c *gin.Context) {
  id, ok := parseID(c)
  if !ok {
    return
  }

  var req UpdateTodoRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    respondError(c, http.StatusBadRequest, validationMessage(err))
    return
  }

  t := Todo{Title: req.Title, IsDone: req.IsDone}
  err := h.DB.QueryRowContext(c.Request.Context(),
    "UPDATE todos SET title = $1, is_done = $2 WHERE id = $3 RETURNING id",
    t.Title, t.IsDone, id).Scan(&t.ID)
  if errors.Is(err, sql.ErrNoRows) {
    respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
    return
  }
  if err != nil {
    log.Println("Update:", err)
    respondError(c, http.StatusInternalServerError, "Gagal mengubah data")
    return
  }

  c.JSON(http.StatusOK, t)
}

// Delete - DELETE /todos/:id
func (h *Handler) Delete(c *gin.Context) {
  id, ok := parseID(c)
  if !ok {
    return
  }

  result, err := h.DB.ExecContext(c.Request.Context(), "DELETE FROM todos WHERE id = $1", id)
  if err != nil {
    log.Println("Delete:", err)
    respondError(c, http.StatusInternalServerError, "Gagal menghapus data")
    return
  }

  // RowsAffected = jumlah baris yang terhapus; 0 berarti ID tidak ada
  if n, _ := result.RowsAffected(); n == 0 {
    respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "Todo berhasil dihapus"})
}
