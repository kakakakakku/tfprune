# tfprune

A CLI tool to remove old Terraform provider versions from `.terraform/providers/` directory.

Terraform does not clean up old provider versions when you upgrade. Over time, this can consume significant disk space. `tfprune` compares the installed providers against `.terraform.lock.hcl` and removes versions that are no longer needed.

## Install

**homebrew tap:**

```sh
$ brew install kakakakakku/tap/tfprune
```

**go install:**

```sh
$ go install github.com/kakakakakku/tfprune@latest
```

## Usage

Run in a Terraform project directory:

```sh
$ tfprune
```

Or specify a path:

```sh
$ tfprune -path /path/to/terraform/project
```

Limit to a specific provider:

```sh
$ tfprune -provider hashicorp/aws
```

Limit to multiple providers:

```sh
$ tfprune -provider hashicorp/aws -provider hashicorp/awscc
```

### Flags

| Flag | Description |
|---|---|
| `-dry-run` | Show what would be removed (no changes made) |
| `-path` | Path to Terraform project directory (default: `.`) |
| `-provider` | Limit to specific provider (e.g. `hashicorp/aws`). Can be specified multiple times |
| `-v` | Show version |
| `-y` | Skip confirmation prompt |
