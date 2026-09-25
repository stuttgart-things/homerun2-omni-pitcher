# Preview write-loop measurement (stuttgart-things#3200)

**This file exists only to open a PR. Do not merge it.**

A preview namespace on `homerun2-test1` was measured writing to etcd at roughly
250 MB/hour — enough to reach the 2.1 GB quota in about seven hours. All of that
traffic came from one Kyverno policy (`homerun2-omni-pitcher-preview-secrets`)
and the ExternalSecrets it generates.

The suspected mechanism is a `generate` rule with `synchronize: true` against an
object that External Secrets Operator rewrites every minute
(`refreshInterval: 1m`): Kyverno reverts what ESO writes, ESO reconciles again,
and neither settles.

That is a **hypothesis**. It fits every measurement taken during the incident,
but by the time it was formed the namespace had been deleted, so the loop could
no longer be observed directly. This PR recreates the conditions deliberately,
on a cluster that now carries the guards from stuttgart-things/argocd#523, so
the question can be answered with numbers instead of inference.

Two things are being watched:

1. the rate of `UpdateRequest` create/update/delete and of writes to
   `externalsecrets`, against the pre-preview baseline;
2. whether `homerun2-omni-pitcher-preview-sweep` reclaims the namespace once
   this PR is closed — that behaviour is still unverified, because during the
   incident the namespace was deleted by hand three minutes before the sweep's
   next run.

Delete this file once the measurement is recorded in stuttgart-things#3200.
