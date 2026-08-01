# Triage #2086 — servicemesh / chain tracing breaks on Kubernetes 1.25+

> **This PR is intentionally triage-only.** It contains **no source
> changes** to KubeKey v4. The objective is to document the bug
> referenced in #2086, establish the v3-vs-v4 scope mismatch, and
> capture the recommended fix path so a maintainer can pick it up.

Issue: https://github.com/kubesphere/kubekey/issues/2086
Reporter: @370569218
Reporter environment: KubeKey **v3.0.13**, CentOS 7.9, K8s **v1.25+**
with `servicemesh` feature enabled.

---

## 1. Bug summary

KubeKey v3.4.1 fails to install the chain-tracing / Jaeger backend when
the `servicemesh` feature is enabled on a Kubernetes cluster running
**v1.25 or newer**.

Root cause: KubeKey's servicemesh/jaeger manifests declare one or more
`CronJob` resources with `apiVersion: batch/v1beta1`. That API group
was **removed in Kubernetes 1.25** (alongside the rest of the
`*.k8s.io/v1beta1` set). The replacement is `apiVersion: batch/v1`,
which is GA, has identical field semantics for `CronJob`, and is
required on any cluster ≥ 1.25.

Affected K8s versions: **1.25+ (any minor/patch)**
Affected KubeKey versions: **v3.x line that still ships the
`servicemesh` feature** (issue filed against v3.4.1; same manifest
likely also present on adjacent 3.x tags).

## 2. Why no source change in this PR

The current checkout is **KubeKey v4** (`main` at `d278e6a8`), the
Go-only task-execution framework modeled on Ansible. A repository-wide
search shows the feature simply **does not exist here yet**:

| Probe | Result |
|-------|--------|
| `grep -rln 'batch/v1beta1' --exclude-dir=.git` | 0 matches |
| `grep -rln 'kind: CronJob' --exclude-dir=.git --include='*.{yaml,yml,tmpl,j2}'` | 0 matches |
| `grep -rln -i 'jaeger\|servicemesh' --exclude-dir=.git` | 1 incidental match in `builtin/core/roles/image-registry/harbor/templates/harbor.yml` (unrelated) |
| `git log --grep='batch\|cronjob\|jaeger\|servicemesh'` | 1 unrelated hit — `3cb23796 Refactor batch execution` (2020 Ansible executor refactor) |

What **does** exist:

- `pkg/const/scheme.go:23` already registers
  `batchv1 "k8s.io/api/batch/v1"`. So if/when servicemesh manifests
  are introduced on a v4 release, the kubekey-side decoding path is
  already on the GA API.

In short: nothing to migrate in this repository at this time. Attempting
to commit a one-line `apiVersion` swap here would either create a
half-formed autodiff against a vendored manifest that doesn't exist,
or pull in a v3-era servicemesh playbook that would break v4's
executor model. The fix belongs on **KubeKey v3's tree**, and a
parallel v4 design needs a separate decision.

## 3. Recommended action (per scope)

This split is the responsible path; a single PR on `main` is **not.**

### 3a. Fix on KubeKey v3 (`release-3.4`)

Apply on the corresponding v3 release branch (e.g. `release-3.4`,
`release-3.x`):

1. Identify every `CronJob` manifest in the v3 servicemesh/jaeger
   tree. Likely roots based on common v3 layouts:
   - `deploy/servicemesh/*.yaml`
   - `deploy/jaeger/*.yaml`
   - any Kustomize/Helm umbrella generating `kind: CronJob`
2. For each, replace:

   ```yaml
   apiVersion: batch/v1beta1
   kind: CronJob
   ```

   with:

   ```yaml
   apiVersion: batch/v1
   kind: CronJob
   ```

3. The `batch/v1` `CronJob` spec is field-compatible with the
   `v1beta1` form for the well-known Jaeger cleanup / es-index /
   es-rollover use cases. No spec rewrite is expected.
   `suspend`, `successfulJobsHistoryLimit`,
   `failedJobsHistoryLimit`, `startingDeadlineSeconds`, and
   `schedule` are all GA in `batch/v1`.
