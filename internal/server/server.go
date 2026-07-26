// Package server implements the HTTP webhook receiver.
package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/qjoly/pr-gitops-grafana-annotation/internal/config"
	"github.com/qjoly/pr-gitops-grafana-annotation/internal/github"
	"github.com/qjoly/pr-gitops-grafana-annotation/internal/grafana"
)

type Server struct {
	cfg      *config.Config
	grafana  *grafana.Client
	logger   *slog.Logger
}

func New(cfg *config.Config, grafanaClient *grafana.Client, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, grafana: grafanaClient, logger: logger}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/webhook/github", s.handleGitHubWebhook)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 5<<20)) // 5 MiB cap
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	if err := github.VerifySignature(s.cfg.GitHubWebhookSecret, body, r.Header.Get("X-Hub-Signature-256")); err != nil {
		s.logger.Warn("rejected webhook: invalid signature", "remote_addr", r.RemoteAddr)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("X-GitHub-Event") != "pull_request" {
		w.WriteHeader(http.StatusOK)
		return
	}

	evt, err := github.ParsePullRequestEvent(body)
	if err != nil {
		s.logger.Error("failed to parse pull_request payload", "error", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if evt.Repository.FullName != s.cfg.GitHubRepo {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !evt.IsMerge() {
		w.WriteHeader(http.StatusOK)
		return
	}

	mergedAt := time.Now().UTC()
	if evt.PullRequest.MergedAt != nil {
		mergedAt = evt.PullRequest.MergedAt.UTC()
	}

	text := fmt.Sprintf("PR #%d merged: %s (by %s) -> %s", evt.Number, evt.PullRequest.Title, evt.PullRequest.User.Login, evt.PullRequest.HTMLURL)

	ctx := r.Context()
	if err := s.grafana.CreateAnnotation(ctx, mergedAt, text, s.cfg.AnnotationTags); err != nil {
		s.logger.Error("failed to create grafana annotation", "error", err, "pr", evt.Number)
		http.Error(w, "failed to create annotation", http.StatusBadGateway)
		return
	}

	s.logger.Info("created grafana annotation for merged PR", "pr", evt.Number, "repo", evt.Repository.FullName, "merged_at", mergedAt)
	w.WriteHeader(http.StatusOK)
}
