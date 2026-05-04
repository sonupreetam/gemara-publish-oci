// SPDX-License-Identifier: Apache-2.0

// Command publish implements go-gemara bundle Assemble → Pack → oras.Copy to a registry
// (see gemaraproj/go-gemara bundle APIs and OCI distribution). This entrypoint is invoked
// from the composite GitHub Action so callers do not maintain a separate publish binary.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gemaraproj/go-gemara"
	"github.com/gemaraproj/go-gemara/bundle"
	"github.com/gemaraproj/go-gemara/fetcher"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func main() {
	registry := flag.String("registry", "", "registry host (e.g. ghcr.io)")
	repository := flag.String("repository", "", "repository path without host (e.g. org/bundles/my-policy)")
	tag := flag.String("tag", "", "tag to apply on the remote repository")
	file := flag.String("file", "", "absolute path to the root Gemara YAML file")
	username := flag.String("username", "", "registry username")
	bundleVersion := flag.String("bundle-version", "1", "bundle manifest bundle-version")
	gemaraVersion := flag.String("gemara-version", "", "bundle manifest gemara-version (optional)")
	validate := flag.Bool("validate", true, "run gemara.Load-style validation before assemble")
	flag.Parse()

	password := os.Getenv("GEMARA_REGISTRY_PASSWORD")
	if *registry == "" || *repository == "" || *tag == "" || *file == "" || *username == "" || password == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "GEMARA_REGISTRY_PASSWORD must be set.")
		os.Exit(2)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read root file: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if *validate {
		if err := validateRoot(ctx, *file, data); err != nil {
			fmt.Fprintf(os.Stderr, "validate: %v\n", err)
			os.Exit(1)
		}
	}

	src := bundle.File{Name: filepath.Base(*file), Data: data}
	m := bundle.Manifest{BundleVersion: *bundleVersion, GemaraVersion: *gemaraVersion}
	asm := bundle.NewAssembler(&fetcher.URI{})
	b, err := asm.Assemble(ctx, m, src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble: %v\n", err)
		os.Exit(1)
	}

	store := memory.New()
	desc, err := bundle.Pack(ctx, store, b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pack: %v\n", err)
		os.Exit(1)
	}

	repoRef := fmt.Sprintf("%s/%s", *registry, *repository)
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "remote repo: %v\n", err)
		os.Exit(1)
	}
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(*registry, auth.Credential{
			Username: *username,
			Password: password,
		}),
	}

	copyOpts := oras.DefaultCopyOptions
	const localRootRef = "gemara-publish/__root__"
	srcRef := desc.Digest.String()
	if _, err := store.Resolve(ctx, srcRef); err != nil {
		if tagErr := store.Tag(ctx, desc, localRootRef); tagErr != nil {
			fmt.Fprintf(os.Stderr, "resolve %q in local store: %v; tag %s: %v\n", srcRef, err, localRootRef, tagErr)
			os.Exit(1)
		}
		srcRef = localRootRef
	}
	if _, err := oras.Copy(ctx, store, srcRef, repo, *tag, copyOpts); err != nil {
		fmt.Fprintf(os.Stderr, "oras copy (from %s): %v\n", srcRef, err)
		os.Exit(1)
	}

	manifestDesc, err := repo.Resolve(ctx, *tag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve remote tag %q after copy: %v\n", *tag, err)
		os.Exit(1)
	}
	digestStr := manifestDesc.Digest.String()
	if err := writeGitHubOutput("digest", digestStr); err != nil {
		fmt.Fprintf(os.Stderr, "write GITHUB_OUTPUT: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("digest=%s\n", digestStr)
}

func writeGitHubOutput(key, value string) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func validateRoot(ctx context.Context, path string, data []byte) error {
	t, err := gemara.DetectType(data)
	if err != nil {
		return fmt.Errorf("detect type: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".gemara-validate-*.yaml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()        //nolint:errcheck,gosec
		os.Remove(tmpPath) //nolint:errcheck
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath) //nolint:errcheck
		return err
	}
	defer os.Remove(tmpPath) //nolint:errcheck

	f := &fetcher.File{}
	switch t {
	case gemara.PolicyArtifact:
		_, err = gemara.Load[gemara.Policy](ctx, f, tmpPath)
	case gemara.GuidanceCatalogArtifact:
		_, err = gemara.Load[gemara.GuidanceCatalog](ctx, f, tmpPath)
	default:
		_, err = gemara.Load[gemara.ControlCatalog](ctx, f, tmpPath)
	}
	return err
}
