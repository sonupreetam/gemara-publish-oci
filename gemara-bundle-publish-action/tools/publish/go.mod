module github.com/gemaraproj/gemara-bundle-publish-action/tools/publish

// PR #62 (bundle Pack/Assemble + oras-go) is not merged on main yet; pin this
// pseudo-version from github.com/jpower432/go-gemara branch feat/add-bundle-types.
// After merge, replace with a tagged github.com/gemaraproj/go-gemara release.
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
