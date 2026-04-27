# Architecture: SDK-owned semantics, action-owned publish orchestration

This repository is **not** the source of truth for Gemara OCI semantics. That belongs in
**[go-gemara](https://github.com/gemaraproj/go-gemara)** and
**[go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60)** (manifest shape, media
types, Pack/Unpack, `oras.Copy` from packed store to registry).

## Intended split (upstream: go-gemara)

Conceptually:

1. **`bundle.Pack`** (SDK) produces an OCI store / artifact.
2. **`oras.Copy`** (SDK, via `oras-go/v2`) moves that artifact to a **remote registry** with a tag.
3. **`bundle.Unpack`** (SDK) on the consumer pulls and reads it.

The GitHub Action provides what CI needs as a standardized contract in **`action.yml`**:

- Secrets and **`oras login`** for source and destination registries.
- Publish entrypoints (`layout-copy`, `sdk`, `gemara-file` compatibility mode).
- Keyless cosign sign/verify for source and destination digests.
- Standard GHCR -> Quay promotion with destination re-sign trust (defaults).
- Structured `GITHUB_OUTPUT` values for source/destination refs and verification state.

## Publish and promotion model

| Concern | In action | Notes |
|---------|-----------|-------|
| Publish source | `layout-copy`, `sdk`, `gemara-file` | `gemara-file` bridges callers that only have root bundle YAML. |
| Source trust | `sign_source`, `verify_source` | Keyless cosign against source digest. |
| Promotion | `promote_to_quay` | Copies source tag to destination tag. |
| Destination trust | `trust_mode` + destination sign/verify toggles | Standard path uses `resign`; `copy-referrers` remains optional compatibility mode. |

## Why this keeps SDK boundaries intact

This action orchestrates publishing steps but does not redefine layer descriptors or media types.
`gemara-file` mode delegates pack/push to a compatibility action until SDK-led publish interfaces are
fully standardized.

## Compliance

- **Do not** add Gemara YAML schema ownership or layer `mediaType` tables here — use **go-gemara**.
- **Do** pin versions (`oras_version`, `cosign_version`, and action refs) and document migration.
- Cross-registry verification is owned by caller workflows that hold destination credentials
  (for example `complytime-policies`), not by this repository's CI baseline.
