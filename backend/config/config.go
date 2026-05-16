package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AI       AIConfig
	Search   SearchConfig
}

type SearchConfig struct {
	Host   string // Meilisearch 服务地址
	APIKey string // API Key（可选）
	Index  string // 索引名称
}

type AIConfig struct {
	// 默认 AI 提供商: tongyi / openai / deepseek / zhipu / ollama
	DefaultProvider string

	// 通义千问 API
	DashScopeAPIKey string
	ChatModel       string // 对话模型: qwen-turbo, qwen-plus, qwen-max

	// OpenAI API
	OpenAIAPIKey string
	OpenAIModel  string
	OpenAIBaseURL string // 自定义端点（可选）

	// DeepSeek API
	DeepSeekAPIKey string
	DeepSeekModel  string

	// 智谱 API
	ZhipuAPIKey string
	ZhipuModel  string

	// Ollama 本地模型
	OllamaBaseURL string
	OllamaModel   string

	// 通用配置
	EmbeddingModel  string // 嵌入模型: text-embedding-v3
	MaxTokens       int    // 最大输出 token
	Temperature     float64 // 温度参数

	// RAG 配置
	ChunkSize           int     // 文档切片大小（字符）
	ChunkOverlap        int     // 切片重叠大小
	TopK                int     // 检索返回文档数
	MinScore            float64 // 最小相似度阈值
	EmbeddingDimensions int     // 嵌入向量维度（text-embedding-v3 = 1024）
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
			AI: AIConfig{
				DefaultProvider:  getEnv("AI_DEFAULT_PROVIDER", "tongyi"),
				DashScopeAPIKey: getEnv("DASHSCOPE_API_KEY", ""),
				ChatModel:       getEnv("AI_CHAT_MODEL", "qwen-turbo"),
				OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
				OpenAIModel:    getEnv("OPENAI_MODEL", "gpt-4o-mini"),
				OpenAIBaseURL:  getEnv("OPENAI_BASE_URL", ""),
				DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),
				DeepSeekModel:  getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
				ZhipuAPIKey:    getEnv("ZHIPU_API_KEY", ""),
				ZhipuModel:     getEnv("ZHIPU_MODEL", "glm-4-flash"),
				OllamaBaseURL:  getEnv("OLLAMA_BASE_URL", ""),
				OllamaModel:    getEnv("OLLAMA_MODEL", "llama3"),
				EmbeddingModel:  getEnv("AI_EMBEDDING_MODEL", "text-embedding-v3"),
				MaxTokens:       2000,
				Temperature:     0.7,
				ChunkSize:           500,
				ChunkOverlap:        50,
				TopK:                5,
				MinScore:            getEnvFloat("AI_MIN_SCORE", 0.3),
				EmbeddingDimensions: 1024,
			},
			Search: SearchConfig{
				Host:   getEnv("MEILISEARCH_HOST", "http://localhost:7700"),
				APIKey: getEnv("MEILISEARCH_API_KEY", ""),
				Index:  getEnv("MEILISEARCH_INDEX", "documents"),
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

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
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
	if AppConfig.AI.DashScopeAPIKey != "" {
		fmt.Printf("AI API Key: %s***\n", maskSecret(AppConfig.AI.DashScopeAPIKey))
	}
	fmt.Printf("嵌入模型: %s\n", AppConfig.AI.EmbeddingModel)
	fmt.Printf("对话模型: %s\n", AppConfig.AI.ChatModel)
	fmt.Println("==============================")
}

// maskSecret 隐藏密钥中间部分
func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
