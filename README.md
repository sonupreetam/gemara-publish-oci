# gemara-publish-oci (bundle publish)

GitHub Action: publish **Gemara OCI bundles** with the **go-gemara** SDK ([PR #62](https://github.com/gemaraproj/go-gemara/pull/62) `bundle` APIs + **oras-go**), in line with [go-gemara issue #63](https://github.com/gemaraproj/go-gemara/issues/63) and the OCI work in [go-gemara #60](https://github.com/gemaraproj/go-gemara/issues/60).

**SDK owns the contract.** Manifest shape, **media types**, and **layer assembly** are defined in **go-gemara** (`bundle.Pack`, `bundle.Unpack`, …). This repository supplies the **composite action** ([`action.yml`](action.yml)), [`tools/publish`](tools/publish) (`go run` on the runner), and CI. The Go **module** for the tool is **`github.com/sonupreetam/gemara-publish-oci/tools/publish`**.

### Shippable today

1. **Composite action** at repo root — `go run` of [`tools/publish`](tools/publish); inputs as below; SDK pin in [`tools/publish/go.mod`](tools/publish/go.mod).
2. **Optional tool image** ([`examples/Dockerfile.publish.sketch`](examples/Dockerfile.publish.sketch) + [`examples/workflow-publish-with-docker-image.yml`](examples/workflow-publish-with-docker-image.yml)).

See [`examples/README.md`](examples/README.md) for a short backlog and workflow copy-paste points.

### Behaviour (technical)

**Assemble** → **pack** into an **in-memory** OCI store → **`oras.Copy`** to the registry. Resolve by manifest **digest** first; if that fails, **tag** the root as `gemara-publish/__root__` and **copy** from that ref. No ORAS **CLI** for manifest/media types—those stay in **`bundle.Pack`**.

## End-to-end verification

CI runs **`e2e-publish-ghcr`** in [`.github/workflows/ci.yml`](.github/workflows/ci.yml) on push and same-repo pull requests. Optional manual: [`.github/workflows/e2e-ghcr.yml`](.github/workflows/e2e-ghcr.yml) (`workflow_dispatch`).

## Two-phase publish (layout on disk, transport only)

If you only need **`oras cp --from-oci-layout`** from an existing on-disk layout, you can use a **separate** ORAS-based workflow or an **older ref** of this repository from before the bundle migration (check git history). The **default** in this branch is the **single-step** bundle path: root YAML → in-memory pack → `oras.Copy`.

## Pinning, releases, and reproducibility

- **Pin the action** with a **semver tag** or **commit SHA** on `uses:` (not `@main` in production). See [`examples/workflow-publish-with-pinned-action.yml`](examples/workflow-publish-with-pinned-action.yml).
- **SDK** until [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) is released on `gemaraproj/go-gemara`: the **`require` / `replace`** in [`tools/publish/go.mod`](tools/publish/go.mod) is the **source of truth** for the fork pin.

## SDK version pin

If `go.mod` carries a **`replace`** to a fork pseudo-version, use that only until an official `gemaraproj/go-gemara` tag contains `bundle` APIs. Then remove `replace`, update `require`, and run `go mod tidy` in `tools/publish/`.

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

      - uses: sonupreetam/gemara-publish-oci@v1
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

Use `uses: ./` when testing from a checkout of this repository.

## Outputs

| Name     | Description |
|----------|-------------|
| `digest` | Root manifest digest after push (e.g. `sha256:…`). |

## Inputs

| Name                | Required | Description |
|---------------------|----------|-------------|
| `registry`          | yes      | Host (e.g. `ghcr.io`). |
| `repository`        | yes      | Path without host. |
| `tag`               | yes      | Remote tag. |
| `file`              | yes      | Root Gemara YAML (relative to `working_directory`). |
| `username` / `password` | yes  | Registry auth. |
| `validate`          | no       | `gemara.Load` before assemble (default `true`). |
| `bundle_version`    | no       | `bundle-version` in manifest. |
| `gemara_version`    | no       | `gemara-version` in manifest. |
| `working_directory` | no       | Base for `file` (default `.`). |

## License

Apache-2.0 (see [`LICENSE`](LICENSE), aligned with go-gemara).
