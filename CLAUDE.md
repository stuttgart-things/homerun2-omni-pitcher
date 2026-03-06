# CLAUDE.md

## Project

homerun2-omni-pitcher — Go HTTP microservice that accepts JSON messages via `POST /pitch` and enqueues them into Redis Streams using the homerun-library.

## Tech Stack

- **Language**: Go 1.24+
- **HTTP**: stdlib `net/http` (no framework)
- **Queue**: Redis Streams via `homerun-library`
- **Build**: ko (`.ko.yaml`), no Dockerfile
- **CI**: Dagger modules (`dagger/main.go`), Taskfile
- **Deploy**: KCL manifests (`kcl/`), Kustomize, Kubernetes
- **Infra**: GitHub Actions, semantic-release, renovate

## Git Workflow

**Branch-per-issue with PR and merge.** Every change gets its own branch, PR, and merge to main.

### Branch naming

- `fix/<issue-number>-<short-description>` for bugs
- `feat/<issue-number>-<short-description>` for features
- `test/<issue-number>-<short-description>` for test-only changes

### Workflow

1. Branch off `main`: `git checkout -b fix/<N>-<desc> main`
2. Make changes, commit with `Closes #<N>` in the message
3. Push: `git push -u origin <branch>`
4. Create PR: `gh pr create --base main`
5. Merge: `gh pr merge <N> --merge --delete-branch`
6. If multiple issues are closely related (e.g., same file), combine into one branch with multiple `Closes #N`

### Commit messages

- Use conventional commits: `fix:`, `feat:`, `test:`, `chore:`, `docs:`
- End with `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>` when Claude authored
- Include `Closes #<issue-number>` to auto-close issues

## Code Conventions

- No Dockerfile — use ko for image builds
- Config via environment variables, loaded once at startup
- Auth via Bearer token middleware (`AUTH_TOKEN` env var)
- Tests: `go test ./...` — unit tests must not require Redis; integration tests run via Dagger with Redis service
- KCL is the source of truth for Kubernetes manifests (not hand-written YAML)

## Key Paths

- `main.go` — entrypoint, routing, graceful shutdown
- `internal/handlers/` — HTTP handlers (pitch, health)
- `internal/middleware/` — auth middleware
- `internal/config/` — env-based config loading
- `internal/models/` — response structs
- `dagger/main.go` — CI functions (Lint, Build, BuildImage, ScanImage, BuildAndTestBinary)
- `kcl/` — Kubernetes manifests (modular: schema.k, labels.k, deploy.k, service.k, ingress.k, secret.k, configmap.k, serviceaccount.k, namespace.k)
- `tests/kcl-deploy-profile.yaml` — KCL deployment profile for parameterized rendering
- `Taskfile.yaml` — task runner for build/test/deploy/release
- `.ko.yaml` — ko build configuration
- `.github/workflows/` — CI/CD (build-test, build-scan-image, release, lint-repo)

## Testing

```bash
# Unit tests (no Redis needed)
go test ./...

# Integration test via Dagger (spins up Redis)
task build-test-binary

# Lint
task lint

# Build + scan image
task build-scan-image-ko

# Render KCL manifests
task render-manifests-quick
```

## Reference Project

`claim-machinery-api` is the reference for infra patterns (modular KCL, kustomize OCI pipeline, Dagger functions, GitHub Actions).
