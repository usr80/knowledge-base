package main

import (
	"fmt"
	"strings"

	"knowledge-base/config"
	"knowledge-base/models"
)

func main() {
	cfg := config.LoadConfig()
	if err := config.InitDB(&cfg.Database); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		return
	}
	db := config.DB

	// 查询切片数量
	var chunkCount int64
	db.Model(&models.DocumentChunk{}).Count(&chunkCount)
	fmt.Printf("总切片数: %d\n", chunkCount)

	// 查询文档数量
	var docCount int64
	db.Model(&models.Document{}).Count(&docCount)
	fmt.Printf("总文档数: %d\n", docCount)

	// 查询每个文档的切片数
	type DocInfo struct {
		DocumentID uint
		Title      string
		ChunkCount int
		HasEmbed   int
	}
	var docInfos []DocInfo
	db.Raw(`
		SELECT dc.document_id, d.title, COUNT(*) as chunk_count,
			SUM(CASE WHEN dc.embedding IS NOT NULL AND dc.embedding != '' THEN 1 ELSE 0 END) as has_embed
		FROM document_chunks dc
		JOIN documents d ON d.id = dc.document_id
		GROUP BY dc.document_id, d.title
	`).Scan(&docInfos)

	fmt.Println("\n文档切片详情:")
	fmt.Println(strings.Repeat("-", 70))
	for _, info := range docInfos {
		fmt.Printf("  文档ID=%d | 切片=%d | 有向量=%d | %s\n",
			info.DocumentID, info.ChunkCount, info.HasEmbed, info.Title)
	}

	// 测试一个切片的 embedding 是否有效
	var sample models.DocumentChunk
	if err := db.Where("embedding IS NOT NULL AND embedding != ''").First(&sample).Error; err != nil {
		fmt.Printf("\n没有有效的 embedding 数据！")
		return
	}

	vec, err := sample.GetEmbedding()
	if err != nil {
		fmt.Printf("\n切片 %d embedding 解析失败: %v\n", sample.ID, err)
		return
	}
	fmt.Printf("\n样本切片 ID=%d, 文档ID=%d, 向量维度=%d\n", sample.ID, sample.DocumentID, len(vec))
	fmt.Printf("内容前100字: %s\n", truncate(sample.Content, 100))
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
