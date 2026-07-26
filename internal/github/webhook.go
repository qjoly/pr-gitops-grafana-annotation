// Package github verifies and decodes GitHub "pull_request" webhook payloads.
package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrInvalidSignature = errors.New("invalid webhook signature")

// PullRequestEvent is the subset of GitHub's pull_request webhook payload we care about.
type PullRequestEvent struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Title    string     `json:"title"`
		HTMLURL  string     `json:"html_url"`
		Merged   bool       `json:"merged"`
		MergedAt *time.Time `json:"merged_at"`
		User     struct {
			Login string `json:"login"`
		} `json:"user"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// IsMerge reports whether this event represents a PR merge.
func (e *PullRequestEvent) IsMerge() bool {
	return e.Action == "closed" && e.PullRequest.Merged
}

// VerifySignature checks the X-Hub-Signature-256 header against the payload using the
// shared webhook secret, per GitHub's HMAC-SHA256 scheme.
func VerifySignature(secret string, payload []byte, signatureHeader string) error {
	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return ErrInvalidSignature
	}
	expectedHex := strings.TrimPrefix(signatureHeader, prefix)
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return ErrInvalidSignature
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	computed := mac.Sum(nil)

	if !hmac.Equal(computed, expected) {
		return ErrInvalidSignature
	}
	return nil
}

// ParsePullRequestEvent decodes a pull_request webhook payload.
func ParsePullRequestEvent(payload []byte) (*PullRequestEvent, error) {
	var evt PullRequestEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
