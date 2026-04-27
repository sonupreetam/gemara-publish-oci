# gemara-publish-oci

GitHub Action for standardized Gemara OCI publishing:

1. Publish to a source registry (GHCR or compatible).
2. Keyless sign + verify source digest.
3. Optionally promote to Quay.
4. Verify trust on destination using the standard re-sign model (default).

The action still does not define Gemara bundle semantics; manifest/media-type ownership remains in
[go-gemara](https://github.com/gemaraproj/go-gemara).

## Publish modes (`publish_mode`)

| Value | Behavior |
|-------|----------|
| `layout-copy` | Copy `oci_layout_path:layout_ref` with `oras cp --from-oci-layout`. |
| `sdk` | Invoke `gemara_binary` + `sdk_args` and resolve digest from destination ref. |
| `gemara-file` | Delegate file-based pack+push (`file`, `validate`, `bundle_version`, `working_directory`) to the pinned compatibility action. |

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
| `digest` / `source_digest` | Source digest. |
| `source_ref` | Source image reference with digest. |
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

## Pinning

Use a full commit SHA for production callers. Avoid floating refs.

## License

See [LICENSE](LICENSE).
