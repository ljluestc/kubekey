# Regression guard for #2086

This branch implements the §3b recommendation from
`docs/pr-2086-batch-v1.md` as a working code change.

## What was added

| File | Purpose |
|------|---------|
| `pkg/const/scheme_test.go` | `TestSchemeRegistersBatchV1CronJob` — asserts the controller-runtime scheme recognizes `batch/v1` `CronJob`. `TestNoBatchV1Beta1InTree` — walks `builtin/`, `plugins/`, `config/` and fails on any embedded `apiVersion: batch/v1beta1` reference. |
| `Makefile` (after line 297) | New target `make lint-regression-batch-v1` that greps the same three root trees and exits non-zero on any hit. |

## How the guard fires

- `make test` runs the Go test (`go test ./...`) → the test fails fast.
- `make lint-regression-batch-v1` runs the same check standalone, useful in CI before `go test` finishes downloading its toolchain, and in pre-commit / pre-rebase hooks.

## Verified locally

```
$ go test -v -run 'TestSchemeRegistersBatchV1CronJob|TestNoBatchV1Beta1InTree' ./pkg/const/
=== RUN   TestSchemeRegistersBatchV1CronJob
--- PASS: TestSchemeRegistersBatchV1CronJob (0.00s)
=== RUN   TestNoBatchV1Beta1InTree
    scheme_test.go:104: scanned 284 manifest files across 3 roots for batch/v1beta1
--- PASS: TestNoBatchV1Beta1InTree (0.00s)

$ make lint-regression-batch-v1
OK: no batch/v1beta1 references
```

A negative smoke test (`apiVersion: batch/v1beta1` written into
`builtin/__guard_smoke__/regression_smoke.yaml`) was used to verify the
guard wakes up both at the test layer (`--- FAIL`) and at the Makefile
layer (`make: *** Error 1`); the smoke tree was deleted before commit.

## Out of scope

- No churn to `pkg/const/scheme.go` itself — the runtime side already
  registers `batch/v1`, so the migration story is entirely about
  ensuring no manifest regresses.
- No new public package surface — `_test.go` and a Makefile target do
  not count as API additions.
- No CI workflow file added; the guard is wired through `make` so a
  workflow author can opt in later by adding
  `- run: make lint-regression-batch-v1` to the existing CI pipelines.

## Suggested rollout

1. Land this branch first; the guard immediately fails on any future
   regression.
2. Open the v3-side fix PR against `kubesphere/kubekey release-3.4`
   per §3a of the original PR description.
3. File the v4 parity design issue per §3b of the original PR
   description; reference this branch as the in-tree enforcement
   mechanism.
