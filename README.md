# pr-gitops-grafana-annotation

Small webhook receiver that listens for merged pull requests on a GitHub GitOps
repository (default: `qjoly/gitops`) and creates a Grafana annotation at the
exact merge time, so the effect of each GitOps change is visible on dashboards.

## How it works

1. GitHub sends a `pull_request` webhook event to `POST /webhook/github`.
2. The request's HMAC-SHA256 signature (`X-Hub-Signature-256`) is verified
   against `GITHUB_WEBHOOK_SECRET`.
3. If the event is a merge (`action == "closed"` and `pull_request.merged ==
   true`) on the configured repository, a global annotation is posted to
   Grafana via `POST /api/annotations`, timestamped at `merged_at`.

## Configuration (environment variables)

| Variable               | Required | Default          | Description                                  |
|-------------------------|----------|------------------|-----------------------------------------------|
| `LISTEN_ADDR`            | no       | `:8080`          | HTTP listen address                           |
| `GITHUB_WEBHOOK_SECRET`  | yes      | -                | Shared secret configured on the GitHub webhook |
| `GITHUB_REPO`            | no       | `qjoly/gitops`   | `owner/name` of the repo to watch             |
| `GRAFANA_URL`            | yes      | -                | Base URL of the Grafana instance              |
| `GRAFANA_TOKEN`          | yes      | -                | Grafana service account token (needs annotation write permission) |
| `ANNOTATION_TAGS`        | no       | `gitops,deploy`  | Comma-separated tags applied to the annotation |

## Local development

```bash
export GITHUB_WEBHOOK_SECRET=dev-secret
export GRAFANA_URL=http://localhost:3000
export GRAFANA_TOKEN=glsa_xxx
go run ./cmd/pr-gitops-grafana-annotation
```

Run tests:

```bash
go test ./...
```

## Building the image

```bash
docker build -t ghcr.io/qjoly/pr-gitops-grafana-annotation:latest .
```

CI (`.github/workflows/build.yml`) runs `go vet`/`go test` on every push and
PR, and builds+pushes the image to `ghcr.io/qjoly/pr-gitops-grafana-annotation`
on pushes to `main` and version tags.

## Deployment

This app is deployed into the `mocha` Kubernetes cluster through the
[`qjoly/gitops`](https://github.com/qjoly/gitops) repository (ArgoCD +
`ApplicationSet` git-directory generator over `mocha/apps/*`). Manifests
already added there, mirroring the existing `rmfakecloud` app:

- `mocha/apps/pr-gitops-grafana-annotation/app.yaml` — ArgoCD `Application`
- `mocha/apps/pr-gitops-grafana-annotation/manifests/deployment.yaml` — Deployment + Service
- `mocha/apps/pr-gitops-grafana-annotation/manifests/ingress.yaml` — Traefik Ingress, TLS via `cloudflare` ClusterIssuer, host `annotator.mocha.thoughtless.eu`
- `mocha/apps/pr-gitops-grafana-annotation/manifests/externalsecret.yaml` — pulls `github_webhook_secret` and `grafana_token` from Vault at path `pr-gitops-grafana-annotation`

**Before this can go live, two things need to happen manually** (not
automatable from here):

1. Add the following keys to Vault at path `pr-gitops-grafana-annotation` (see
   `docs/secret-management.md` in the gitops repo for the exact ESO/Vault
   workflow used elsewhere):
   - `github_webhook_secret` — a random secret string, must match the webhook secret configured in step 2.
   - `grafana_token` — a Grafana service account token with the `Annotations:Write` (or
     Editor) permission. Create it under Grafana → Administration → Service
     accounts.
2. On the `qjoly/gitops` GitHub repository, add a webhook:
   - Payload URL: `https://annotator.mocha.thoughtless.eu/webhook/github`
   - Content type: `application/json`
   - Secret: same value as `github_webhook_secret` above
   - Events: only the **Pull requests** event

Then commit/push the new `mocha/apps/pr-gitops-grafana-annotation/` directory
in the gitops repo — the `mocha-apps` ArgoCD ApplicationSet will pick it up
automatically.
