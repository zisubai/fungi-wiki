package importjob

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo Repository }

func RegisterAdminRoutes(router *gin.RouterGroup, repo Repository) {
	handler := &Handler{repo}
	router.GET("", handler.list)
	router.POST("/species", handler.importSpecies)
}

func (handler *Handler) importSpecies(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 10<<20)
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请选择不超过 10MB 的 CSV 或 XLSX 文件"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(400, gin.H{"message": "无法读取上传文件"})
		return
	}
	defer file.Close()
	rows, err := Parse(fileHeader.Filename, file)
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, ErrUnsupportedFile) {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, gin.H{"message": err.Error()})
		return
	}
	if len(rows) > 5000 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "单次最多导入 5000 行"})
		return
	}
	batch, err := handler.repo.Import(ctx.Request.Context(), fileHeader.Filename, rows)
	if err != nil {
		ctx.JSON(500, gin.H{"message": "import failed"})
		return
	}
	ctx.JSON(http.StatusCreated, batch)
}
func (handler *Handler) list(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	items, err := handler.repo.List(ctx.Request.Context(), limit)
	if err != nil {
		ctx.JSON(500, gin.H{"message": "internal server error"})
		return
	}
	ctx.JSON(200, gin.H{"items": items})
}
