# CI/CD

## GitHub Actions Workflows

### Core

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `build-test.yaml` | PR / push to main | Dagger lint + govulncheck (report-only — see [#158](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/158)) + build + test |
| `build-scan-image.yaml` | PR / push to main | ko build, then Trivy image scan (`HIGH,CRITICAL`, report-only — see [#158](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/158)). Both jobs push to `ghcr.io`: the PR job tags `pr-<num>-<sha>` + `pr-<num>` for preview envs, the main job tags `:main` |
| `release.yaml` | After image build / manual | Semantic release + stage image + push kustomize OCI |
| `pages.yaml` | After release / manual | Deploy MkDocs to GitHub Pages |
| `lint-repo.yaml` | PR / push to main | Repository linting |

### PR-preview env

These four together drive the per-PR ephemeral preview environment on `homerun2-dev` for PRs carrying the `preview` label.

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `build-scan-image.yaml` (PR job) | PR opened/updated | ko image tagged `pr-<num>-<sha>` + `pr-<num>` consumed by the per-PR ArgoCD Application |
| `push-kustomize-pr.yaml` | PR opened/updated | Kustomize OCI tagged `pr-<num>-<sha>` (renders `kcl/main.k` against `tests/kcl-deploy-profile.yaml`) |
| `comment-preview-url.yaml` | PR opened/reopened | Sticky bot comment with the preview URL, namespace, and ArgoCD link |
| `cleanup-pr-artifacts.yaml` | PR closed | Deletes both ghcr.io packages so version histories don't fill with PR debris |

See [Preview Environments](preview-environments.md) for the full flow, AppSet anatomy, and troubleshooting.

## Dagger Functions

The `dagger/` module provides:

| Function | Description |
|----------|-------------|
| `Lint` | Go linting via golangci-lint |
| `Build` | Build Go binary |
| `BuildImage` | Build container image with ko |
| `ScanImage` | Trivy vulnerability scan (container image + OS packages) |
| `Govulncheck` | Go dependency vulnerabilities reachable from our code |
| `BuildAndTestBinary` | Build + Redis integration test |

## Taskfile

Common tasks available via `task`:

```bash
task lint              # Run golangci-lint
task build             # Build Go binary
task test              # Run tests
task render-manifests  # Render KCL manifests
task build-scan-image-ko  # Build + scan with ko
```

## Release Process

Releases are automated via semantic-release:

1. Push to `main` triggers build + image workflow
2. On success, release workflow runs semantic-release
3. If releasable commits exist, a new version tag is created
4. Container image is staged from `:main` to `:vX.Y.Z`
5. Kustomize base is pushed as OCI artifact to GHCR

## Vulnerability scanning

Two scanners run in CI, covering different layers. Neither substitutes for the other.

| Scanner | Layer | Where | Gate |
|---------|-------|-------|------|
| Trivy (`ScanImage`) | Container image + OS packages | `build-scan-image.yaml` | Report-only |
| govulncheck (`Govulncheck`) | Go module graph | `build-test.yaml` | Report-only |

govulncheck is call-graph aware: it reports a vulnerability only when the vulnerable symbol is
reachable from our code, so it produces far less noise than a plain dependency diff. Reports upload
as build artifacts (`govulncheck-report`, `trivy-image-report-*`).

Both are **report-only** while the current backlog is worked down — see
[#158](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/158). To make either a hard
gate, pass `--fail-on-vuln` to `Govulncheck`, or set `continue-on-error: false` on the Trivy job.

Run govulncheck locally:

```bash
task govulncheck            # report-only
task govulncheck FAIL=true  # non-zero exit when anything is reachable
```
