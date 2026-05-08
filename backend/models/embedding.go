package models

import (
	"encoding/json"
	"time"
)

// DocumentChunk 文档切片（用于向量存储）
type DocumentChunk struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	DocumentID uint      `gorm:"index;not null" json:"documentID"`
	UserID     uint      `gorm:"index;not null" json:"userID"`
	ChunkIndex int       `gorm:"not null" json:"chunkIndex"`      // 切片序号
	Content    string    `gorm:"type:text;not null" json:"content"` // 切片内容
	Embedding  []byte    `gorm:"type:blob" json:"-"`             // 向量数据（JSON 序列化）
	CreatedAt  time.Time `json:"createdAt"`
}

func (dc *DocumentChunk) TableName() string {
	return "document_chunks"
}

// SetEmbedding 设置向量（序列化为 JSON）
func (dc *DocumentChunk) SetEmbedding(vec []float64) error {
	data, err := json.Marshal(vec)
	if err != nil {
		return err
	}
	dc.Embedding = data
	return nil
}

// GetEmbedding 获取向量（反序列化）
func (dc *DocumentChunk) GetEmbedding() ([]float64, error) {
	if len(dc.Embedding) == 0 {
		return nil, nil
	}
	var vec []float64
	err := json.Unmarshal(dc.Embedding, &vec)
	return vec, err
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userID"`
	SessionID string    `gorm:"index;size:64" json:"sessionID"`  // 会话 ID
	Role      string    `gorm:"size:20;not null" json:"role"`   // user / assistant
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func (cm *ChatMessage) TableName() string {
	return "chat_messages"
}

// ChatSession 对话会话
type ChatSession struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userID"`
	SessionID string    `gorm:"uniqueIndex;size:64;not null" json:"sessionID"`
	Title     string    `gorm:"size:255" json:"title"`           // 会话标题（首条消息摘要）
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (cs *ChatSession) TableName() string {
	return "chat_sessions"
}
