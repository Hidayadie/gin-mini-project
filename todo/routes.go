package todo


import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan semua route todo ke group yang diberikan
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
  g := rg.Group("/todos")
  {
    g.GET("", h.List)
    g.GET("/:id", h.GetByID)
    g.POST("", h.Create)
    g.PUT("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
  }
}
