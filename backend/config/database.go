package config

import (
	"knowledge-base/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *DatabaseConfig) error {
	var err error
	
	dsn := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + cfg.Port + ")/" + cfg.DBName + "?charset=" + cfg.Charset + "&parseTime=True&loc=Local"
	
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// 尝试创建数据库
		dsnNoDB := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + cfg.Port + ")/?charset=" + cfg.Charset + "&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsnNoDB), &gorm.Config{})
		if err != nil {
			log.Printf("连接数据库失败: %v", err)
			return err
		}
		db.Exec("CREATE DATABASE IF NOT EXISTS " + cfg.DBName + " CHARACTER SET " + cfg.Charset + " COLLATE " + cfg.Charset + "_unicode_ci")
		
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Printf("连接数据库失败: %v", err)
			return err
		}
	}
	
	// 自动迁移表结构
	err = DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Tag{},
		&models.Document{},
		&models.DocumentTag{},
	)
	if err != nil {
		log.Printf("表结构迁移失败: %v", err)
		return err
	}
	
	log.Println("数据库连接成功")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}