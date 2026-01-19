package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	S3       S3Config
}

type ServerConfig struct {
	Port         string
	Environment  string
	AllowOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret       string
	RefreshSecret      string
	AccessExpiration   time.Duration
	RefreshExpiration  time.Duration
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	PublicURL       string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			AllowOrigins: []string{getEnv("CORS_ORIGIN", "http://localhost:5173")},
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "auction"),
			Password: getEnv("DB_PASSWORD", "auction123"),
			DBName:   getEnv("DB_NAME", "auction_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-super-secret-access-key-change-in-production"),
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-in-production"),
			AccessExpiration:   time.Duration(getEnvInt("JWT_ACCESS_EXPIRATION_MINUTES", 15)) * time.Minute,
			RefreshExpiration:  time.Duration(getEnvInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		},
		S3: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("S3_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("S3_SECRET_KEY", "minioadmin123"),
			BucketName:      getEnv("S3_BUCKET", "auction-images"),
			UseSSL:          getEnvBool("S3_USE_SSL", false),
			PublicURL:       getEnv("S3_PUBLIC_URL", ""),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

func (c *RedisConfig) Addr() string {
	return c.Host + ":" + c.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
