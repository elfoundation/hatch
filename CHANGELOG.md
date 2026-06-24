# Changelog

All notable changes to Hatch will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CLI documentation with comprehensive command reference
- GitHub Actions release workflow for binary builds
- Standalone binaries for Linux, macOS, and Windows
- SHA256 checksums for all binaries
- SBOM generation for releases

### Changed
- Updated README with CLI installation instructions

### Fixed
- N/A

## [0.1.0] - 2024-01-15

### Added
- Initial release of Hatch
- HTTP request capture and storage
- Request inspection and search
- Request replay functionality
- Mock response configuration
- OpenAPI documentation generation
- Web UI for visual inspection
- Docker support with Caddy reverse proxy
- SQLite storage backend
- REST API v1

### Changed
- N/A

### Fixed
- N/A

## [0.0.1] - 2024-01-01

### Added
- Project scaffolding
- Basic architecture design
- Development environment setup

### Changed
- N/A

### Fixed
- N/A

---

## Release Process

### Creating a Release

1. **Update CHANGELOG.md** with the new version and changes
2. **Create a git tag** for the version:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. **GitHub Actions will automatically**:
   - Build binaries for all platforms
   - Create SHA256 checksums
   - Generate SBOM
   - Create GitHub Release with assets

### Manual Release

For manual releases or testing:

```bash
# Trigger workflow manually
gh workflow run release.yml -f tag=v1.0.0

# Or create tag and push
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Platform Support

| Platform | Architecture | Binary Name |
|----------|--------------|-------------|
| Linux | x86_64 | `hatch-linux-amd64` |
| Linux | ARM64 | `hatch-linux-arm64` |
| macOS | Intel | `hatch-darwin-amd64` |
| macOS | Apple Silicon | `hatch-darwin-arm64` |
| Windows | x86_64 | `hatch-windows-amd64.exe` |

### Binary Installation

After downloading:

```bash
# Linux/macOS
chmod +x hatch-*
sudo mv hatch-* /usr/local/bin/hatch

# Windows (PowerShell)
Rename-Item hatch-windows-amd64.exe hatch.exe
```

### Verification

Verify binary integrity using SHA256 checksums:

```bash
# Linux/macOS
sha256sum -c hatch-linux-amd64.sha256

# Windows (PowerShell)
Get-FileHash hatch-windows-amd64.exe -Algorithm SHA256
```