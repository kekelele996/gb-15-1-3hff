package config

import (
	"os"
	"strconv"
)

// Config 集中解析全部后端环境变量配置。
type Config struct {
	ServerPort          string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	RedisAddr           string
	JWTSecret           string
	JWTExpireHours      int
	MinIOEndpoint       string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
	MinIOUseSSL         bool
	LogLevel            string
	SimilarityThreshold float64
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// Load 从环境变量加载配置，所有配置均提供默认值，保证本地开发可直接运行。
func Load() *Config {
	return &Config{
		ServerPort:          getenv("SERVER_PORT", "8080"),
		DBHost:              getenv("DB_HOST", "localhost"),
		DBPort:              getenv("DB_PORT", "5432"),
		DBUser:              getenv("DB_USER", "paperflow_user"),
		DBPassword:          getenv("DB_PASSWORD", "paperflow_pwd"),
		DBName:              getenv("DB_NAME", "paperflow_db"),
		RedisAddr:           getenv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:           getenv("JWT_SECRET", "change_me_to_a_long_random_string"),
		JWTExpireHours:      envInt("JWT_EXPIRE_HOURS", 72),
		MinIOEndpoint:       getenv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:      getenv("MINIO_ROOT_USER", "minioadmin"),
		MinIOSecretKey:      getenv("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinIOBucket:         getenv("MINIO_BUCKET", "paperflow-files"),
		MinIOUseSSL:         getenv("MINIO_USE_SSL", "false") == "true",
		LogLevel:            getenv("LOG_LEVEL", "info"),
		SimilarityThreshold: envFloat("SIMILARITY_THRESHOLD", 30.0),
	}
}