4. Verify on a real 1.25+ cluster:

   ```bash
   kubectl --dry-run=server -f deploy/servicemesh apply
   kubectl apply -f deploy/servicemesh
   kubectl wait --for=condition=Ready -n jaeger-system pod -l app=cron
   ```

5. Backport the change to all maintained v3.x release lines where
   the issue reproduces.

### 3b. Tracking issue for v4

If/when the `servicemesh` feature is being ported to v4, file a **v4
design issue** that:

- Names `batch/v1` as the only accepted `CronJob` API in any new
  playbook that ships here.
- Bakes a CI guard: lint step that fails if any
  `apiVersion: batch/v1beta1` regresses into the tree.
- Tracks parity with the v3 servicemesh manifest set so the v4 port
  isn't a re-discovery exercise.

This PR does **not** open that issue — it only records the requirement
here.

## 4. Files touched by this PR

The PR's diff is exactly one file:

| File | Why |
|------|-----|
| `docs/pr-2086-batch-v1.md` | This triage document |

No `builtin/`, `pkg/`, `cmd/`, `config/`, or `api/` change. No CRD
regenerate, no manifest edit, no version bump.

## 5. Acceptance criteria

- [ ] A reviewer (Maintainer) confirms the v3-vs-v4 mismatch
      described in §2.
- [ ] A corresponding fix is queued **on the v3 release branch** (`fix
      v3 servicemesh cronjob api to batch/v1`), with a link from
      #2086.
- [ ] v4 design issue filed (or this PR amended) to pin
      `batch/v1` as the supported CronJob API going forward.
- [ ] This branch is **not** merged to `main` until at least one of
      the above is true; the branch exists as a durable breadcrumb.

## 6. Out of scope

- Touching v3 vendored manifests from this PR. A cross-repo PR is the
  wrong vehicle for a v3 fix.
- Renaming or refactoring the `servicemesh` feature workflow.
- Any change to `pkg/const/scheme.go` — it already imports
  `k8s.io/api/batch/v1`, so it does not need a follow-up.
- Migrations for other removed-in-1.25 APIs that may be in the v3
  tree (`PodSecurityPolicy`, `HorizontalPodAutoscaler` v2beta2,
  etc.). Each deserves its own triage PR if it shows up.

## 7. Test plan

This PR is docs-only — there is no executable code to test. The
verification surface is:

1. `git ls-files builtin pkg cmd api config | xargs grep -l 'batch/v1beta1'`
   on `main` → empty (confirms §2).
2. `git diff origin/main -- docs/pr-2086-batch-v1.md` → shows the
   single new file.
3. `go build ./...` from repo root → succeeds (sanity check that the
   addition didn't accidentally corrupt anything; expected since this
   PR touches no `.go`).
4. `golangci-lint run` → succeeds for the same reason.

A reviewer running `make lint make test make kk` should observe no
diff in behavior. Any failure points at operator error in this PR and
should block merge.

## 8. Risk / rollback

Risk profile: **~none**.

- No runtime or build-time code changes.
- No CRD regeneration, no API surface change, no breaking JSON.
- Rollback = `git revert` this commit (or close the PR). The
  v3-side fix is unaffected by what we do here.

## 9. Follow-ups (intentionally NOT in this PR)

- Open the actual v3 fix PR against `release-3.4`.
- File the v4 parity design issue tracking §3b.
- If the v3 fix is backported to multiple `release-3.x` lines, open
  one PR per line.

---

## Source citations

- Issue #2086 — reporter's reproduction on KubeKey v3.4.1
  + Kubernetes 1.25+.
- Kubernetes 1.25 release notes — removal of `batch/v1beta1`,
  `PodDisruptionBudget`, and other APIs in favor of GA equivalents.
- `pkg/const/scheme.go:23` — kubekey already registers `batch/v1`
  into the controller-runtime scheme (confirms the runtime side
  needs no migration).
- `AGENTS.md` — KubeKey agent conventions (logging levels, error
  handling, conventional commit format, agent pipeline); followed
  for commit message and PR scope discipline.
