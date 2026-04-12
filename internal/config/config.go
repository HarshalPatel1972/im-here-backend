package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GitHubToken         string
	DatabaseURL         string
	SMTPUser            string
	SMTPPassword        string
	SMTPHost            string
	SMTPPort            string
	PollIntervalSeconds int
	LogLevel            string
}

func Load() *Config {
	cfg := &Config{
		GitHubToken:         requireEnv("GITHUB_TOKEN"),
		DatabaseURL:         requireEnv("DATABASE_URL"),
		SMTPUser:            requireEnv("SMTP_USER"),
		SMTPPassword:        requireEnv("SMTP_PASSWORD"),
		SMTPHost:            getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:            getEnv("SMTP_PORT", "587"),
		PollIntervalSeconds: requireEnvInt("POLL_INTERVAL_SECONDS"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
	}
	return cfg
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("I'm Here requires %s to run. Please set it in your environment or .env file.", key))
	}
	return val
}

func requireEnvInt(key string) int {
	val := requireEnv(key)
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("I'm Here requires %s to be an integer. Got: %s", key, val))
	}
	return i
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
