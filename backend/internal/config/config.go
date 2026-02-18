package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort         string `mapstructure:"SERVER_PORT"`
	ServerHost         string `mapstructure:"SERVER_HOST"`
	ServerIP           string `mapstructure:"SERVER_IP"`
	DatabaseURL        string `mapstructure:"DATABASE_URL"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTExpiration      string `mapstructure:"JWT_EXPIRATION"`
	DockerHost         string `mapstructure:"DOCKER_HOST"`
	CloudflareAPIToken string `mapstructure:"CLOUDFLARE_API_TOKEN"`
	CloudflareEmail    string `mapstructure:"CLOUDFLARE_EMAIL"`
	TraefikEmail       string `mapstructure:"TRAEFIK_EMAIL"`
	BackupBasePath     string `mapstructure:"BACKUP_BASE_PATH"`
	WordPressImage     string `mapstructure:"WORDPRESS_IMAGE"`
	DNSResolver        string `mapstructure:"DNS_RESOLVER"`
}

var AppConfig *Config

func LoadConfig() *Config {
	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		ServerHost:         getEnv("SERVER_HOST", "0.0.0.0"),
		ServerIP:           getEnv("SERVER_IP", "127.0.0.1"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://wpplatform_user:wpplatform_pass@localhost:5432/wplatform_db?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpiration:      getEnv("JWT_EXPIRATION", "24h"),
		DockerHost:         getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		CloudflareAPIToken: getEnv("CLOUDFLARE_API_TOKEN", ""),
		CloudflareEmail:    getEnv("CLOUDFLARE_EMAIL", ""),
		TraefikEmail:       getEnv("TRAEFIK_EMAIL", "admin@wp-platform.com"),
		BackupBasePath:     getEnv("BACKUP_BASE_PATH", "/var/backups/wp"),
		WordPressImage:     getEnv("WORDPRESS_IMAGE", "wordpress:6-php8.2-apache"),
		DNSResolver:        getEnv("DNS_RESOLVER", "8.8.8.8"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func Init() {
	AppConfig = LoadConfig()
	log.Printf("Configuration loaded: Server=%s:%s", AppConfig.ServerHost, AppConfig.ServerPort)
}
