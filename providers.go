package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StaleProvider represents a provider version that should be removed.
type StaleProvider struct {
	Source  string // e.g. "registry.terraform.io/hashicorp/aws"
	Version string // e.g. "6.30.0"
	Path    string // absolute path to the version directory
}

// FindStaleProviders scans the .terraform/providers/ directory and returns
// provider versions that are not in the lock file.
// If filterProviders is non-empty, only those providers are considered.
func FindStaleProviders(providersDir string, locked map[string]string, filterProviders []string) ([]StaleProvider, error) {
	// Build filter set: use short form "hashicorp/aws" for matching
	filterSet := make(map[string]bool, len(filterProviders))
	for _, p := range filterProviders {
		filterSet[p] = true
	}

	var stale []StaleProvider

	// Walk: registry/namespace/type/version
	// e.g. registry.terraform.io/hashicorp/aws/6.34.0
	registries, err := readDirNames(providersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read providers directory: %w", err)
	}

	for _, registry := range registries {
		registryPath := filepath.Join(providersDir, registry)
		namespaces, err := readDirNames(registryPath)
		if err != nil {
			continue
		}

		for _, namespace := range namespaces {
			namespacePath := filepath.Join(registryPath, namespace)
			types, err := readDirNames(namespacePath)
			if err != nil {
				continue
			}

			for _, providerType := range types {
				source := fmt.Sprintf("%s/%s/%s", registry, namespace, providerType)
				shortName := fmt.Sprintf("%s/%s", namespace, providerType)

				// Apply provider filter
				if len(filterSet) > 0 && !filterSet[shortName] && !filterSet[source] {
					continue
				}

				typePath := filepath.Join(namespacePath, providerType)
				versions, err := readDirNames(typePath)
				if err != nil {
					continue
				}

				lockedVersion, isLocked := locked[source]

				for _, version := range versions {
					if isLocked && version == lockedVersion {
						continue
					}
					stale = append(stale, StaleProvider{
						Source:  source,
						Version: version,
						Path:    filepath.Join(typePath, version),
					})
				}
			}
		}
	}

	return stale, nil
}

// DirSize calculates the total size of a directory in bytes.
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// FormatSize formats bytes into a human-readable string.
func FormatSize(bytes int64) string {
	const (
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// ShortSource returns the namespace/type portion of a provider source.
// e.g. "registry.terraform.io/hashicorp/aws" -> "hashicorp/aws"
func ShortSource(source string) string {
	parts := strings.SplitN(source, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return source
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
