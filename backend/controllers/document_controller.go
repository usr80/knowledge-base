package controllers

import (
	"net/http"
	"strconv"

	"knowledge-base/services"

	"github.com/gin-gonic/gin"
)

type DocumentController struct {
	docService *services.DocumentService
}

func NewDocumentController() *DocumentController {
	return &DocumentController{
		docService: services.NewDocumentService(),
	}
}

type CreateDocumentRequest struct {
	Title      string   `json:"title" binding:"required,max=255"`
	Content    string   `json:"content"`
	Summary    string   `json:"summary"`
	CategoryID *uint    `json:"categoryID"`
	Tags       []string `json:"tags"`
}

func (c *DocumentController) Create(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var req CreateDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := c.docService.Create(userID.(uint), req.Title, req.Content, req.Summary, req.CategoryID, req.Tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"data":    doc,
	})
}

func (c *DocumentController) GetByID(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档 ID"})
		return
	}

	doc, err := c.docService.GetByID(uint(id), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": doc})
}

type ListDocumentsRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	CategoryID *uint  `form:"categoryID"`
	Keyword    string `form:"keyword"`
}

func (c *DocumentController) List(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var req ListDocumentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	docs, total, err := c.docService.List(userID.(uint), req.Page, req.PageSize, req.CategoryID, req.Keyword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list":      docs,
			"total":     total,
			"page":      req.Page,
			"pageSize":  req.PageSize,
		},
	})
}

type UpdateDocumentRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Summary    string   `json:"summary"`
	CategoryID *uint    `json:"categoryID"`
	Tags       []string `json:"tags"`
}

func (c *DocumentController) Update(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档 ID"})
		return
	}

	var req UpdateDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.docService.Update(userID.(uint), uint(id), req.Title, req.Content, req.Summary, req.CategoryID, req.Tags); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (c *DocumentController) Delete(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档 ID"})
		return
	}

	if err := c.docService.Delete(userID.(uint), uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}