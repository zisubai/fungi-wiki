package culturecondition

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct{ repo Repository }

func RegisterPublicRoutes(r *gin.RouterGroup, repo Repository) {
	r.GET("/:id/culture-conditions", (&Handler{repo}).list)
}
func RegisterAdminRoutes(r *gin.RouterGroup, repo Repository) {
	h := &Handler{repo}
	r.GET("/:id/culture-conditions", h.list)
	r.PUT("/:id/culture-conditions", h.replace)
}
func (h *Handler) list(c *gin.Context) {
	items, e := h.repo.List(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
func (h *Handler) replace(c *gin.Context) {
	var in ReplaceInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	for _, x := range in.Items {
		if x.PHMin != nil && (*x.PHMin < 0 || *x.PHMin > 14) || x.PHMax != nil && (*x.PHMax < 0 || *x.PHMax > 14) {
			c.JSON(400, gin.H{"message": "pH must be between 0 and 14"})
			return
		}
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
		c.JSON(404, gin.H{"message": "species not found"})
	} else {
		c.JSON(500, gin.H{"message": "internal server error"})
	}
}
