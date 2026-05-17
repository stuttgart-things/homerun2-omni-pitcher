# Preview Environments

Every pull request opened against `main` can spin up an ephemeral, fully-deployed instance of omni-pitcher on the `homerun2-dev` Kubernetes cluster — alongside redis-stack so reviewers can `curl /pitch` and watch events land in the stream. The environment lives for as long as the PR is open and tears down automatically on merge or close.

omni-pitcher was the pilot for the homerun2 PR-preview rollout; this is the original shape that core-catcher and scout now mirror.

This page covers how to use it, what each PR gets, the components that make it work, and how to troubleshoot.

## Quick start

1. Open a PR against `main`.
2. Add the `preview` label: `gh pr edit <num> --add-label preview`.
3. Wait 5–10 minutes for the image build, the kustomize-OCI push, and Argo's PullRequest generator poll (every 600s).
4. The preview-bot leaves a sticky comment on the PR with the URL.

Closing or merging the PR tears the namespace down automatically.

## What you get per PR

Each preview lives in its own namespace: `homerun2-omni-pitcher-pr-<num>` on `homerun2-dev`. The namespace contains:

| Workload | Purpose |
|--|--|
| `homerun2-omni-pitcher` | The system under test (this PR's commit) |
| `redis-stack` | The bus omni-pitcher writes into; persistence disabled (ephemeral) |
| `seed-test-events` (one-shot Job) | Posts a 5-event fixture to omni-pitcher right after the Deployment becomes Ready, so the stream is non-empty on first inspection |

omni-pitcher is the SUT and has no co-tenants — unlike scout and core-catcher previews, there's no upstream/downstream component to pair with. Reviewers exercise it directly via `curl`.

Reachable at: `https://omni-pr-<num>.homerun2-dev.sthings-vsphere.labul.sva.de`

## Why the `preview` label gate

Without the label, every renovate / dependabot dep-bump PR would spawn a namespace. Two problems:

- Branches predating the build-pr workflow have no `pr-<num>-<sha>` image or kustomize artifacts published — half-empty namespaces with sync errors.
- Bots open dozens of PRs per week; the preview infrastructure isn't built for that scale.

Human-opened PRs opt in via the label. Bots don't apply it, so they're excluded by default. The Argo AppSet's PullRequest generator filters on `labels: [preview]`.

## The flow, end to end

```
git push (PR opens)
   ├─► comment-preview-url.yaml  ─►  sticky bot comment with URL
   ├─► build-scan-image.yaml     ─►  ko-built image at ghcr.io/.../homerun2-omni-pitcher:pr-<num>-<sha>
   ├─► push-kustomize-pr.yaml    ─►  kustomize OCI at ghcr.io/.../homerun2-omni-pitcher-kustomize:pr-<num>-<sha>
   └─► build-test.yaml + lint    ─►  CI gates

Argo PullRequest generator (poll every 600s)
   └─► detects PR with `preview` label
       └─► renders parent Application `homerun2-omni-pitcher-pr-<num>` in argocd ns
           └─► chart emits child Applications targeting `homerun2-omni-pitcher-pr-<num>` ns
               on the homerun2-dev cluster

Kyverno ClusterPolicies (auto-fire on namespace create)
   ├─► generate ResourceQuota + LimitRange
   ├─► generate 3 ExternalSecrets → ESO materializes Secrets from Vault
   └─► generate one-shot seed Job (posts fixture after Deployment Ready)

PR close
   ├─► AppSet drops the entry → finalizer cascade prunes child Apps + workloads
   ├─► cleanup-pr-artifacts.yaml deletes both ghcr.io packages
   └─► Kyverno ClusterCleanupPolicy reaps any empty namespace shell left behind
```

## The four PR-preview workflows in this repo

All four are in `.github/workflows/` and trigger on `pull_request` events targeting `main`.

| Workflow | Trigger | Output |
|--|--|--|
| `build-scan-image.yaml` | PR opened/updated | ko-built image tagged `pr-<num>-<sha>` + `pr-<num>` |
| `push-kustomize-pr.yaml` | PR opened/updated | kustomize OCI tagged `pr-<num>-<sha>` (renders `kcl/main.k` against `tests/kcl-deploy-profile.yaml`) |
| `comment-preview-url.yaml` | PR opened/reopened | Sticky comment with URL, namespace, ArgoCD link. Thin caller of `stuttgart-things/github-workflow-templates/.github/workflows/call-comment-preview-url.yaml` |
| `cleanup-pr-artifacts.yaml` | PR closed | Deletes both ghcr.io packages so version histories don't fill with PR debris |

All four delegate to reusable workflows in `stuttgart-things/github-workflow-templates`.

## The Argo AppSet, briefly

Lives at `stuttgart-things/stuttgart-things` under `clusters/labul/vsphere/platform-sthings/argocd/homerun2-dev/omni-pitcher-pr-preview-appset.yaml`. The shape:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: homerun2-omni-pitcher-pr-preview
  namespace: argocd
spec:
  generators:
    - pullRequest:
        github:
          owner: stuttgart-things
          repo: homerun2-omni-pitcher
          tokenRef: { secretName: homerun2-omni-pitcher-pat, key: token }
          labels: [preview]               # the gate
        requeueAfterSeconds: 600          # poll cadence
  template:
    metadata:
      name: 'homerun2-omni-pitcher-pr-{{ .number }}'
      finalizers: [resources-finalizer.argocd.argoproj.io]   # cascade on prune
    spec:
      source:
        repoURL: https://github.com/stuttgart-things/argocd.git
        path: apps/homerun2/install
        helm:
          valuesObject:
            destination:
              name: homerun2-dev
              namespace: 'homerun2-omni-pitcher-pr-{{ .number }}'
            omniPitcher:
              enabled: true
              version: 'pr-{{ .number }}-{{ .head_sha }}'
              hostname: 'omni-pr-{{ .number }}.homerun2-dev.sthings-vsphere.labul.sva.de'
              inlineHttpRoute: true       # Option B — see below
            redisStack:
              enabled: true
              persistence: { enabled: false }
              auth: { existingSecret: redis-stack-auth }
            # all other components off
            httpRoute:
              enabled: true
              gateway: { name: homerun2-dev-gateway, namespace: default }
      syncPolicy:
        automated: { prune: true, selfHeal: true }
        syncOptions: [CreateNamespace=true, ServerSideApply=true]
```

The AppSet renders one **parent** Argo `Application` per labelled PR. The parent's source is the `apps/homerun2/install` chart in the `stuttgart-things/argocd` catalog. The chart emits **child** Applications (one per enabled component: omni-pitcher, redis-stack) on the homerun2-dev cluster.

`destination.name: homerun2-dev` (not a URL) means the chart targets the workload cluster by its registered Argo cluster name, so IP / DNS changes don't break manifests.

## The five cluster overlay manifests

Sit alongside the AppSet in `…/argocd/homerun2-dev/`:

| File | What it does |
|--|--|
| `omni-pitcher-pr-preview-appset.yaml` | The ApplicationSet above |
| `homerun2-omni-pitcher-preview-quota.yaml` | Kyverno `ClusterPolicy` → generates `ResourceQuota` + `LimitRange` in each PR namespace |
| `homerun2-omni-pitcher-preview-secrets.yaml` | Kyverno `ClusterPolicy` → generates 3 `ExternalSecret`s; ESO pulls from Vault `homerun2-pr/data/preview-env` |
| `homerun2-omni-pitcher-preview-seed-data.yaml` | Kyverno `ClusterPolicy` → generates the one-shot seed Job |
| `homerun2-omni-pitcher-preview-sweep.yaml` | Kyverno `ClusterCleanupPolicy` → cron-reaps empty PR namespace shells |

These are deployed *once per cluster*. Per-PR, they fire automatically when the AppSet creates the namespace.

## HTTPRoute: Option B (inline in the kustomize OCI)

The HTTPRoute exposing omni-pitcher externally is rendered by `kcl/httproute.k` and ships **inside the kustomize OCI**, alongside the Service. They land in the same kustomize apply, eliminating the cross-Application race that previously let Cilium's gateway controller stamp a sticky `BackendNotFound` (tracked under [stuttgart-things/argocd#116](https://github.com/stuttgart-things/argocd/issues/116)). This repo was the first to ship Option B; the chart-side helper + flag landed in [stuttgart-things/argocd#117](https://github.com/stuttgart-things/argocd/pull/117) and [#119](https://github.com/stuttgart-things/argocd/pull/119). Three places have to agree:

| Repo | Setting |
|--|--|
| `homerun2-omni-pitcher` (this repo) | `tests/kcl-deploy-profile.yaml` → `config.httpRouteEnabled: true` |
| `stuttgart-things/argocd` | `apps/homerun2/install` → `omniPitcher.inlineHttpRoute` flag patches the rendered HTTPRoute's parentRef + hostname per env, and excludes omni-pitcher from the standalone httproute Application |
| `stuttgart-things/stuttgart-things` | Set `omniPitcher.inlineHttpRoute: true` in the AppSet's `valuesObject` |

With all three set, `HTTPRoute/homerun2-omni-pitcher` lands `ResolvedRefs: True` on first reconcile. No manual `kubectl annotate httproute reconcile-bump=$(date +%s) --overwrite` required.

Admission-defaulted fields (`parentRefs.group`/`kind`, `backendRefs.group`/`kind`/`weight`) are rendered explicitly by `kcl/httproute.k` so the chart-rendered shape matches what Cilium writes back — no perpetual `OutOfSync` from defaulting drift.

## Lifecycle

| Event | Result |
|--|--|
| PR opened with `preview` label | Sticky bot comment posted; CI builds image + kustomize OCI; AppSet picks it up within 600s; namespace + workloads spin up |
| PR updated (new commit) | Image + kustomize OCI rebuilt with new `<sha>`; AppSet detects the head-SHA change; rolling update of Deployments |
| PR `preview` label removed | AppSet drops the entry; finalizer prune cascades teardown |
| PR closed (merged or rejected) | AppSet drops the entry → teardown; `cleanup-pr-artifacts.yaml` deletes ghcr.io packages |

The `resources-finalizer.argocd.argoproj.io` finalizer on the parent Application is critical — without it, Argo would delete the parent instantly when the AppSet drops it, orphaning child Apps + workload pods. With it, Argo runs prune on every managed resource first.

## Troubleshooting

| Symptom | Likely cause | Fix |
|--|--|--|
| No bot comment, no namespace | `preview` label missing | `gh pr edit <num> --add-label preview` |
| Bot comment present, namespace never appears | AppSet hasn't polled yet | Wait up to 10 min, or `kubectl -n argocd annotate appset homerun2-omni-pitcher-pr-preview argocd.argoproj.io/refresh=hard` |
| Parent Application sync error: `failed to load: oci pull` | Image / kustomize OCI build still running or failed | Check the PR's Actions tab — `build-pr` and `push-kustomize` must both be green |
| Pods stuck `ImagePullBackOff` | ghcr.io tag not yet pushed (CI still running) or PR closed (cleanup workflow already ran) | Wait for build / reopen the PR |
| Pods CrashLoopBackOff with `WRONGPASS` | ESO hasn't materialized `redis-stack-auth` Secret yet | Check `kubectl -n homerun2-omni-pitcher-pr-<num> get externalsecret`; refresh if not Ready |
| HTTPRoute `ResolvedRefs: False` | Service didn't land before HTTPRoute (pre-Option-B environments only) | Should not happen now; if it does: `kubectl annotate httproute homerun2-omni-pitcher reconcile-bump=$(date +%s) --overwrite -n homerun2-omni-pitcher-pr-<num>` and file an issue |
| `POST /pitch` returns 401 | `AUTH_TOKEN` env not set on the Deployment or `homerun2-omni-pitcher-token` Secret not materialized | Check the Deployment env + the per-namespace ExternalSecret status |
| `POST /pitch` returns 500 with `WRONGPASS` | omni-pitcher started before redis-stack was ready, retried, gave up | Should be smoothed by the bounded 30s startup retry; if seen, restart the pod |
| Seed Job ran but only 4 events posted | Known shell-script JSON-splitting bug — last event drops | Tracked as a follow-up in [stuttgart-things/homerun2-omni-pitcher#116](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/116) |
| Namespace stuck Terminating after PR close | Finalizer on a CRD instance | `kubectl get all,externalsecret -n homerun2-omni-pitcher-pr-<num>` to find the blocker |

## See also

- [stuttgart-things/argocd `apps/homerun2`](https://github.com/stuttgart-things/argocd/tree/main/apps/homerun2) — the install chart + Kyverno policy charts the AppSet consumes
- [stuttgart-things/homerun2-omni-pitcher#116](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/116) — the umbrella rollout issue tracking all 8 components
- [stuttgart-things/argocd#116](https://github.com/stuttgart-things/argocd/issues/116) — the HTTPRoute creation-order race writeup that motivated Option B
- [stuttgart-things/github-workflow-templates](https://github.com/stuttgart-things/github-workflow-templates) — the four reusable PR-preview workflows this repo delegates to
