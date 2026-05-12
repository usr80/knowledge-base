package controllers

import (
	"net/http"

	"knowledge-base/services"

	"github.com/gin-gonic/gin"
)

type SearchController struct{}

func NewSearchController() *SearchController {
	return &SearchController{}
}

type SearchRequest struct {
	Keyword    string `form:"keyword"`
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	CategoryID *uint  `form:"categoryID"`
}

// Search 搜索文档
func (c *SearchController) Search(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req SearchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	searchSvc := services.GetSearchService()

	// 如果没有关键词，返回空结果
	if req.Keyword == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"list":     []interface{}{},
				"total":    0,
				"page":     req.Page,
				"pageSize": req.PageSize,
			},
		})
		return
	}

	result, err := searchSvc.Search(userID.(uint), req.Keyword, req.Page, req.PageSize, req.CategoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list":     result.Documents,
			"total":    result.Total,
			"page":     result.Page,
			"pageSize": result.PageSize,
		},
	})
}

// RebuildIndex 重建索引（管理员功能）
func (c *SearchController) RebuildIndex(ctx *gin.Context) {
	searchSvc := services.GetSearchService()

	if err := searchSvc.RebuildIndex(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "索引重建任务已提交"})
}
