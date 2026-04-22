# Examples — work you can do **now**

This folder is **not** wired into the root [`action.yml`](../action.yml). It collects **reproducible, pin-friendly** patterns (pinned images, explicit `docker run` / compose entrypoints, semver-pinned `uses:`) so contributors can ship value **without** waiting for an upstream `gemara` CLI. It takes **lightweight inspiration** from common container CI practice; it does **not** mirror any other project’s layout.

## What exists today (use or extend)

| Artifact | What it is | Next step (concrete) |
|----------|------------|----------------------|
| [`workflow-publish-with-pinned-action.yml`](workflow-publish-with-pinned-action.yml) | Caller workflow: checkout + **semver-pinned** `uses: …/gemara-bundle-publish-action@v…` | Copy into a **content repo** under `.github/workflows/`; replace `file` / `repository` / triggers. |
| [`Dockerfile.publish.sketch`](Dockerfile.publish.sketch) | Multi-stage **build** of [`tools/publish`](../tools/publish) → static binary in distroless | Add a **`.github/workflows`** job here (or in org) to `docker build` + push to **GHCR**; pin consumers by **digest**. |
| [`docker-compose.publish.sketch.yml`](docker-compose.publish.sketch.yml) | **Local smoke**: build image, mount repo read-only, run publish with flags | Export **`GEMARA_REGISTRY_PASSWORD`** (registry token) in your shell before `docker compose run`; never commit token values. Adjust `file=` to a real root YAML. |
| [`workflow-publish-with-docker-image.yml`](workflow-publish-with-docker-image.yml) | Optional CI path: **no `go run`** on the runner—only `docker run` a pinned image | After GHCR image exists, use this in repos that prefer container-only jobs. |
| [`.github/workflows/e2e-ghcr.yml`](../.github/workflows/e2e-ghcr.yml) | **Maintainer** `workflow_dispatch`: publish [`tools/publish/testdata/minimal-catalog.yaml`](../tools/publish/testdata/minimal-catalog.yaml) to GHCR and assert `digest` output | Run from the Actions tab on this repo (needs `packages: write`). |
| **Two-phase (layout → registry)** | Not an example file in this folder—use a **transport-only** action (e.g. [gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci) `layout-copy`) after producing an **OCI image layout** with SDK or other tooling | Keeps **pack** out of the transport step; see root [README.md](../README.md) “Two-phase publish”. |

## Principles

1. **Pin what you run** — action `@v1.0.0` or **image digest**, not floating `latest`, for release pipelines.  
2. **One entry point** — one workflow job, or one `docker compose … run`, so env and args stay explicit.  
3. **Optional image** — build `tools/publish` once, run the binary from GHCR on cold jobs (faster than `go run` every time).

These are minimal **stubs for this repo only**; grow them when the action lands under `gemaraproj` and GHCR image names are decided.

## Suggested backlog (ordered)

1. **Land the action repo** under `gemaraproj`, tag **`v0.1.0`**, update the workflow example to that ref.  
2. **Image publish** — workflow that builds `Dockerfile.publish.sketch` and pushes `ghcr.io/gemaraproj/gemara-bundle-publish-tool:<tag>` (name TBD by maintainers).  
3. **Dogfood** — one public fixture repo (or branch) that publishes a bundle on tag using the **pinned composite** example.  
4. **SDK pin** — when [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) is on a **released** `go-gemara` tag, drop the `replace` in `tools/publish/go.mod` and re-release the action + image.

## Future (upstream CLI — optional, not blocking)

If **go-gemara** later ships a **`gemara … publish`** (or similar) command, you can replace `go run …/tools/publish` in root `action.yml` with `go install …@tag` + that CLI, **or** bake the official binary into the Docker image and keep a thin composite. No separate hypothetical `action.yml` is maintained here until flags exist in upstream docs.
