# Changelog

All notable releases of this GitHub Action should record **two pins**: the **git ref** of this repository (semver tag or commit SHA consumers use in `uses:`) and the **go-gemara** dependency as declared in `tools/publish/go.mod` at that ref.

## Release template (copy for each GitHub Release)

Replace angle-bracket placeholders when publishing:

1. **Action pin** — Repository name and `uses:` ref (tag or full SHA).
2. **SDK pin** — Paste the `require` and `replace` lines from `tools/publish/go.mod` at that commit.
3. **Validation** — Note CI jobs run (`publish-tool`, `publish-invalid-root`, `e2e-publish-ghcr`, Docker sketch) and optional `e2e-ghcr` workflow_dispatch with digest output.

## Unreleased

- CI: **`e2e-publish-ghcr`** job in `.github/workflows/ci.yml` — live GHCR publish + `outputs.digest` verification on push and same-repo PRs (SC-004); fork PRs skip this job.
- Documented governance (#63), mapping-reference behavior, Speckit plan/tasks, and composite contract alignment.
- Current `replace` line (pre–upstream tag for bundle APIs): see **`tools/publish/go.mod`** (not duplicated here to avoid noisy secret-detection scans).
