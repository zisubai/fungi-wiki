package recommendation

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"strings"
)

func RegisterPublicRoutes(r *gin.RouterGroup, repo Repository) {
	r.POST("/combinations", func(c *gin.Context) {
		var input CombinationInput
		if err := c.ShouldBindJSON(&input); err != nil || input.FunctionTags[0] == input.FunctionTags[1] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "select two different function tags"})
			return
		}
		response, err := repo.RecommendCombination(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "combination recommendation failed"})
			return
		}
		c.JSON(http.StatusOK, response)
	})
	r.POST("", func(c *gin.Context) {
		var in Input
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(400, gin.H{"message": "requirement is required"})
			return
		}
		in.Requirement = strings.TrimSpace(in.Requirement)
		if len([]rune(in.Requirement)) < 2 || len([]rune(in.Requirement)) > 2000 {
			c.JSON(400, gin.H{"message": "requirement must contain 2 to 2000 characters"})
			return
		}
		if in.PH != nil && (*in.PH < 0 || *in.PH > 14) {
			c.JSON(400, gin.H{"message": "ph must be between 0 and 14"})
			return
		}
		response, err := repo.Recommend(c.Request.Context(), in)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "recommendation failed"})
			return
		}
		c.JSON(200, response)
	})
}

func RegisterFeedbackRoutes(r *gin.RouterGroup, repo Repository) {
	r.POST("/combinations/:id/feedback", func(c *gin.Context) {
		var id pgtype.UUID
		if err := id.Scan(c.Param("id")); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid combination recommendation id"})
			return
		}
		var input FeedbackInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "feedbackType must be helpful or unhelpful"})
			return
		}
		input.Content = strings.TrimSpace(input.Content)
		if len([]rune(input.Content)) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "content must not exceed 1000 characters"})
			return
		}
		if err := repo.CombinationFeedback(c.Request.Context(), c.Param("id"), input); err != nil {
			if errors.Is(err, ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"message": "combination recommendation not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "feedback failed"})
			}
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "feedback recorded"})
	})
	r.POST("/:id/feedback", func(c *gin.Context) {
		var id pgtype.UUID
		if err := id.Scan(c.Param("id")); err != nil {
			c.JSON(400, gin.H{"message": "invalid recommendation id"})
			return
		}
		var input FeedbackInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"message": "feedbackType must be helpful or unhelpful"})
			return
		}
		input.Content = strings.TrimSpace(input.Content)
		if len([]rune(input.Content)) > 1000 {
			c.JSON(400, gin.H{"message": "content must not exceed 1000 characters"})
			return
		}
		if err := repo.Feedback(c.Request.Context(), c.Param("id"), input); err != nil {
			if errors.Is(err, ErrNotFound) {
				c.JSON(404, gin.H{"message": "recommendation not found"})
			} else {
				c.JSON(500, gin.H{"message": "feedback failed"})
			}
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "feedback recorded"})
	})
}
func RegisterAdminRoutes(r *gin.RouterGroup, repo Repository) {
	r.POST("/combinations/:id/experiments", func(c *gin.Context) {
		var id pgtype.UUID
		if err := id.Scan(c.Param("id")); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid combination recommendation id"})
			return
		}
		var input CombinationExperimentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "outcome must be compatible, incompatible or inconclusive"})
			return
		}
		input.Notes = strings.TrimSpace(input.Notes)
		if input.CandidateIndex == nil || *input.CandidateIndex < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "candidateIndex must be zero or greater"})
			return
		}
		if input.PH != nil && (*input.PH < 0 || *input.PH > 14) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "ph must be between 0 and 14"})
			return
		}
		if len([]rune(input.Notes)) > 2000 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "notes must not exceed 2000 characters"})
			return
		}
		experiment, err := repo.CreateCombinationExperiment(c.Request.Context(), c.Param("id"), input)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"message": "combination recommendation or candidate not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "experiment creation failed"})
			}
			return
		}
		c.JSON(http.StatusCreated, experiment)
	})
	r.GET("", func(c *gin.Context) {
		report, err := repo.Quality(c.Request.Context(), 30)
		if err != nil {
			c.JSON(500, gin.H{"message": "internal server error"})
			return
		}
		c.JSON(200, report)
	})
}
