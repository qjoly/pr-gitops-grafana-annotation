package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	secret := "top-secret"
	payload := []byte(`{"hello":"world"}`)

	if err := VerifySignature(secret, payload, sign(secret, payload)); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}

	if err := VerifySignature(secret, payload, sign("wrong-secret", payload)); err == nil {
		t.Fatal("expected error for mismatched secret")
	}

	if err := VerifySignature(secret, payload, "not-a-valid-header"); err == nil {
		t.Fatal("expected error for malformed header")
	}
}

func TestIsMerge(t *testing.T) {
	evt := &PullRequestEvent{Action: "closed"}
	evt.PullRequest.Merged = true
	if !evt.IsMerge() {
		t.Fatal("expected IsMerge to be true")
	}

	evt.PullRequest.Merged = false
	if evt.IsMerge() {
		t.Fatal("expected IsMerge to be false when not merged")
	}
}

func TestParsePullRequestEvent(t *testing.T) {
	payload := []byte(`{
		"action": "closed",
		"number": 42,
		"pull_request": {
			"title": "Bump image tag",
			"html_url": "https://github.com/qjoly/gitops/pull/42",
			"merged": true,
			"merged_at": "2026-07-26T12:00:00Z",
			"user": {"login": "qjoly"},
			"base": {"ref": "main"}
		},
		"repository": {"full_name": "qjoly/gitops"}
	}`)

	evt, err := ParsePullRequestEvent(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.Number != 42 || !evt.IsMerge() || evt.Repository.FullName != "qjoly/gitops" {
		t.Fatalf("unexpected parsed event: %+v", evt)
	}
}
