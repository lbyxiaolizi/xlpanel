package knowledgebase

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrArticleNotFound  = errors.New("article not found")
	ErrCategoryNotFound = errors.New("category not found")
)

// Service provides knowledge base operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new knowledge base service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// --- Categories ---

// CreateCategory creates a new knowledge base category
func (s *Service) CreateCategory(name, description, iconClass string, parentID *uint64, sortOrder int) (*domain.KnowledgeBaseCategory, error) {
	slug := s.generateSlug(name)

	category := &domain.KnowledgeBaseCategory{
		ParentID:    parentID,
		Name:        name,
		Slug:        slug,
		Description: description,
		IconClass:   iconClass,
		SortOrder:   sortOrder,
		Active:      true,
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *Service) GetCategory(id uint64) (*domain.KnowledgeBaseCategory, error) {
	var category domain.KnowledgeBaseCategory
	if err := s.db.Preload("Children").Preload("Articles", "status = ?", "published").
		First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

// GetCategoryBySlug retrieves a category by slug
func (s *Service) GetCategoryBySlug(slug string) (*domain.KnowledgeBaseCategory, error) {
	var category domain.KnowledgeBaseCategory
	if err := s.db.Preload("Children").Preload("Articles", "status = ?", "published").
		Where("slug = ? AND active = ?", slug, true).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

// ListCategories lists all categories
func (s *Service) ListCategories(activeOnly bool) ([]domain.KnowledgeBaseCategory, error) {
	var categories []domain.KnowledgeBaseCategory
	query := s.db.Where("parent_id IS NULL").Order("sort_order ASC")
	if activeOnly {
		query = query.Where("active = ?", true)
	}
	if err := query.Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// UpdateCategory updates a category
func (s *Service) UpdateCategory(id uint64, name, description, iconClass string, sortOrder int, active bool) error {
	return s.db.Model(&domain.KnowledgeBaseCategory{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        name,
			"description": description,
			"icon_class":  iconClass,
			"sort_order":  sortOrder,
			"active":      active,
		}).Error
}

// DeleteCategory deletes a category
func (s *Service) DeleteCategory(id uint64) error {
	return s.db.Delete(&domain.KnowledgeBaseCategory{}, id).Error
}

// --- Articles ---

// CreateArticle creates a new knowledge base article
func (s *Service) CreateArticle(categoryID, authorID uint64, title, content, excerpt string, featured bool, tags []string) (*domain.KnowledgeBaseArticle, error) {
	slug := s.generateSlug(title)

	tagsMap := make(domain.JSONMap)
	tagsMap["tags"] = tags

	article := &domain.KnowledgeBaseArticle{
		CategoryID:    categoryID,
		Title:         title,
		Slug:          slug,
		Content:       content,
		Excerpt:       excerpt,
		AuthorID:      authorID,
		Status:        "draft",
		Featured:      featured,
		AllowComments: true,
		Tags:          tagsMap,
	}

	if err := s.db.Create(article).Error; err != nil {
		return nil, err
	}

	return article, nil
}

// GetArticle retrieves an article by ID
func (s *Service) GetArticle(id uint64) (*domain.KnowledgeBaseArticle, error) {
	var article domain.KnowledgeBaseArticle
	if err := s.db.Preload("Category").Preload("Author").Preload("Attachments").
		First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	return &article, nil
}

// GetArticleBySlug retrieves an article by slug
func (s *Service) GetArticleBySlug(slug string) (*domain.KnowledgeBaseArticle, error) {
	var article domain.KnowledgeBaseArticle
	if err := s.db.Preload("Category").Preload("Author").Preload("Attachments").
		Where("slug = ? AND status = ?", slug, "published").First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	return &article, nil
}

// ListArticles lists articles with optional filters
func (s *Service) ListArticles(categoryID *uint64, status string, featured bool, limit, offset int) ([]domain.KnowledgeBaseArticle, int64, error) {
	var articles []domain.KnowledgeBaseArticle
	var total int64

	query := s.db.Model(&domain.KnowledgeBaseArticle{})
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if featured {
		query = query.Where("featured = ?", true)
	}

	query.Count(&total)

	if err := query.Preload("Category").Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// UpdateArticle updates an article
func (s *Service) UpdateArticle(id uint64, title, content, excerpt, metaTitle, metaDescription string, featured bool, tags []string) error {
	tagsMap := make(domain.JSONMap)
	tagsMap["tags"] = tags

	return s.db.Model(&domain.KnowledgeBaseArticle{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":            title,
			"content":          content,
			"excerpt":          excerpt,
			"featured":         featured,
			"meta_title":       metaTitle,
			"meta_description": metaDescription,
			"tags":             tagsMap,
		}).Error
}

// PublishArticle publishes an article
func (s *Service) PublishArticle(id uint64) error {
	now := time.Now()
	return s.db.Model(&domain.KnowledgeBaseArticle{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": &now,
		}).Error
}

// UnpublishArticle unpublishes an article
func (s *Service) UnpublishArticle(id uint64) error {
	return s.db.Model(&domain.KnowledgeBaseArticle{}).Where("id = ?", id).
		Update("status", "draft").Error
}

// DeleteArticle deletes an article
func (s *Service) DeleteArticle(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete attachments
		if err := tx.Delete(&domain.KBArticleAttachment{}, "article_id = ?", id).Error; err != nil {
			return err
		}
		// Delete feedback
		if err := tx.Delete(&domain.KBArticleFeedback{}, "article_id = ?", id).Error; err != nil {
			return err
		}
		// Delete article
		return tx.Delete(&domain.KnowledgeBaseArticle{}, id).Error
	})
}

// IncrementViewCount increments the view count for an article
func (s *Service) IncrementViewCount(id uint64) error {
	return s.db.Model(&domain.KnowledgeBaseArticle{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// RecordFeedback records feedback on an article
func (s *Service) RecordFeedback(articleID uint64, customerID *uint64, helpful bool, comment, ipAddress string) error {
	feedback := &domain.KBArticleFeedback{
		ArticleID:  articleID,
		CustomerID: customerID,
		Helpful:    helpful,
		Comment:    comment,
		IPAddress:  ipAddress,
	}

	if err := s.db.Create(feedback).Error; err != nil {
		return err
	}

	// Update article counters
	if helpful {
		return s.db.Model(&domain.KnowledgeBaseArticle{}).Where("id = ?", articleID).
			Update("helpful_yes", gorm.Expr("helpful_yes + 1")).Error
	}
	return s.db.Model(&domain.KnowledgeBaseArticle{}).Where("id = ?", articleID).
		Update("helpful_no", gorm.Expr("helpful_no + 1")).Error
}

// SearchArticles searches articles by keyword
func (s *Service) SearchArticles(query string, limit int) ([]domain.KnowledgeBaseArticle, error) {
	var articles []domain.KnowledgeBaseArticle

	// Log the search
	s.logSearch(query, nil)

	// Search in title and content
	searchQuery := "%" + strings.ToLower(query) + "%"
	if err := s.db.Where("status = ? AND (LOWER(title) LIKE ? OR LOWER(content) LIKE ?)", "published", searchQuery, searchQuery).
		Preload("Category").
		Order("view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	return articles, nil
}

// GetPopularArticles returns the most viewed articles
func (s *Service) GetPopularArticles(limit int) ([]domain.KnowledgeBaseArticle, error) {
	var articles []domain.KnowledgeBaseArticle
	if err := s.db.Where("status = ?", "published").
		Preload("Category").
		Order("view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

// GetRelatedArticles returns related articles based on category and tags
func (s *Service) GetRelatedArticles(articleID uint64, limit int) ([]domain.KnowledgeBaseArticle, error) {
	var article domain.KnowledgeBaseArticle
	if err := s.db.First(&article, articleID).Error; err != nil {
		return nil, err
	}

	var articles []domain.KnowledgeBaseArticle
	if err := s.db.Where("id != ? AND category_id = ? AND status = ?", articleID, article.CategoryID, "published").
		Order("view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

// AddAttachment adds an attachment to an article
func (s *Service) AddAttachment(articleID uint64, fileName, filePath, contentType string, fileSize int64) (*domain.KBArticleAttachment, error) {
	attachment := &domain.KBArticleAttachment{
		ArticleID:   articleID,
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    fileSize,
		ContentType: contentType,
	}

	if err := s.db.Create(attachment).Error; err != nil {
		return nil, err
	}

	return attachment, nil
}

// DeleteAttachment deletes an attachment
func (s *Service) DeleteAttachment(attachmentID uint64) error {
	return s.db.Delete(&domain.KBArticleAttachment{}, attachmentID).Error
}

// IncrementAttachmentDownload increments the download count for an attachment
func (s *Service) IncrementAttachmentDownload(attachmentID uint64) error {
	return s.db.Model(&domain.KBArticleAttachment{}).Where("id = ?", attachmentID).
		Update("downloads", gorm.Expr("downloads + 1")).Error
}

// logSearch logs a search query
func (s *Service) logSearch(query string, customerID *uint64) {
	// Count results
	var count int64
	searchQuery := "%" + strings.ToLower(query) + "%"
	s.db.Model(&domain.KnowledgeBaseArticle{}).
		Where("status = ? AND (LOWER(title) LIKE ? OR LOWER(content) LIKE ?)", "published", searchQuery, searchQuery).
		Count(&count)

	log := &domain.KBSearchLog{
		Query:       query,
		ResultCount: int(count),
		CustomerID:  customerID,
	}
	s.db.Create(log)
}

// GetSearchStats returns popular search queries
func (s *Service) GetSearchStats(limit int) ([]SearchStat, error) {
	var stats []SearchStat
	if err := s.db.Model(&domain.KBSearchLog{}).
		Select("query, COUNT(*) as count, AVG(result_count) as avg_results").
		Group("query").
		Order("count DESC").
		Limit(limit).
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// generateSlug generates a URL-safe slug from a string
func (s *Service) generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}

// SearchStat represents search statistics
type SearchStat struct {
	Query      string  `json:"query"`
	Count      int     `json:"count"`
	AvgResults float64 `json:"avg_results"`
}
