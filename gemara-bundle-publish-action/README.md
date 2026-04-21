# Gemara bundle publish (GitHub Action)

Workstream for [go-gemara issue #63](https://github.com/gemaraproj/go-gemara/issues/63): publish **Gemara OCI bundles** using the **SDK today** ([PR #62](https://github.com/gemaraproj/go-gemara/pull/62) `bundle` APIs + **oras-go**), **without** blocking on a future upstream `gemara` CLI.

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

## SDK version pin

Until [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) merges and ships as a tagged release, `tools/publish/go.mod` uses:

```text
replace github.com/gemaraproj/go-gemara => github.com/jpower432/go-gemara v0.0.0-20260418000148-0d0e23202fa1
```

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
        with:
          registry: ghcr.io
          repository: ${{ github.repository_owner }}/bundles/my-policy
          tag: v1.0.0
          file: policies/policy.yaml
          bundle_version: "1"
          gemara_version: v1.0.0
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
```

Replace `uses:` with `./path/to/repo` when testing from a clone. The root `file` is passed to `bundle.Assemble` (with `fetcher.URI`) so **extends** / **imports** resolved via mapping-reference URLs are included in the bundle.

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

## Marketplace and ownership

See the [issue #63](https://github.com/gemaraproj/go-gemara/issues/63) checklist: dedicated **gemaraproj** repo, tagged releases, [Marketplace publishing](https://docs.github.com/actions/creating-actions/publishing-actions-in-github-marketplace), and named maintainers.

## License

Apache-2.0 (aligned with go-gemara).
