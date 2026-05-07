package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

// 全局配置实例
var AppConfig *Config
var loadOnce sync.Once

// LoadConfig 加载配置（支持多种方式）
// 优先级：命令行参数 > 环境变量 > .env 文件 > 默认值
func LoadConfig() *Config {
	loadOnce.Do(func() {
		// 1. 解析命令行参数
		envFile := flag.String("env", ".env", "环境配置文件路径")
		env := flag.String("e", "", "环境名称 (dev/test/prod)，自动加载 .env.{env}")
		showConfig := flag.Bool("show-config", false, "显示当前配置（调试用）")
		flag.Parse()

		// 2. 加载 .env 文件
		loadEnvFile(*envFile, *env)

		// 3. 构建配置
		AppConfig = &Config{
			Server: ServerConfig{
				Port: getEnv("SERVER_PORT", "8080"),
				Mode: getEnv("GIN_MODE", "debug"),
			},
			Database: DatabaseConfig{
				Driver:   getEnv("DB_DRIVER", "mysql"),
				Host:     getEnv("DB_HOST", "localhost"),
				Port:     getEnv("DB_PORT", "3306"),
				User:     getEnv("DB_USER", "root"),
				Password: getEnv("DB_PASSWORD", ""),
				DBName:   getEnv("DB_NAME", "knowledge_base"),
				Charset:  getEnv("DB_CHARSET", "utf8mb4"),
			},
			JWT: JWTConfig{
				Secret:     getEnv("JWT_SECRET", "change-this-secret-key"),
				ExpireHour: 24 * 7,
			},
		}

		// 4. 调试模式下显示配置
		if *showConfig {
			printConfig()
		}
	})

	return AppConfig
}

// loadEnvFile 加载环境配置文件
func loadEnvFile(envFile string, env string) {
	// 优先加载指定环境的配置文件
	if env != "" {
		envSpecificFile := fmt.Sprintf(".env.%s", env)
		if fileExists(envSpecificFile) {
			if err := godotenv.Load(envSpecificFile); err != nil {
				log.Printf("警告: 加载 %s 失败: %v", envSpecificFile, err)
			} else {
				log.Printf("已加载配置: %s", envSpecificFile)
			}
		}
	}

	// 加载主配置文件（不覆盖已存在的环境变量）
	if fileExists(envFile) {
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("警告: 加载 %s 失败: %v", envFile, err)
		} else {
			log.Printf("已加载配置: %s", envFile)
		}
	}

	// 加载本地覆盖配置（用于开发时的个性化配置，不提交到 git）
	if fileExists(".env.local") {
		if err := godotenv.Overload(".env.local"); err != nil {
			log.Printf("警告: 加载 .env.local 失败: %v", err)
		} else {
			log.Printf("已加载配置: .env.local (覆盖模式)")
		}
	}
}

func (c *DatabaseConfig) DSN() string {
	charset := c.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName, charset)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// printConfig 打印当前配置（隐藏敏感信息）
func printConfig() {
	fmt.Println("\n========== 当前配置 ==========")
	fmt.Printf("服务器端口: %s\n", AppConfig.Server.Port)
	fmt.Printf("运行模式: %s\n", AppConfig.Server.Mode)
	fmt.Printf("数据库地址: %s:%s\n", AppConfig.Database.Host, AppConfig.Database.Port)
	fmt.Printf("数据库用户: %s\n", AppConfig.Database.User)
	fmt.Printf("数据库名: %s\n", AppConfig.Database.DBName)
	fmt.Printf("JWT密钥: %s***\n", maskSecret(AppConfig.JWT.Secret))
	fmt.Println("==============================\n")
}

// maskSecret 隐藏密钥中间部分
func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
