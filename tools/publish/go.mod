module github.com/sonupreetam/gemara-publish-oci/tools/publish

// PR #62 (bundle Pack/Assemble + oras-go): gemaraproj/go-gemara@v0.3.0 and @main
// (checked 2026-04-21) do not ship github.com/gemaraproj/go-gemara/bundle yet.
// Pin this pseudo-version from github.com/jpower432/go-gemara until a gemaraproj
// tag contains the bundle package (tasks.md T020).
go 1.25.0

toolchain go1.25.8

require (
	github.com/gemaraproj/go-gemara v0.3.0
	oras.land/oras-go/v2 v2.6.0
)

require (
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

replace github.com/gemaraproj/go-gemara => github.com/jpower432/go-gemara v0.0.0-20260418000148-0d0e23202fa1
