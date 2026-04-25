package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DBPath        string
	CaddyAdminURL string
	PublicURL     string
	ContainerPort string
	DockerNetwork string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "./data/brimble.db"),
		CaddyAdminURL: getEnv("CADDY_ADMIN_URL", "http://caddy:2019"),
		PublicURL:     getEnv("PUBLIC_URL", "http://localhost"),
		ContainerPort: getEnv("CONTAINER_PORT", "3000"),
		DockerNetwork: getEnv("DOCKER_NETWORK", "mini-brimble_brimble"),
	}
}

func (c *Config) LiveURL(deploymentID string) string {
	return fmt.Sprintf("%s/deploys/%s/", c.PublicURL, deploymentID)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
