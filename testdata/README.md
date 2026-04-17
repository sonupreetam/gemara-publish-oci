# Test fixtures

## `minimal-layout/`

Minimal **OCI image layout** (valid `index.json` + `blobs/`) used by CI to exercise `oras cp --from-oci-layout`.

Generated once with ORAS v1.2.0:

1. Run a local registry: `docker run -d -p 5000:5000 registry:2`
2. From the repo root, create a tiny file and push:
   `oras push localhost:5000/test/sample:v1 --plain-http ./path/to/file.txt:application/vnd.oci.image.layer.v1.tar+gzip`
3. Copy into a layout:
   `oras cp --from-plain-http localhost:5000/test/sample:v1 --to-oci-layout ./testdata/minimal-layout:v1`

The reference inside the layout is **`v1`** (see `org.opencontainers.image.ref.name` in `index.json`). Use `layout_ref: v1` with this fixture.
