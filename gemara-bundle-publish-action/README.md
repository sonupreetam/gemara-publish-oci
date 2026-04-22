# Gemara bundle publish (GitHub Action)

Workstream for [go-gemara issue #63](https://github.com/gemaraproj/go-gemara/issues/63): publish **Gemara OCI bundles** using the **SDK today** ([PR #62](https://github.com/gemaraproj/go-gemara/pull/62) `bundle` APIs + **oras-go**), **without** blocking on a future upstream `gemara` CLI. This sits under the broader **OCI standardization** effort in [go-gemara issue #60](https://github.com/gemaraproj/go-gemara/issues/60).

**SDK owns the contract.** Manifest shape, **media types**, and **layer assembly** are defined and implemented in **go-gemara** (`bundle.Pack`, `bundle.Unpack`, and related types). This repository supplies **CI orchestration** (inputs, `go run` of [`tools/publish`](tools/publish), registry auth wiring) and must **track SDK releases** when the bundle layout changes.

### Shippable today (pick what fits your org)

1. **Composite action** ([`action.yml`](action.yml)) — `go run tools/publish` on the runner; same inputs as issue #63; pins SDK via [`tools/publish/go.mod`](tools/publish/go.mod).
2. **Optional tool image** ([`examples/Dockerfile.publish.sketch`](examples/Dockerfile.publish.sketch) + [`examples/workflow-publish-with-docker-image.yml`](examples/workflow-publish-with-docker-image.yml)) — build once, **pin by tag or digest**, `docker run` in CI (reproducible cold jobs; same binary as `tools/publish`, not a second implementation).

The [`examples/README.md`](examples/README.md) file is the **ordered backlog** (land repo → push image → dogfood → drop `replace` when PR #62 is released). That is the main place for “what to work on next.”

### Behaviour (technical)

**Assemble** → **pack** into an **in-memory** OCI store → **`oras.Copy`** to the registry (prepack-then-copy, aligned with maintainer discussion). Resolve by manifest **digest** first; if that fails, **tag** the root as `gemara-publish/__root__` and **copy** from that ref. No `oras` CLI for manifest/media-type details—those stay in **`bundle.Pack`**.

See **Comparison with PR #62** below for differences vs the PR README’s disk + `CopyGraph` example.

## Comparison with [PR #62](https://github.com/gemaraproj/go-gemara/pull/62)

PR #62 (SDK and README on that branch) and this action share the same **contract**: `bundle.Assemble` → `bundle.Pack` → oras-go to move the artifact to a registry. Differences are mostly **where** the packed bits live before push and **which** oras-go entrypoint is used.

| | PR #62 README example | This action (`tools/publish`) |
|---|------------------------|-------------------------------|
| **Assemble** | `NewAssembler(&fetcher.URI{})`, `Assemble(ctx, m, src)` | Same |
| **Pack target** | OCI **filesystem** layout (`oci.New("./bundle-output")`) — easy to inspect or `oras cp` from disk | **In-memory** store (`memory.New()`) — no temp directory, fits ephemeral runners |
| **Pre-push tagging** | Tag the **layout** with the version string, `Resolve`, then push | After pack, **`store.Resolve(digest)`** (like `oras resolve`); if that errors, **`store.Tag`** → `gemara-publish/__root__`, then **`oras.Copy`** from that ref |
| **Registry transport** | `oras.CopyGraph` from layout → remote, then `repo.Tag` on the remote | Single **`oras.Copy`** from memory → remote with destination ref = your **tag** (same idea as “copy prepackaged local → remote”) |
| **Auth** | Not shown in the README snippet | `remote` client configured with **`auth.Client`** / `StaticCredential` (e.g. GHCR token) |
| **Optional checks** | — | Optional **`gemara.Load`** on the root file before assemble |
| **Unpack** | README shows `bundle.Unpack(ctx, repo, tag)` for consumers | Out of scope for publish; consumers use the SDK or another workflow |

**Maintainer discussion** (issue #63 thread) also mentioned **memory + `oras.Copy`** by digest; the **PR #62 README** still documents **disk layout + `CopyGraph`**. Both are valid: disk is better for debugging and repeatable local bundles; memory is minimal for CI. This repo follows the **memory + `oras.Copy`** path unless you change it to match the README literally.

## End-to-end verification (SDK alignment)

Before bumping the SDK pin or declaring compatibility with a new **go-gemara** drop, run an **end-to-end** publish against a **test registry** (for example GHCR in a throwaway package or tag) using the same **`tools/publish` / composite action** revision you intend to ship. The goal is to confirm **assemble → pack → push → resolve by tag** against the **bundle implementation** you depend on ([PR #62](https://github.com/gemaraproj/go-gemara/pull/62) or its merged successor), including the **digest** output matching what the registry serves. CI runs **`e2e-publish-ghcr`** in [`.github/workflows/ci.yml`](.github/workflows/ci.yml) on push and same-repo pull requests (live GHCR publish + digest check). The manual workflow [`.github/workflows/e2e-ghcr.yml`](.github/workflows/e2e-ghcr.yml) (`workflow_dispatch`) is also available for maintainers.

## Two-phase publish (transport only)

If you prefer to **pack to an OCI image layout on disk** (SDK or other tooling) and then **copy** to a registry with minimal glue, use a **transport-only** pattern (for example [`oras cp --from-oci-layout`](https://oras.land/docs/how_to_guides/pushing_and_pulling/) via a small composite action such as [gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci)). This repo’s **default** path remains **single-step**: root YAML → in-memory pack → `oras.Copy`.

## Pinning, releases, and reproducibility

- **Pin the action** with a **semver tag** or **full commit SHA** on `uses:` (not floating `@main` in release pipelines). See [`examples/workflow-publish-with-pinned-action.yml`](examples/workflow-publish-with-pinned-action.yml).
- **SDK version** until [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) ships on **gemaraproj/go-gemara**: the **`require` / `replace`** block in [`tools/publish/go.mod`](tools/publish/go.mod) is the **source of truth** for which commit produced the bundle. When this action publishes **semver tags**, **release notes** should list the **go-gemara** version or pseudo-version validated for that tag.

## Operational notes and edge cases

- **401 / 403**: Indicates missing or insufficient registry credentials or repository permissions (for GHCR, the token needs **`packages: write`** where appropriate). The action does not print the password; avoid pasting full HTTP debug logs into tickets if they might contain tokens.
- **Network errors** mid-push: the step fails; registries may or may not expose a **fully atomic** publish—do not assume readers never see a partial state unless your registry documents stronger guarantees.
- **Large dependency graphs**: allow adequate **job timeout** and runner resources; unresolved fetches can slow or fail assembly. In caller workflows, set an explicit job timeout (for example **`timeout-minutes: 15`**—raise for very large graphs) on the job that runs this action.
- **Pre-release SDK `replace`**: development builds may point at a fork pseudo-version; **releases** should move to a **`require`** on an official **gemaraproj/go-gemara** tag once [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) is released.

## SDK version pin

Until [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) merges and ships as a tagged release, `tools/publish/go.mod` carries a **`replace`** to a fork pseudo-version (open that file for the exact line—avoid duplicating it in README so secret scanners stay quiet).

After merge, remove the `replace`, set `require` to the new **gemaraproj/go-gemara** release tag, and re-run `go mod tidy`.

## Usage

```yaml
permissions:
  contents: read
  packages: write

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: gemaraproj/gemara-bundle-publish-action@v1
        id: publish
        with:
          registry: ghcr.io
          repository: ${{ github.repository_owner }}/bundles/my-policy
          tag: v1.0.0
          file: policies/policy.yaml
          bundle_version: "1"
          gemara_version: v1.0.0
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Record digest
        run: echo "Published ${{ steps.publish.outputs.digest }}"
```

Replace `uses:` with `./path/to/repo` when testing from a clone. The root `file` is passed to `bundle.Assemble` (with `fetcher.URI`) so **extends** / **imports** resolved via mapping-reference URLs are included in the bundle.

### Remote content (mapping references)

`bundle.NewAssembler(&fetcher.URI{})` resolves **mapping-reference URLs** and other URI-backed references while assembling. Your CI job must have **network access** to those URLs at publish time, and the URLs must be reachable from the runner (self-hosted runners behind a firewall may need egress rules). Content is **fetched and embedded** according to the **SDK bundle contract**—the action does not redefine resolution rules.

## Outputs

| Name | Description |
|------|-------------|
| `digest` | Manifest digest for the pushed tag (for example `sha256:…`), after a successful copy. Suitable for immutable references and downstream jobs. |

## Inputs

| Name | Required | Description |
|------|----------|-------------|
| `registry` | yes | Registry host (e.g. `ghcr.io`). |
| `repository` | yes | Repository path without host (e.g. `org/bundles/name`). |
| `tag` | yes | Tag applied on the remote by `oras.Copy`. |
| `file` | yes | Root Gemara YAML path, relative to `working_directory`. |
| `username` | yes | Registry username. |
| `password` | yes | Registry password or token. |
| `validate` | no | If `true` (default), run `gemara.DetectType` + `gemara.Load` on the root file before assemble. |
| `bundle_version` | no | Stored in bundle manifest (`bundle-version`, default `1`). |
| `gemara_version` | no | Stored in bundle manifest (`gemara-version`, optional). |
| `working_directory` | no | Base directory for `file` (default `.`). |

## Registry compatibility

Bundles use the **Gemara bundle** OCI layout from PR #62 (`application/vnd.gemara.bundle.v1`, etc.). Registries or clients that only expect a **single** legacy layer (`application/vnd.gemara.catalog.v1+yaml`, …) may need separate alignment.

## Governance (#63)

Program work tracked in [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63) (outside this repo’s functional requirements): **identify action maintainers**, **land the repo** under the right GitHub org (for example **gemaraproj**), **[Marketplace publishing](https://docs.github.com/actions/creating-actions/publishing-actions-in-github-marketplace)**, and **tagged releases**. Follow that issue for status; this README documents the **technical** publish surface only.

## License

Apache-2.0 (aligned with go-gemara).
