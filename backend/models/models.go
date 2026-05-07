package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Username    string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email       *string   `gorm:"uniqueIndex;size:100" json:"email"`
	Password    string    `gorm:"not null" json:"-"`
	Nickname    string    `gorm:"size:50" json:"nickname"`
	Avatar      string    `gorm:"size:255" json:"avatar"`
	Role        string    `gorm:"size:20;default:user" json:"role"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	Documents   []Document `gorm:"foreignKey:UserID" json:"-"`
	Categories  []Category `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

type Category struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userID"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	ParentID  *uint     `gorm:"index" json:"parentID"`
	Icon      string    `gorm:"size:50" json:"icon"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	
	User       User      `gorm:"foreignKey:UserID" json:"-"`
	Parent    *Category `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (c *Category) TableName() string {
	return "categories"
}

type Tag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userID"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Color     string    `gorm:"size:20" json:"color"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	
	User      User      `gorm:"foreignKey:UserID" json:"-"`
}

func (t *Tag) TableName() string {
	return "tags"
}

type Document struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"userID"`
	CategoryID *uint     `gorm:"index" json:"categoryID"`
	Title      string    `gorm:"size:255;not null" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	Summary    string    `gorm:"size:500" json:"summary"`
	IsPublic   int       `gorm:"default:0" json:"isPublic"`
	ViewCount  int       `gorm:"default:0" json:"viewCount"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	Category *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag      `gorm:"many2many:document_tags;" json:"tags,omitempty"`
}

func (d *Document) TableName() string {
	return "documents"
}

type DocumentTag struct {
	DocumentID uint `gorm:"primaryKey" json:"documentID"`
	TagID      uint `gorm:"primaryKey" json:"tagID"`
}

func (dt *DocumentTag) TableName() string {
	return "document_tags"
}