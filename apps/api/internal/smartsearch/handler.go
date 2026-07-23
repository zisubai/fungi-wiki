package smartsearch

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo *Repository }

func RegisterPublicRoutes(group *gin.RouterGroup, repo *Repository) {
	group.GET("", (&Handler{repo}).search)
}
func RegisterUserRoutes(group *gin.RouterGroup, repo *Repository) {
	h := &Handler{repo}
	group.GET("/favorites", h.favorites)
	group.PUT("/favorites/:speciesId", h.addFavorite)
	group.DELETE("/favorites/:speciesId", h.removeFavorite)
	group.GET("/search-history", h.history)
	group.DELETE("/search-history", h.clearHistory)
}
func RegisterAdminRoutes(group *gin.RouterGroup, repo *Repository) {
	h := &Handler{repo}
	group.GET("/synonyms", h.synonyms)
	group.POST("/synonyms", h.saveSynonym)
	group.PUT("/synonyms/:id", h.saveSynonym)
	group.DELETE("/synonyms/:id", h.deleteSynonym)
	group.GET("/rules", h.rules)
	group.POST("/rules", h.saveRule)
	group.PUT("/rules/:id", h.saveRule)
	group.DELETE("/rules/:id", h.deleteRule)
	group.POST("/reindex", h.reindex)
}
func intQuery(c *gin.Context, key string, fallback int) int {
	x, e := strconv.Atoi(c.Query(key))
	if e != nil {
		return fallback
	}
	return x
}
func (h *Handler) search(c *gin.Context) {
	temperature, e := parseFloat(c.Query("temperature"))
	if e != nil {
		c.JSON(400, gin.H{"message": "temperature must be a number"})
		return
	}
	ph, e := parseFloat(c.Query("ph"))
	if e != nil || ph != nil && (*ph < 0 || *ph > 14) {
		c.JSON(400, gin.H{"message": "ph must be between 0 and 14"})
		return
	}
	p := Params{Query: c.Query("q"), FunctionTag: c.Query("functionTag"), Temperature: temperature, PH: ph, SafetyLevel: c.Query("safetyLevel"), SourceEnvironment: c.Query("sourceEnvironment"), Sort: c.Query("sort"), Limit: intQuery(c, "limit", 20), Offset: intQuery(c, "offset", 0)}
	result, e := h.repo.Search(c.Request.Context(), p)
	if e != nil {
		c.JSON(500, gin.H{"message": "search failed"})
		return
	}
	if userID := c.GetString("userID"); userID != "" && (strings.TrimSpace(p.Query) != "" || p.FunctionTag != "") {
		_ = h.repo.LogHistory(c.Request.Context(), userID, p, result.Total)
	}
	c.JSON(200, result)
}
func (h *Handler) favorites(c *gin.Context) {
	items, e := h.repo.Favorites(c.Request.Context(), c.GetString("userID"))
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *Handler) addFavorite(c *gin.Context) {
	if e := h.repo.Favorite(c.Request.Context(), c.GetString("userID"), c.Param("speciesId"), true); e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.Status(204)
}
func (h *Handler) removeFavorite(c *gin.Context) {
	if e := h.repo.Favorite(c.Request.Context(), c.GetString("userID"), c.Param("speciesId"), false); e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.Status(204)
}
func (h *Handler) history(c *gin.Context) {
	items, e := h.repo.History(c.Request.Context(), c.GetString("userID"))
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *Handler) clearHistory(c *gin.Context) {
	if e := h.repo.ClearHistory(c.Request.Context(), c.GetString("userID")); e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.Status(204)
}
func (h *Handler) synonyms(c *gin.Context) {
	x, e := h.repo.ListSynonyms(c.Request.Context())
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"items": x})
}
func (h *Handler) saveSynonym(c *gin.Context) {
	var in SynonymInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	x, e := h.repo.SaveSynonym(c.Request.Context(), c.Param("id"), in)
	if e != nil {
		c.JSON(409, gin.H{"message": e.Error()})
		return
	}
	status := http.StatusCreated
	if c.Param("id") != "" {
		status = 200
	}
	c.JSON(status, x)
}
func (h *Handler) deleteSynonym(c *gin.Context) {
	if e := h.repo.DeleteSynonym(c.Request.Context(), c.Param("id")); e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.Status(204)
}
func (h *Handler) rules(c *gin.Context) {
	x, e := h.repo.ListRules(c.Request.Context())
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"items": x})
}
func (h *Handler) saveRule(c *gin.Context) {
	var in RuleInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	x, e := h.repo.SaveRule(c.Request.Context(), c.Param("id"), in)
	if e != nil {
		c.JSON(409, gin.H{"message": e.Error()})
		return
	}
	status := http.StatusCreated
	if c.Param("id") != "" {
		status = 200
	}
	c.JSON(status, x)
}
func (h *Handler) deleteRule(c *gin.Context) {
	if e := h.repo.DeleteRule(c.Request.Context(), c.Param("id")); e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.Status(204)
}
func (h *Handler) reindex(c *gin.Context) {
	count, e := h.repo.Reindex(c.Request.Context())
	if e != nil {
		c.JSON(503, gin.H{"message": e.Error()})
		return
	}
	c.JSON(200, gin.H{"indexed": count})
}
