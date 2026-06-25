# Release Process

This document describes how to create a new release of Hatch using the CI/CD pipeline.

## Overview

Hatch uses GitHub Actions to automatically build cross-platform binaries and publish them to GitHub Releases when a version tag is pushed. The pipeline builds binaries for:

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

Each binary is accompanied by a SHA256 checksum file for verification.

## Prerequisites

- Write access to the Hatch repository
- Understanding of [semantic versioning](https://semver.org/)
- Ability to push tags to the main branch

## Release Steps

### 1. Prepare the Release

Before creating a tag, ensure:

- All changes for the release are merged to `main`
- The `CHANGELOG.md` is updated with the new version's changes (if applicable)
- All CI checks pass on `main`

### 2. Create and Push a Tag

```bash
# Ensure you're on the latest main
git checkout main
git pull origin main

# Create a semantic version tag
git tag v1.0.0

# Push the tag to GitHub
git push origin v1.0.0
```

The tag **must** follow the pattern `v*` (e.g., `v1.0.0`, `v1.2.3-beta.1`).

### 3. Monitor the Build

After pushing the tag:

1. Go to the [Actions tab](https://github.com/elfoundation/hatch/actions) of the repository
2. Find the workflow run for the "Release" workflow
3. Monitor the build progress for each platform

The workflow will:
- Build binaries for all platforms in parallel
- Generate SHA256 checksums
- Create a GitHub Release with release notes
- Upload all binaries and checksums as release assets

### 4. Verify the Release

Once the workflow completes:

1. Check the [Releases page](https://github.com/elfoundation/hatch/releases) for the new release
2. Verify all binaries are present:
   - `hatch-linux-amd64` + `hatch-linux-amd64.sha256`
   - `hatch-linux-arm64` + `hatch-linux-arm64.sha256`
   - `hatch-darwin-amd64` + `hatch-darwin-amd64.sha256`
   - `hatch-darwin-arm64` + `hatch-darwin-arm64.sha256`
   - `hatch-windows-amd64.exe` + `hatch-windows-amd64.exe.sha256`
3. Download a binary and verify it works:
   ```bash
   # Example for Linux amd64
   ./hatch-linux-amd64 --version
   ```

### 5. Update Documentation (if needed)

If the release includes significant changes, update:

- `README.md` (if installation instructions change)
- `CHANGELOG.md` (if not already updated)
- Any relevant documentation in `docs/`

## Manual Trigger

The release workflow can also be triggered manually via GitHub Actions UI:

1. Go to Actions → Release → Run workflow
2. Enter the release tag (e.g., `v1.0.0`)
3. Click "Run workflow"

**Note:** Manual triggers still require the tag to exist in the repository.

## Troubleshooting

### Build Fails

- Check the [Actions logs](https://github.com/elfoundation/hatch/actions) for error details
- Ensure the tag follows the `v*` pattern
- Verify Go version compatibility (currently Go 1.25)

### Missing Binaries

- Check if all matrix jobs completed successfully
- Verify the artifact upload steps completed without errors

### Release Notes Not Generated

- The workflow attempts to extract notes from `CHANGELOG.md`
- If no notes are found, a default template is used
- You can manually edit the release notes on the Releases page

### Checksum Verification

To verify a binary's integrity:

```bash
# Linux/macOS
sha256sum -c hatch-linux-amd64.sha256

# Windows (PowerShell)
Get-FileHash hatch-windows-amd64.exe -Algorithm SHA256
```

## Architecture Decisions

### Why Cross-Platform Builds?

Hatch is designed to be used by developers on various platforms. Providing pre-built binaries eliminates the need for users to install Go and build from source.

### Why SHA256 Checksums?

Checksums allow users to verify that downloaded binaries haven't been tampered with and match the official release.

### Why Semantic Versioning?

Semantic versioning communicates the nature of changes (major, minor, patch) to users and dependency managers.

## Related Documents

- [CLI Reference](cli.md) - Command-line interface documentation
- [Local Development](local-dev.md) - Setting up a development environment
- [Architecture](hatch-architecture.md) - System design overview