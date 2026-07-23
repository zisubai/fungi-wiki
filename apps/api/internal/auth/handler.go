package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Handler struct {
	repo   Repository
	tokens *TokenService
}

func NewHandler(r Repository, t *TokenService) *Handler { return &Handler{r, t} }
func RegisterRoutes(r *gin.RouterGroup, repo Repository, tokens *TokenService) {
	h := NewHandler(repo, tokens)
	r.POST("/login", h.login)
	r.POST("/register", h.register)
	r.GET("/me", Authenticate(tokens), h.me)
}
func (h *Handler) register(c *gin.Context) {
	var input RegisterInput
	if e := c.ShouldBindJSON(&input); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	user, e := h.repo.Create(c.Request.Context(), CreateUserInput{Email: input.Email, Password: input.Password, DisplayName: input.DisplayName, Role: "member"})
	if e != nil {
		c.JSON(409, gin.H{"message": "email already exists"})
		return
	}
	token, expires, e := h.tokens.Issue(user)
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, LoginResponse{token, expires, user})
}
func RegisterAdminRoutes(r *gin.RouterGroup, repo Repository) {
	h := &Handler{repo: repo}
	r.GET("", h.listUsers)
	r.POST("", h.createUser)
}
func (h *Handler) login(c *gin.Context) {
	var in LoginInput
	if e := c.ShouldBindJSON(&in); e != nil {
		c.JSON(400, gin.H{"message": "请输入有效邮箱和密码"})
		return
	}
	u, e := h.repo.Authenticate(c.Request.Context(), in.Email, in.Password)
	if e != nil {
		if errors.Is(e, ErrInvalidCredentials) || errors.Is(e, ErrDisabled) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "邮箱或密码错误"})
		} else {
			c.JSON(500, gin.H{"message": "internal server error"})
		}
		return
	}
	token, expires, e := h.tokens.Issue(u)
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(200, LoginResponse{token, expires, u})
}
func (h *Handler) me(c *gin.Context) {
	u, e := h.repo.Get(c.Request.Context(), c.GetString("userID"))
	if e != nil {
		c.JSON(401, gin.H{"message": "unauthorized"})
		return
	}
	c.JSON(200, u)
}
func (h *Handler) listUsers(c *gin.Context) {
	items, e := h.repo.List(c.Request.Context())
	if e != nil {
		c.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *Handler) createUser(c *gin.Context) {
	var input CreateUserInput
	if e := c.ShouldBindJSON(&input); e != nil {
		c.JSON(400, gin.H{"message": e.Error()})
		return
	}
	user, e := h.repo.Create(c.Request.Context(), input)
	if e != nil {
		c.JSON(409, gin.H{"message": "email already exists"})
		return
	}
	c.JSON(http.StatusCreated, user)
}
func Authenticate(tokens *TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
			return
		}
		claims, e := tokens.Parse(strings.TrimPrefix(header, "Bearer "))
		if e != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("role", claims.Role)
		c.Next()
	}
}
func OptionalAuthenticate(tokens *TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			if claims, err := tokens.Parse(strings.TrimPrefix(header, "Bearer ")); err == nil {
				c.Set("userID", claims.Subject)
				c.Set("role", claims.Role)
			}
		}
		c.Next()
	}
}
func AuthorizeAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "admin" {
			c.Next()
			return
		}
		path := c.FullPath()
		method := c.Request.Method
		if strings.HasPrefix(path, "/api/admin/users") {
			c.AbortWithStatusJSON(403, gin.H{"message": "admin role required"})
			return
		}
		if role == "operator" {
			if strings.HasSuffix(path, "/approve") || strings.HasSuffix(path, "/reject") {
				c.AbortWithStatusJSON(403, gin.H{"message": "expert or admin role required"})
				return
			}
			c.Next()
			return
		}
		if role == "expert" {
			if method == http.MethodGet || strings.HasSuffix(path, "/approve") || strings.HasSuffix(path, "/reject") {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"message": "forbidden"})
	}
}
