package controllers

import (
	"net/http"
	"strconv"

	"knowledge-base/services"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	catService *services.CategoryService
}

func NewCategoryController() *CategoryController {
	return &CategoryController{
		catService: services.NewCategoryService(),
	}
}

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	ParentID *uint  `json:"parentID"`
	Icon     string `json:"icon"`
}

func (c *CategoryController) Create(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var req CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := c.catService.Create(userID.(uint), req.Name, req.ParentID, req.Icon)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"data":    category,
	})
}

func (c *CategoryController) List(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	categories, err := c.catService.List(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": categories})
}

func (c *CategoryController) GetByID(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类 ID"})
		return
	}

	category, err := c.catService.GetByID(uint(id), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": category})
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func (c *CategoryController) Update(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类 ID"})
		return
	}

	var req UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.catService.Update(userID.(uint), uint(id), req.Name, req.Icon); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (c *CategoryController) Delete(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类 ID"})
		return
	}

	if err := c.catService.Delete(userID.(uint), uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}