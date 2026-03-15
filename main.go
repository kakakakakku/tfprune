package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var version = "dev"

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var (
		dryRun    bool
		yes       bool
		showVer   bool
		path      string
		providers stringSlice
	)

	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be removed without deleting")
	flag.BoolVar(&yes, "y", false, "Skip confirmation prompt")
	flag.BoolVar(&showVer, "v", false, "Show version")
	flag.StringVar(&path, "path", ".", "Path to Terraform project directory")
	flag.Var(&providers, "provider", "Limit to specific provider (e.g. hashicorp/aws). Can be specified multiple times")
	flag.Parse()

	if showVer {
		fmt.Printf("tfprune %s\n", version)
		return
	}

	if err := run(path, dryRun, yes, providers); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(basePath string, dryRun, yes bool, filterProviders []string) error {
	lockFilePath := filepath.Join(basePath, ".terraform.lock.hcl")
	providersDir := filepath.Join(basePath, ".terraform", "providers")

	// Check that required files/dirs exist
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		return fmt.Errorf("%s not found. Run this command in a Terraform project directory", lockFilePath)
	}
	if _, err := os.Stat(providersDir); os.IsNotExist(err) {
		return fmt.Errorf("%s not found. Run 'terraform init' first", providersDir)
	}

	// Parse lock file
	locked, err := ParseLockFile(lockFilePath)
	if err != nil {
		return err
	}

	// Find stale providers
	stale, err := FindStaleProviders(providersDir, locked, filterProviders)
	if err != nil {
		return err
	}

	if len(stale) == 0 {
		fmt.Println("No stale providers found. Everything is clean!")
		return nil
	}

	// Calculate sizes and display
	var totalSize int64
	fmt.Println("The following provider versions will be removed:")
	fmt.Println()
	for _, s := range stale {
		size, _ := DirSize(s.Path)
		totalSize += size
		fmt.Printf("  - %s %s (%s)\n", ShortSource(s.Source), s.Version, FormatSize(size))
	}
	fmt.Println()
	fmt.Printf("Total: %d provider version(s), %s\n", len(stale), FormatSize(totalSize))

	if dryRun {
		return nil
	}

	// Confirmation prompt
	if !yes {
		fmt.Print("\nProceed? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Delete
	var removedSize int64
	for _, s := range stale {
		size, _ := DirSize(s.Path)
		if err := os.RemoveAll(s.Path); err != nil {
			fmt.Fprintf(os.Stderr, "  Failed to remove %s: %v\n", s.Path, err)
			continue
		}
		removedSize += size
		fmt.Printf("  Removed %s %s\n", ShortSource(s.Source), s.Version)
	}

	fmt.Printf("\nDone! Freed %s\n", FormatSize(removedSize))
	return nil
}
