package todo

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Handler menyimpan dependensi yang dibutuhkan semua handler todo
type Handler struct {
	DB *gorm.DB
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

// List - GET /todos
// Query opsional: ?done=true|false, ?page=1, ?limit=10
func (h *Handler) List(c *gin.Context) {
	query := h.DB.WithContext(c.Request.Context()).Model(&Todo{})

	if done := c.Query("done"); done != "" {
		isDone, err := strconv.ParseBool(done)
		if err != nil {
			respondError(c, http.StatusBadRequest, "Parameter done harus true atau false")
			return
		}

		query = query.Where("is_done = ?", isDone)
	}

	// Pagination: nilai tidak valid diganti nilai bawaan
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Session(&gorm.Session{}) membuat salinan aman agar query yang sama
	// bisa dipakai dua kali (Count lalu Find) tanpa saling memengaruhi
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		log.Println("List count:", err)
		respondError(c, http.StatusInternalServerError, "Gagal mengambil data")
		return
	}

	todos := []Todo{} // slice kosong agar hasilnya [] (bukan null)

	err := query.Order("id").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&todos).Error

	if err != nil {
		log.Println("List find:", err)
		respondError(c, http.StatusInternalServerError, "Gagal mengambil data")
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, todos)
}

// GetByID - GET /todos/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var t Todo
	err := h.DB.WithContext(c.Request.Context()).First(&t, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
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

	// ID diisi otomatis oleh GORM setelah INSERT berhasil
	t := Todo{
		Title:  req.Title,
		IsDone: req.IsDone,
	}

	if err := h.DB.WithContext(c.Request.Context()).Create(&t).Error; err != nil {
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

	db := h.DB.WithContext(c.Request.Context())

	var t Todo
	if err := db.First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
			return
		}

		log.Println("Update find:", err)
		respondError(c, http.StatusInternalServerError, "Gagal mengubah data")
		return
	}

	// Pakai map agar is_done = false tetap ikut diubah
	// (Updates dengan struct mengabaikan field bernilai nol)
	err := db.Model(&t).Updates(map[string]any{
		"title":   req.Title,
		"is_done": req.IsDone,
	}).Error

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

	result := h.DB.WithContext(c.Request.Context()).Delete(&Todo{}, id)

	if result.Error != nil {
		log.Println("Delete:", result.Error)
		respondError(c, http.StatusInternalServerError, "Gagal menghapus data")
		return
	}

	// RowsAffected = jumlah baris yang terhapus; 0 berarti ID tidak ada
	if result.RowsAffected == 0 {
		respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo berhasil dihapus"})
}
