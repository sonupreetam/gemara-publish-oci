# gemara-publish-oci

GitHub Action for **transport only**: push a Gemara OCI artifact to a container registry. **Pack, manifest shape, and media types** belong in **[go-gemara](https://github.com/gemaraproj/go-gemara)** ([issue #60](https://github.com/gemaraproj/go-gemara/issues/60)). This repo is intentionally small: **`action.yml` only** (one composite `run` step; no extra scripts) — see **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**. For **maintainer review** (scope, SDK split, open questions), see **[docs/DESIGN-FOR-REVIEW.md](docs/DESIGN-FOR-REVIEW.md)**.

## Principles (SDK vs. this action)

- **SDK** = Pack / Unpack + `oras.Copy` semantics (what gets pushed).
- **Action** = registry auth + run a **released** CLI (ORAS today; **`gemara`** when available) + expose **`digest`**.
- **Do not** encode layer-level `oras push` assembly here; that stays in the SDK or org-infra until the contract is stable.

## Modes (`publish_mode`)

| Value | Behavior |
|-------|----------|
| **`layout-copy`** (default) | Installs ORAS and runs **`oras cp --from-oci-layout`** to the registry. Use while Pack exports an **OCI image layout** on disk (same family of layout **complyctl** caches — see [001 research](https://github.com/complytime/complyctl/blob/main/specs/001-gemara-native-workflow/research.md)). |
| **`sdk`** | Runs **`gemara_binary`** with **`sdk_args`** after `oras login`. Use once **go-gemara** ships a stable push CLI so **one codebase** owns Pack + `oras.Copy`. Exports **`GEMARA_REGISTRY`**, **`GEMARA_REPOSITORY`**, **`GEMARA_TAG`** for the CLI. |

For **direct multi-file `oras push`** from repo trees (pre-Pack), see [complytime-policies OCI spec](https://github.com/complytime/complytime-policies/blob/main/docs/oci-publish-spec.md) and [org-infra#172](https://github.com/complytime/org-infra/issues/172).

## Inputs (summary)

| Input | Description |
|-------|-------------|
| `publish_mode` | `layout-copy` or `sdk` |
| `registry` | Registry host (default `ghcr.io`) |
| `repository` | Path without host (e.g. `org/name`) |
| `tag` | Tag to push |
| `oci_layout_path` / `pack_path` | OCI layout dir (`layout-copy` only) |
| `layout_ref` | Layout reference for `oras cp PATH:REF` (`layout-copy` only, **required** in that mode) |
| `gemara_binary` | Path or name on PATH (`sdk` only, **required** in that mode) |
| `sdk_args` | Arguments for the gemara CLI (`sdk` only) |
| `username` | Registry user; defaults to `GITHUB_ACTOR` when using `password` |
| `password` | Token; omit only with `plain_http: true` |
| `oras_version` | ORAS version for `layout-copy` and digest resolution (default `1.2.0`) |
| `plain_http` | `true` for HTTP registries (e.g. local CI) |

## Outputs

| Output | Description |
|--------|-------------|
| `digest` | Manifest digest (`sha256:...`) |

## Usage — `layout-copy` (GHCR)

```yaml
permissions:
  contents: read
  packages: write

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: OWNER/gemara-publish-oci@v0.1.0
        id: publish
        with:
          publish_mode: layout-copy
          registry: ghcr.io
          repository: ${{ github.repository }}
          tag: sha-${{ github.sha }}
          oci_layout_path: ./layout
          layout_ref: v1
          password: ${{ secrets.GITHUB_TOKEN }}
```

## Usage — `sdk` (when `gemara` CLI exists)

```yaml
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: go install github.com/gemaraproj/go-gemara/cmd/gemara@VERSION   # example — TBD
      - uses: OWNER/gemara-publish-oci@v0.1.0
        id: publish
        with:
          publish_mode: sdk
          gemara_binary: gemara
          sdk_args: publish oci
          registry: ghcr.io
          repository: ${{ github.repository }}
          tag: sha-${{ github.sha }}
          password: ${{ secrets.GITHUB_TOKEN }}
```

Subcommand and flags for **`sdk_args`** must match **go-gemara**; the snippet is illustrative until a release documents them.

## Pinning

Use **`@v0.1.0`** or a **full commit SHA**, not `@main`. In the examples above, replace **`OWNER`** with the GitHub org or user that hosts this repository (for example `gemaraproj` after transfer).

## License

See [LICENSE](LICENSE).
