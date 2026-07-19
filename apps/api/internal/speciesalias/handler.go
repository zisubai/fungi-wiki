package speciesalias

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Handler struct{ repo Repository }

func RegisterPublicRoutes(r *gin.RouterGroup, x Repository) {
	r.GET("/:id/aliases", (&Handler{x}).list)
}
func RegisterAdminRoutes(r *gin.RouterGroup, x Repository) {
	h := &Handler{x}
	r.GET("/:id/aliases", h.list)
	r.PUT("/:id/aliases", h.replace)
}
func (h *Handler) list(c *gin.Context) {
	items, e := h.repo.List(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *Handler) replace(c *gin.Context) {
	var in ReplaceInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	seen := map[string]bool{}
	for i := range in.Items {
		in.Items[i].Name = strings.TrimSpace(in.Items[i].Name)
		key := strings.ToLower(in.Items[i].Name)
		if key == "" || seen[key] {
			c.JSON(400, gin.H{"message": "alias names must be non-empty and unique"})
			return
		}
		seen[key] = true
	}
	items, e := h.repo.Replace(c.Request.Context(), c.Param("id"), in.Items)
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func fail(c *gin.Context, e error) {
	if errors.Is(e, ErrSpeciesNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "species not found"})
	} else {
		c.JSON(500, gin.H{"message": "internal server error"})
	}
}
