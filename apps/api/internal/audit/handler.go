package audit

import (
	"errors"
	"github.com/gin-gonic/gin"
)

type Handler struct{ repo Repository }

func RegisterAdminRoutes(r *gin.RouterGroup, x Repository) {
	h := &Handler{x}
	r.GET("", h.list)
	r.POST("/species/:id/submit", h.submit)
	r.POST("/:id/approve", h.approve)
	r.POST("/:id/reject", h.reject)
}
func (h *Handler) list(c *gin.Context) {
	xs, e := h.repo.List(c.Request.Context(), c.Query("status"))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, gin.H{"items": xs})
}
func (h *Handler) submit(c *gin.Context) {
	x, e := h.repo.Submit(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(201, x)
}
func (h *Handler) review(c *gin.Context, ok bool) {
	var in ReviewInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	x, e := h.repo.Review(c.Request.Context(), c.Param("id"), ok, in.Comment)
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, x)
}
func (h *Handler) approve(c *gin.Context) { h.review(c, true) }
func (h *Handler) reject(c *gin.Context)  { h.review(c, false) }
func fail(c *gin.Context, e error) {
	switch {
	case errors.Is(e, ErrNotFound), errors.Is(e, ErrSpeciesNotFound):
		c.JSON(404, gin.H{"message": e.Error()})
	case errors.Is(e, ErrInvalidState):
		c.JSON(409, gin.H{"message": e.Error()})
	default:
		c.JSON(500, gin.H{"message": "internal server error"})
	}
}
