# gemara-publish-oci

GitHub Action for Gemara OCI publish and trust orchestration. Phases (single `uses:`; toggles control each phase):

| Phase | What it does |
|--------|----------------|
| **1 — Publish** | Push to the **source** registry: `layout-copy` (ORAS from an OCI layout), `sdk` (your `gemara_binary`), or `gemara-file` (`tools/publish` + go-gemara). |
| **2 — Trust (source)** | Optional keyless **cosign** sign/verify on the **source** digest. |
| **3 — Promote (optional)** | Optional **ORAS** copy to Quay (or another registry) with `trust_mode`. |
| **4 — Trust (destination)** | Optional sign/verify on the **destination** digest when promotion and flags request it. |

**Outputs:** `digest` / `source_digest` and `source_ref` are always **normalized** to the form `sha256:<64 lowercase hex>` / `registry/repo@sha256:…` for stable downstream use.

**Who owns the contract:** [go-gemara](https://github.com/gemaraproj/go-gemara) owns bundle media types, manifest shape, and Pack/Assemble. This action does not.

### Boundaries (what is *not* in this action)

- **SLSA provenance**, **SBOM**, and **vulnerability** attestations typically produced with `actions/attest*` or org **reusable workflows** (for example in `complytime/org-infra`) are **out of scope** here—add them in **separate job steps** or a caller workflow *after* publish, if your policy needs them.
- **Branch / environment / approval** gates are **not** enforced inside the composite. Use **caller workflow** `if:` and GitHub **Environments** (see [Caller patterns](#caller-patterns) below).
- Composites only receive **string** inputs. Boolean-like flags are the literal strings **`"true"`** and **`"false"`** (see table below).

| Typical “boolean” input | Set to |
|------------------------|--------|
| `sign_source`, `verify_source`, `plain_http`, `promote_to_quay`, `sign_destination`, `verify_destination` | `"true"` or `"false"` |

### Caller patterns

Reusable **workflows** can set `concurrency:`, `environment:`, and `if: github.ref_protected` on a **job**. A **composite** cannot—so apply these on the **job** (or parent workflow) that invokes this action:

- **Concurrency:** add a `concurrency:` group on the publish job (for example keyed by `repository` + `tag` or by digest) so overlapping releases do not clobber one another.
- **Protected branches / tags:** use `if: github.ref_protected` (or your org’s rules) on the same job for production releases.
- **Environments (manual approval):** set `environment: production` (or similar) on the **job** so the step that uses this action runs under environment protection rules, matching how org-infra sign workflows use `sign_environment`.

## Publish modes (`publish_mode`)

| Value | Behavior |
|-------|----------|
| `layout-copy` | Copy `oci_layout_path:layout_ref` with `oras cp --from-oci-layout`. |
| `sdk` | Invoke `gemara_binary` + `sdk_args` and resolve digest from destination ref. |
| `gemara-file` | Pack the root Gemara YAML with **go-gemara** (`tools/publish`; same module as callers `go run`), then push to the registry (no nested action). |

## Promotion and trust

- Set `promote_to_quay: "true"` to run the standard GHCR -> Quay promotion path.
- Standard path defaults (no extra inputs needed):
  - `trust_mode: resign`
  - `sign_destination: "true"`
  - `verify_destination: "true"`
- Optional compatibility trust modes remain available:
  - `copy-only`: copy payload tag only.
  - `copy-referrers`: recursive copy to include referrer graph when registry support is available.
- Source and destination verification use Fulcio issuer
  `https://token.actions.githubusercontent.com`.
- This repository's CI validates source-publish and source-signing behavior. Cross-registry
  promotion verification is authoritative in caller repositories (for example
  `complytime-policies`) where Quay credentials and release controls live.

## Key inputs

| Input | Description |
|-------|-------------|
| `registry`, `repository`, `tag` | Source destination for publish. |
| `username`, `password` | Source registry auth. |
| `sign_source`, `verify_source` | Source signature controls. |
| `promote_to_quay` | Enable Quay promotion. |
| `quay_registry`, `quay_image`, `quay_tag`, `quay_username`, `quay_password` | Promotion destination/auth. |
| `trust_mode` | `copy-only`, `copy-referrers`, or `resign`. |
| `verify_destination` | Destination signature verification control. |
| `allowed_identity_regex` | Optional cosign identity regex override. |

## Outputs

| Output | Description |
|--------|-------------|
| `digest` / `source_digest` | Source manifest digest, **always** `sha256:` + 64 hex (lowercase). |
| `source_ref` | `registry/repository@sha256:…` (digest normalized as above). |
| `destination_digest` | Destination digest after promotion. |
| `destination_ref` | Destination image reference with digest. |
| `verified_source` | `true` if source verify passed. |
| `verified_destination` | `true` if destination verify passed. |
| `trust_mode` | Effective trust mode used. |

## Minimal caller example (Option 3)

```yaml
permissions:
  contents: read
  packages: write
  id-token: write

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - id: publish
        uses: complytime/oci-artifact@<pinned-sha>
        with:
          publish_mode: gemara-file
          registry: ghcr.io
          repository: ${{ github.repository }}
          tag: ${{ github.ref_name }}
          file: bundles/cis-fedora-l1-workstation.yaml
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
          promote_to_quay: "true"
          quay_image: continuouscompliance/complytime-policies
          quay_username: ${{ secrets.QUAY_ROBOT_USERNAME }}
          quay_password: ${{ secrets.QUAY_ROBOT_TOKEN }}
```

## Repository layout

- **`tools/publish/`** — small Go CLI used **only** when `publish_mode: gemara-file`. It wires `bundle.Assemble` / `bundle.Pack` and `oras.Copy`; SDK semantics stay in **go-gemara** `v0.4.0+`.
- **`testdata/minimal-layout/`** — tiny OCI layout for **this repo’s CI** (`layout-copy`). It is not part of the action runtime for consumers.

## Pinning

Use a full commit SHA for production callers. Avoid floating refs.

## License

See [LICENSE](LICENSE).
