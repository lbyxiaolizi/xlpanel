package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/service/knowledgebase"
)

// KnowledgeBaseHandler handles knowledge base API endpoints
type KnowledgeBaseHandler struct {
	service *knowledgebase.Service
}

// NewKnowledgeBaseHandler creates a new knowledge base handler
func NewKnowledgeBaseHandler(service *knowledgebase.Service) *KnowledgeBaseHandler {
	return &KnowledgeBaseHandler{service: service}
}

// ListCategories lists knowledge base categories
// @Summary List categories
// @Description Get a list of knowledge base categories
// @Tags Knowledge Base
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kb/categories [get]
func (h *KnowledgeBaseHandler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(true) // Public only
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetCategory gets a category with its articles
// @Summary Get category
// @Description Get a category with its articles
// @Tags Knowledge Base
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kb/categories/{slug} [get]
func (h *KnowledgeBaseHandler) GetCategory(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.service.GetCategoryBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	// Pass category.ID as pointer
	catID := category.ID
	articles, _, err := h.service.ListArticles(&catID, "published", false, 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
		"articles": articles,
	})
}

// GetArticle gets a knowledge base article
// @Summary Get article
// @Description Get a knowledge base article by slug
// @Tags Knowledge Base
// @Produce json
// @Param slug path string true "Article slug"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kb/articles/{slug} [get]
func (h *KnowledgeBaseHandler) GetArticle(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.service.GetArticleBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	// Increment view count
	h.service.IncrementViewCount(article.ID)

	// Get related articles
	related, err := h.service.GetRelatedArticles(article.ID, 5)
	if err != nil {
		related = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"article": article,
		"related": related,
	})
}

// SearchArticles searches knowledge base articles
// @Summary Search articles
// @Description Search knowledge base articles
// @Tags Knowledge Base
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kb/search [get]
func (h *KnowledgeBaseHandler) SearchArticles(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	results, err := h.service.SearchArticles(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// RateArticle rates an article
// @Summary Rate article
// @Description Rate a knowledge base article as helpful or not
// @Tags Knowledge Base
// @Accept json
// @Produce json
// @Param slug path string true "Article slug"
// @Param request body RateArticleRequest true "Rating request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kb/articles/{slug}/rate [post]
func (h *KnowledgeBaseHandler) RateArticle(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.service.GetArticleBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var req RateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use RecordFeedback instead of RateArticle
	if err := h.service.RecordFeedback(article.ID, nil, req.Helpful, "", c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thank you for your feedback"})
}

// GetPopularArticles gets popular articles
// @Summary Get popular articles
// @Description Get the most popular knowledge base articles
// @Tags Knowledge Base
// @Produce json
// @Param limit query int false "Limit results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kb/popular [get]
func (h *KnowledgeBaseHandler) GetPopularArticles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	articles, err := h.service.GetPopularArticles(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// Admin handlers

// AdminListCategories lists all categories including hidden
// @Summary Admin: List all categories
// @Description Get a list of all categories (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/categories [get]
func (h *KnowledgeBaseHandler) AdminListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(false) // All including hidden
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// AdminCreateCategory creates a category
// @Summary Admin: Create category
// @Description Create a knowledge base category (admin only)
// @Tags Admin Knowledge Base
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true "Category request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/categories [post]
func (h *KnowledgeBaseHandler) AdminCreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.CreateCategory(req.Name, req.Description, "", req.ParentID, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Category created",
		"category": category,
	})
}

// AdminUpdateCategory updates a category
// @Summary Admin: Update category
// @Description Update a knowledge base category (admin only)
// @Tags Admin Knowledge Base
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body UpdateCategoryRequest true "Category request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/categories/{id} [put]
func (h *KnowledgeBaseHandler) AdminUpdateCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateCategory(categoryID, req.Name, req.Description, "", req.SortOrder, req.Public); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category updated"})
}

// AdminDeleteCategory deletes a category
// @Summary Admin: Delete category
// @Description Delete a knowledge base category (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/categories/{id} [delete]
func (h *KnowledgeBaseHandler) AdminDeleteCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := h.service.DeleteCategory(categoryID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// AdminListArticles lists all articles
// @Summary Admin: List all articles
// @Description Get a list of all articles (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Param category query int false "Category ID"
// @Param status query string false "Status filter"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/articles [get]
func (h *KnowledgeBaseHandler) AdminListArticles(c *gin.Context) {
	categoryStr := c.Query("category")
	var categoryID *uint64
	if categoryStr != "" {
		catID, _ := strconv.ParseUint(categoryStr, 10, 64)
		categoryID = &catID
	}
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	articles, total, err := h.service.ListArticles(categoryID, status, false, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"total":    total,
	})
}

// AdminCreateArticle creates an article
// @Summary Admin: Create article
// @Description Create a knowledge base article (admin only)
// @Tags Admin Knowledge Base
// @Accept json
// @Produce json
// @Param request body CreateArticleRequest true "Article request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/articles [post]
func (h *KnowledgeBaseHandler) AdminCreateArticle(c *gin.Context) {
	adminID, _ := c.Get("admin_id")

	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.service.CreateArticle(
		req.CategoryID,
		adminID.(uint64),
		req.Title,
		req.Content,
		req.Excerpt,
		false,
		req.Tags,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Article created",
		"article": article,
	})
}

// AdminUpdateArticle updates an article
// @Summary Admin: Update article
// @Description Update a knowledge base article (admin only)
// @Tags Admin Knowledge Base
// @Accept json
// @Produce json
// @Param id path int true "Article ID"
// @Param request body UpdateArticleRequest true "Article request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/articles/{id} [put]
func (h *KnowledgeBaseHandler) AdminUpdateArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateArticle(
		articleID,
		req.Title,
		req.Content,
		req.Excerpt,
		"",
		"",
		false,
		req.Tags,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article updated"})
}

// AdminPublishArticle publishes an article
// @Summary Admin: Publish article
// @Description Publish a knowledge base article (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/articles/{id}/publish [post]
func (h *KnowledgeBaseHandler) AdminPublishArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	if err := h.service.PublishArticle(articleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article published"})
}

// AdminUnpublishArticle unpublishes an article
// @Summary Admin: Unpublish article
// @Description Unpublish a knowledge base article (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/articles/{id}/unpublish [post]
func (h *KnowledgeBaseHandler) AdminUnpublishArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	if err := h.service.UnpublishArticle(articleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article unpublished"})
}

// AdminDeleteArticle deletes an article
// @Summary Admin: Delete article
// @Description Delete a knowledge base article (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/articles/{id} [delete]
func (h *KnowledgeBaseHandler) AdminDeleteArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	if err := h.service.DeleteArticle(articleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}

// AdminGetSearchStats gets search statistics
// @Summary Admin: Get search statistics
// @Description Get popular search queries (admin only)
// @Tags Admin Knowledge Base
// @Produce json
// @Param limit query int false "Limit results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/kb/search-stats [get]
func (h *KnowledgeBaseHandler) AdminGetSearchStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	stats, err := h.service.GetSearchStats(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// Request/Response types
type RateArticleRequest struct {
	Helpful bool `json:"helpful"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
	Public      bool    `json:"public"`
}

type UpdateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
	Public      bool    `json:"public"`
}

type CreateArticleRequest struct {
	CategoryID uint64   `json:"category_id" binding:"required"`
	Title      string   `json:"title" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Excerpt    string   `json:"excerpt"`
	Tags       []string `json:"tags"`
}

type UpdateArticleRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Excerpt string   `json:"excerpt"`
	Tags    []string `json:"tags"`
}
