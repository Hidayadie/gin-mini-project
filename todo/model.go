package todo

// Todo adalah data satu todo. Tag gorm menjelaskan kolom di tabel todos.
type Todo struct {
	ID     int    `json:"id" gorm:"primaryKey"`
	Title  string `json:"title" gorm:"not null"`
	IsDone bool   `json:"is_done" gorm:"not null"`
}

// CreateTodoRequest adalah body JSON untuk POST /todos
type CreateTodoRequest struct {
	Title  string `json:"title" binding:"required,min=3"`
	IsDone bool   `json:"is_done"`
}

// UpdateTodoRequest adalah body JSON untuk PUT /todos/:id
type UpdateTodoRequest struct {
	Title  string `json:"title" binding:"required,min=3"`
	IsDone bool   `json:"is_done"`
}

