package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host             string
	Port             string
	RegistryPath     string
	StaticAssetsPath string
}

func Load() (Config, error) {
	cfg := Config{
		Host:             envOr("HOST", "127.0.0.1"),
		Port:             envOr("PORT", "8080"),
		RegistryPath:     envOr("WEBSITES_REGISTRY_PATH", "/home/stuserver/data/websites.json"),
		StaticAssetsPath: envOr("STATIC_ASSETS_PATH", "static"),
	}

	if strings.TrimSpace(cfg.Host) == "" || strings.ContainsAny(cfg.Host, " \t\r\n") {
		return Config{}, fmt.Errorf("HOST must be a non-empty hostname or IP address")
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be an integer from 1 to 65535")
	}
	if strings.TrimSpace(cfg.RegistryPath) == "" {
		return Config{}, fmt.Errorf("WEBSITES_REGISTRY_PATH must not be empty")
	}
	if strings.TrimSpace(cfg.StaticAssetsPath) == "" {
		return Config{}, fmt.Errorf("STATIC_ASSETS_PATH must not be empty")
	}
	cfg.Host = strings.TrimPrefix(strings.TrimSuffix(cfg.Host, "]"), "[")
	if strings.Contains(cfg.Host, ":") && net.ParseIP(cfg.Host) == nil {
		return Config{}, fmt.Errorf("HOST must not include a port")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
