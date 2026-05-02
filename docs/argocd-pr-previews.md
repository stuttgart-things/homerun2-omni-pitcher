# ArgoCD PR Previews

Each open PR on this repository can spin up its own ephemeral environment - a single namespace running **redis-stack + omni-pitcher** side-by-side - via Argo CD's `pullRequest` ApplicationSet generator.

The ApplicationSet, umbrella chart, and full setup live in [`stuttgart-things/argocd` -> `apps/omni-pitcher-preview/`](https://github.com/stuttgart-things/argocd/tree/main/apps/omni-pitcher-preview).

## Workflow

1. Open a PR against `main` of this repo.
2. Add the `preview` label - the ApplicationSet only picks up labelled PRs (drop the label filter on the ArgoCD side if you want every PR auto-previewed).
3. Wait ~3 minutes (`requeueAfterSeconds: 180`). Argo CD creates:
   - `Application/omni-pitcher-pr-<n>`               (wrapper, in `argocd` ns)
   - `Application/omni-pitcher-pr-<n>-redis-stack`   (child)
   - `Application/omni-pitcher-pr-<n>-omni-pitcher`  (child)
   - `Namespace/omni-pitcher-pr-<n>`                 with the redis-stack StatefulSet + omni-pitcher Deployment.
4. Push more commits - Argo CD re-syncs the same Application against the new head SHA.
5. Close / merge the PR - Argo CD prunes the wrapper Application; both children and the namespace go with it.

## What CI must publish per PR

The ApplicationSet template defaults `omniPitcher.version` to the PR's `head_short_sha`. CI must publish, tagged with that short SHA:

- container image  `ghcr.io/stuttgart-things/homerun2-omni-pitcher:<sha>`
- kustomize OCI    `ghcr.io/stuttgart-things/homerun2-omni-pitcher-kustomize:<sha>`

If your tag scheme differs (e.g. `pr-<number>`), override `omniPitcher.version` in the ApplicationSet template on the ArgoCD side.

## Smoke-test a live preview

```bash
NS=omni-pitcher-pr-42
kubectl -n $NS port-forward svc/homerun2-omni-pitcher 8080:80 &
curl -X POST http://localhost:8080/pitch \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer preview-auth-token" \
  -d '{"title":"Test","message":"hello","severity":"info","author":"me"}'
```

Default preview credentials (override via the ApplicationSet `valuesObject` for sensitive flows):
- `redisPassword`: `preview-redis-password`
- `authToken`:     `preview-auth-token`

## See also

- Setup + chart:        [`stuttgart-things/argocd` - `apps/omni-pitcher-preview/`](https://github.com/stuttgart-things/argocd/tree/main/apps/omni-pitcher-preview)
- Production stack:     [`stuttgart-things/argocd` - `apps/homerun2/`](https://github.com/stuttgart-things/argocd/tree/main/apps/homerun2)
- KCL render (for non-Argo deploys): [`kcl/`](../kcl/)
