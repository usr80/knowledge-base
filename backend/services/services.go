package services

import (
	"errors"
	"time"

	"knowledge-base/config"
	"knowledge-base/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Register(username string, email *string, password string) (*models.User, error) {
	// 检查用户名是否存在
	var existUser models.User
	if err := config.DB.Where("username = ?", username).First(&existUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}
	if email != nil && *email != "" {
		if err := config.DB.Where("email = ?", *email).First(&existUser).Error; err == nil {
			return nil, errors.New("邮箱已被使用")
		}
	}

	user := &models.User{
		Username: username,
		Email:     email,
		Nickname:  username,
		Role:     "user",
		Status:   1,
	}
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := config.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(username, password string) (string, *models.User, error) {
	var user models.User
	if err := config.DB.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		// 安全考虑：不区分用户不存在和密码错误
		return "", nil, errors.New("用户名或密码错误")
	}

	if !user.CheckPassword(password) {
		return "", nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return "", nil, errors.New("账户已被禁用")
	}

	// 生成 JWT
	cfg := config.LoadConfig()
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * time.Duration(cfg.JWT.ExpireHour)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateUser(id uint, nickname, email, avatar string) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if email != "" {
		updates["email"] = email
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	return config.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	if !user.CheckPassword(oldPassword) {
		return errors.New("原密码错误")
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	return config.DB.Model(&user).Update("password", user.Password).Error
}

type DocumentService struct{}

func NewDocumentService() *DocumentService {
	return &DocumentService{}
}

func (s *DocumentService) Create(userID uint, title, content, summary string, categoryID *uint, tags []string) (*models.Document, error) {
	doc := &models.Document{
		UserID:      userID,
		Title:      title,
		Content:    content,
		Summary:    summary,
		CategoryID: categoryID,
		Status:     1,
	}

	if err := config.DB.Create(doc).Error; err != nil {
		return nil, err
	}

	// 处理标签
	if len(tags) > 0 {
		for _, tagName := range tags {
			var tag models.Tag
			result := config.DB.Where("user_id = ? AND name = ?", userID, tagName).First(&tag)
			if result.Error == gorm.ErrRecordNotFound {
				tag = models.Tag{
					UserID: userID,
					Name:  tagName,
					Color: "#1890ff",
				}
				config.DB.Create(&tag)
			}
			config.DB.Model(doc).Association("Tags").Append(&tag)
		}
	}

	// 同步到搜索索引
	go func() {
		searchSvc := GetSearchService()
		// 重新加载关联数据
		var fullDoc models.Document
		if err := config.DB.Preload("Category").Preload("Tags").First(&fullDoc, doc.ID).Error; err == nil {
			searchSvc.IndexDocument(&fullDoc)
		}
	}()

	return doc, nil
}

func (s *DocumentService) GetByID(id, userID uint) (*models.Document, error) {
	var doc models.Document
	if err := config.DB.Preload("Category").Preload("Tags").Where("id = ? AND user_id = ?", id, userID).First(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *DocumentService) List(userID uint, page, pageSize int, categoryID *uint, keyword string) ([]models.Document, int64, error) {
	var docs []models.Document
	var total int64

	query := config.DB.Model(&models.Document{}).Where("user_id = ?", userID)
	
	if categoryID != nil && *categoryID > 0 {
		query = query.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	query.Count(&total)
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	
	offset := (page - 1) * pageSize
	if err := query.Preload("Category").Preload("Tags").Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&docs).Error; err != nil {
		return nil, 0, err
	}
	
	return docs, total, nil
}

func (s *DocumentService) Update(userID uint, id uint, title, content, summary string, categoryID *uint, tags []string) error {
	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if content != "" {
		updates["content"] = content
	}
	if summary != "" {
		updates["summary"] = summary
	}
	if categoryID != nil {
		updates["category_id"] = categoryID
	}

	if err := config.DB.Model(&models.Document{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates).Error; err != nil {
		return err
	}

	// 更新标签
	if len(tags) > 0 {
		var tagList []models.Tag
		for _, tagName := range tags {
			var tag models.Tag
			result := config.DB.Where("user_id = ? AND name = ?", userID, tagName).First(&tag)
			if result.Error == gorm.ErrRecordNotFound {
				tag = models.Tag{
					UserID: userID,
					Name:  tagName,
					Color: "#1890ff",
				}
				config.DB.Create(&tag)
			}
			tagList = append(tagList, tag)
		}
		config.DB.Model(&models.Document{ID: id}).Association("Tags").Replace(tagList)
	}

	// 同步到搜索索引
	go func() {
		searchSvc := GetSearchService()
		var fullDoc models.Document
		if err := config.DB.Preload("Category").Preload("Tags").First(&fullDoc, id).Error; err == nil {
			searchSvc.IndexDocument(&fullDoc)
		}
	}()

	return nil
}

func (s *DocumentService) Delete(userID, id uint) error {
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Document{}).Error; err != nil {
		return err
	}

	// 从搜索索引删除
	go func() {
		searchSvc := GetSearchService()
		searchSvc.DeleteDocument(id)
	}()

	return nil
}

type CategoryService struct{}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

func (s *CategoryService) Create(userID uint, name string, parentID *uint, icon string) (*models.Category, error) {
	category := &models.Category{
		UserID:    userID,
		Name:     name,
		ParentID: parentID,
		Icon:     icon,
		Status:   1,
	}
	if err := config.DB.Create(category).Error; err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) List(userID uint) ([]models.Category, error) {
	var categories []models.Category
	if err := config.DB.Where("user_id = ?", userID).Order("sort_order ASC, created_at DESC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *CategoryService) GetByID(id, userID uint) (*models.Category, error) {
	var category models.Category
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *CategoryService) Update(userID, id uint, name string, icon string) error {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if icon != "" {
		updates["icon"] = icon
	}
	return config.DB.Model(&models.Category{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates).Error
}

func (s *CategoryService) Delete(userID, id uint) error {
	// 删除子分类
	config.DB.Where("parent_id = ? AND user_id = ?", id, userID).Delete(&models.Category{})
	return config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Category{}).Error
}