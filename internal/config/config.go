package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 应用程序配置
type Config struct {
	// 服务器配置
	ListenAddr  string
	TLSCertFile string
	TLSKeyFile  string
	BaseURL     string
	DataDir     string

	// 认证配置
	AuthDisable      bool
	AuthSecureCookie bool

	// SMTP 配置
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// 安全配置
	EnableRateLimit bool
	RateLimitRPS    int // 每秒请求数

	// JWT 配置
	JWTSecret     string
	JWTExpiration time.Duration

	// CORS 配置
	CORSAllowedOrigins []string
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	cfg := &Config{
		// 服务器配置
		ListenAddr:  getEnv("LISTEN_ADDR", ":8088"),
		TLSCertFile: strings.TrimSpace(getEnv("TLS_CERT_FILE", "")),
		TLSKeyFile:  strings.TrimSpace(getEnv("TLS_KEY_FILE", "")),
		BaseURL:     strings.TrimSpace(getEnv("BASE_URL", "")),
		DataDir:     getEnv("DATA_DIR", "./data"),

		// 认证配置
		AuthDisable:      getEnvBool("AUTH_DISABLE", false),
		AuthSecureCookie: getEnvBool("AUTH_SECURE_COOKIE", false),

		// SMTP 配置
		SMTPHost:     strings.TrimSpace(getEnv("SMTP_HOST", "")),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     strings.TrimSpace(getEnv("SMTP_USER", "")),
		SMTPPassword: strings.TrimSpace(getEnv("SMTP_PASSWORD", "")),
		SMTPFrom:     strings.TrimSpace(getEnv("SMTP_FROM", "")),

		// 安全配置
		EnableRateLimit: getEnvBool("ENABLE_RATE_LIMIT", false),
		RateLimitRPS:    getEnvInt("RATE_LIMIT_RPS", 10),

		// JWT 配置
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),

		// CORS 配置
		CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 如果配置了 TLS，自动启用安全 Cookie
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		cfg.AuthSecureCookie = true
	}

	return cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证 TLS 配置
	if (c.TLSCertFile != "" && c.TLSKeyFile == "") || (c.TLSCertFile == "" && c.TLSKeyFile != "") {
		return fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE must both be set or both be empty")
	}

	// 验证 JWT 配置
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// 验证 SMTP 配置（如果启用）
	if c.SMTPHost != "" {
		if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
			return fmt.Errorf("invalid SMTP_PORT: %d", c.SMTPPort)
		}
	}

	// 验证速率限制配置
	if c.EnableRateLimit && c.RateLimitRPS <= 0 {
		return fmt.Errorf("RATE_LIMIT_RPS must be positive when rate limiting is enabled")
	}

	return nil
}

// IsTLSEnabled 检查是否启用 TLS
func (c *Config) IsTLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

// IsSMTPEnabled 检查是否启用 SMTP
func (c *Config) IsSMTPEnabled() bool {
	return c.SMTPHost != ""
}

// GetListenAddress 获取监听地址
func (c *Config) GetListenAddress() string {
	return c.ListenAddr
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
