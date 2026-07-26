package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/qjoly/pr-gitops-grafana-annotation/internal/config"
	"github.com/qjoly/pr-gitops-grafana-annotation/internal/grafana"
	"github.com/qjoly/pr-gitops-grafana-annotation/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	grafanaClient := grafana.NewClient(cfg.GrafanaURL, cfg.GrafanaToken)
	srv := server.New(cfg, grafanaClient, logger)

	logger.Info("starting server", "addr", cfg.ListenAddr, "repo", cfg.GitHubRepo, "grafana_url", cfg.GrafanaURL)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Routes()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
