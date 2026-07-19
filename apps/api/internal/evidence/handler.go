package evidence

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct{ repo Repository }

func RegisterPublicRoutes(r *gin.RouterGroup, x Repository) {
	r.GET("/:id/evidences", (&Handler{x}).list)
}
func RegisterAdminRoutes(r *gin.RouterGroup, x Repository) {
	h := &Handler{x}
	r.GET("/:id/evidences", h.list)
	r.POST("/:id/evidences", h.create)
	r.DELETE("/:id/evidences/:evidenceId", h.delete)
}
func (h *Handler) list(c *gin.Context) {
	xs, e := h.repo.List(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, gin.H{"items": xs})
}
func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	if in.EvidenceScore < 0 || in.EvidenceScore > 100 {
		c.JSON(400, gin.H{"message": "evidenceScore must be between 0 and 100"})
		return
	}
	x, e := h.repo.Create(c.Request.Context(), c.Param("id"), in)
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusCreated, x)
}
func (h *Handler) delete(c *gin.Context) {
	e := h.repo.Delete(c.Request.Context(), c.Param("id"), c.Param("evidenceId"))
	if e != nil {
		fail(c, e)
		return
	}
	c.Status(204)
}
func fail(c *gin.Context, e error) {
	if errors.Is(e, ErrNotFound) || errors.Is(e, ErrSpeciesNotFound) {
		c.JSON(404, gin.H{"message": e.Error()})
	} else {
		c.JSON(500, gin.H{"message": "internal server error"})
	}
}
