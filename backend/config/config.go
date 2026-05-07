package config

import (
	"fmt"
	"os"
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
	Host    string
	Port    string
	User    string
	Password string
	DBName  string
	Charset string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "mysql"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3307"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "j4PNMPGi52RAkDP2"),
			DBName:   getEnv("DB_NAME", "knowledge_base"),
			Charset: getEnv("DB_CHARSET", "utf8mb4"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "knowledge-base-secret-key-2024"),
			ExpireHour: 24 * 7, // 7 days
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.Charset)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}