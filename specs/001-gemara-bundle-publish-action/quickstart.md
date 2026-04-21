# Quickstart: Gemara Bundle Publish Action

## Prerequisites

- A **Gemara** root YAML (policy, catalog, or guidance catalog) and any **URI-resolvable** dependencies your repo uses.
- Registry credentials with **push** permission (for GHCR: `permissions: packages: write` and `GITHUB_TOKEN` or PAT).
- **HTTPS** registry hostname (v1 composite does not support plain HTTP first-class).

## Use from another repository

1. Copy [examples/workflow-publish-with-pinned-action.yml](../../examples/workflow-publish-with-pinned-action.yml) into `.github/workflows/` in your content repo.
2. Set `registry`, `repository`, `tag`, `file`, `username`, `password`.
3. Pin `uses:` to a **semver tag** or **commit SHA** of this action (not `@main` for releases).
4. Read `steps.<id>.outputs.digest` after publish if you need the manifest digest.

## Local smoke (`go run`)

```bash
cd tools/publish
export GEMARA_REGISTRY_PASSWORD='<token>'
go run . \
  -registry=ghcr.io \
  -repository='<org>/<repo-or-package-path>' \
  -tag='dev-test' \
  -file="$(pwd)/testdata/minimal-catalog.yaml" \
  -username='<actor>'
```

Expect `digest=sha256:…` on success.

## E2E on GitHub

**CI (default):** on push and same-repo PRs, workflow **[CI](../../.github/workflows/ci.yml)** runs job **`e2e-publish-ghcr`** (GHCR publish + digest assertion; skipped on fork PRs). **Manual:** run **[E2E publish to GHCR](../../.github/workflows/e2e-ghcr.yml)** via **Actions → workflow_dispatch**. Both require **`packages: write`** where applicable.

## Two-phase (layout + transport only)

To push a **pre-built OCI layout** with ORAS CLI glue, use a transport-only action such as [gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci) (`layout-copy`). Do not reimplement **pack** in that step.

## Where to read more

- [README.md](../../README.md) — full inputs, outputs, edge cases, SDK pin.
- [spec.md](./spec.md) — requirements and success criteria.
- [plan.md](./plan.md) — implementation plan and constitution check.
