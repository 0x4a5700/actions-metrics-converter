# Developer guide

## Repository layout

```
.
├── app/                        Go source
│   ├── main.go                 Entry point: HTTP server setup, signal handling
│   ├── go.mod / go.sum
│   ├── internal/
│   │   ├── config.go           AppConfig (build metadata surfaced at /health)
│   │   ├── apperr/             Typed errors (MissingEnvVar)
│   │   ├── handlers/
│   │   │   ├── webhook.go      Signature verification, payload routing
│   │   │   └── health.go       /health endpoint
│   │   ├── telemetry/
│   │   │   ├── provider.go     OTel TracerProvider init (OTLP HTTP exporter)
│   │   │   ├── span.go         JobSpans / RunSpans value types
│   │   │   └── emit.go         Span creation and export
│   │   └── workflow/
│   │       ├── process_job.go  Maps workflow_job payload → JobSpans
│   │       └── process_run.go  Maps workflow_run payload → RunSpans
│   └── pkg/github/
│       └── github.go           GitHub webhook payload structs
├── charts/actions-metrics-converter/   Helm chart
├── Dockerfile                  Multi-stage build (golang:alpine → scratch)
└── .github/workflows/ci.yml    CI: fmt, vet, test
```

## Prerequisites

- Go 1.25+
- Docker (for container builds)
- Helm 3 (for chart work)

## Running locally

```bash
cd app

export GITHUB_WEBHOOK_SECRET=dev-secret
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

go run .
```

The server starts on `:3018`. If you don't have a local OTLP collector the OTel SDK will log export errors but the server will still run.

A minimal collector for local testing with [Grafana Alloy](https://grafana.com/docs/alloy/) or the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) listening on `4318` is sufficient.

## Tests

```bash
cd app
go test ./...
```

The CI pipeline also checks formatting and runs `go vet`:

```bash
gofmt -l .       # should produce no output
go vet ./...
go test ./...
```

Test data for the GitHub payload parser lives in `pkg/github/testdata/`.

## Trace ID scheme

TraceID and SpanID are derived deterministically from GitHub run/job IDs so that:

- All `workflow_job` events for the same run share a TraceID.
- The `workflow_run` span and all job spans reference the same parent SpanID.

```
TraceID  = first 16 bytes of SHA-256("run:<runID>")
SpanID   = big-endian uint64 encoding of the numeric run or job ID
```

This means replayed or retried webhook deliveries for the same run/job produce identical trace and span IDs. Backends that deduplicate on span ID will merge them; backends that don't will show duplicates.

## Span model

Each processed event produces **two spans**:

| Span name | Start | End |
|---|---|---|
| `<name>: queue` | `created_at` | `started_at` |
| `<name>: run` | `started_at` | `completed_at` |

Both spans share the same attributes. Span status is set to `Error` for conclusions `failure`, `cancelled`, `timed_out`, `action_required`; `Ok` for `success`, `skipped`, `neutral`.

### Attributes on job spans

| Attribute | Source |
|---|---|
| `ci.job.conclusion` | `workflow_job.conclusion` |
| `ci.job.labels` | `workflow_job.labels` (comma-separated) |
| `ci.run.attempt` | `workflow_job.run_attempt` |
| `ci.runner.name` | `workflow_job.runner_name` |
| `ci.runner.group_name` | `workflow_job.runner_group_name` |
| `vcs.repository.full_name` | `repository.full_name` |
| `vcs.ref.head.name` | `workflow_job.head_branch` |
| `vcs.commit.sha` | `workflow_job.head_sha` |

### Attributes on run spans

| Attribute | Source |
|---|---|
| `ci.run.conclusion` | `workflow_run.conclusion` |
| `ci.run.id` | `workflow_run.id` |
| `ci.run.number` | `workflow_run.run_number` |
| `ci.run.attempt` | `workflow_run.run_attempt` |
| `ci.run.event` | `workflow_run.event` |
| `vcs.repository.full_name` | `repository.full_name` |
| `vcs.ref.head.name` | `workflow_run.head_branch` |
| `vcs.commit.sha` | `workflow_run.head_sha` |

## Adding support for new event types

1. Add structs to `pkg/github/github.go` if the payload shape is new.
2. Add a `Process*` function in `internal/workflow/` following the pattern of `process_job.go`.
3. Add an `Emit*` function in `internal/telemetry/emit.go` if the span structure differs.
4. Wire it up in `internal/handlers/webhook.go` inside `handlePayload`.

## Failed payloads

When a payload cannot be unmarshalled, the raw request (headers + body) is written to `FAILED_PAYLOAD_DIR` (default `.`) as a timestamped `.txt` file. These files are useful for updating the GitHub struct definitions when GitHub adds new fields or changes the payload shape.

## Docker build

The `Dockerfile` uses a two-stage build:

1. `golang:alpine` compiles the binary with `CGO_ENABLED=0`.
2. A `scratch` image carries only the binary — no shell, no libc.

Build args `gitHash` and `buildVersion` are embedded via `-ldflags` and surfaced at `/health`.

```bash
docker build \
  --build-arg gitHash=$(git rev-parse HEAD) \
  --build-arg buildVersion=$(git describe --tags --always) \
  -t actions-metrics-converter .
```

## Helm chart development

```bash
# Lint
helm lint charts/actions-metrics-converter

# Render templates locally
helm template my-release charts/actions-metrics-converter \
  --set githubWebhookSecret.value=test \
  --set otel.exporterOtlpEndpoint=http://localhost:4318

# Dry-run install against a cluster
helm install actions-metrics-converter charts/actions-metrics-converter \
  --dry-run \
  --set githubWebhookSecret.value=test \
  --set otel.exporterOtlpEndpoint=http://localhost:4318
```

The ingress template renders a Contour `HTTPProxy`. If your cluster uses a different ingress controller you will need to replace `templates/httpproxy.yaml` with an appropriate `Ingress` or equivalent resource.

## CI

The GitHub Actions workflow (`.github/workflows/ci.yml`) runs on every push to `main` and on pull requests. It checks formatting with `gofmt`, runs `go vet`, runs the test suite, builds the Docker image, and lints/packages the Helm chart. On pushes to `main` the image and chart are published to GHCR.

## Versioning and releases

Versioning is fully automated by [semantic-release](https://semantic-release.gitbook.io/) (configured in `.releaserc.json`) — do not bump versions by hand. On every push to `main`, the release job analyses commit messages, and if a release is warranted it creates a git tag and a GitHub release, then:

- the Docker image is additionally tagged with the new version (alongside `latest` and `sha-*`),
- the Helm chart is packaged with `version` and `appVersion` set to the same number and pushed to GHCR. The values in `Chart.yaml` are placeholders overridden at package time.

Commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/):

- `fix:` → patch release
- `feat:` → minor release
- `feat!:` / `fix!:` or a `BREAKING CHANGE:` footer → major release
- `ci:`, `docs:`, `chore:`, `refactor:`, `test:`, `build:` → no release

Commits that don't trigger a release still publish `latest` and `sha-*` image tags, but the chart is only pushed when a new version is released.
