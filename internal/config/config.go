// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ListenAddr        string
	GitHubWebhookSecret string
	GitHubRepo        string // "owner/name", e.g. "qjoly/gitops"
	GrafanaURL        string
	GrafanaToken      string
	AnnotationTags    []string
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:          getEnv("LISTEN_ADDR", ":8080"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		GitHubRepo:          getEnv("GITHUB_REPO", "qjoly/gitops"),
		GrafanaURL:          strings.TrimRight(os.Getenv("GRAFANA_URL"), "/"),
		GrafanaToken:        os.Getenv("GRAFANA_TOKEN"),
		AnnotationTags:      splitAndTrim(getEnv("ANNOTATION_TAGS", "gitops,deploy")),
	}

	if cfg.GitHubWebhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}
	if cfg.GrafanaURL == "" {
		return nil, fmt.Errorf("GRAFANA_URL is required")
	}
	if cfg.GrafanaToken == "" {
		return nil, fmt.Errorf("GRAFANA_TOKEN is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
