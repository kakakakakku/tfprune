package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type LockFile struct {
	Providers []ProviderLock `hcl:"provider,block"`
}

type ProviderLock struct {
	Source  string `hcl:"source,label"`
	Version string `hcl:"version"`

	// Remaining fields (constraints, hashes) are not needed but must be captured
	Remain hcl.Body `hcl:",remain"`
}

// ParseLockFile parses a .terraform.lock.hcl file and returns a map of provider source to version.
// Example: "registry.terraform.io/hashicorp/aws" -> "6.34.0"
func ParseLockFile(path string) (map[string]string, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse lock file: %s", diags.Error())
	}

	var lockFile LockFile
	diags = gohcl.DecodeBody(file.Body, nil, &lockFile)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode lock file: %s", diags.Error())
	}

	providers := make(map[string]string, len(lockFile.Providers))
	for _, p := range lockFile.Providers {
		providers[p.Source] = p.Version
	}
	return providers, nil
}
