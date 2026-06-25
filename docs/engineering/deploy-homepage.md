# Deploying the Hatch Homepage

This document describes how the Hatch static homepage is deployed to hatch.surf.

## Overview

The homepage is deployed to two locations:

1. **GitHub Pages** - Serves as a fallback and for GitHub-based discovery
2. **hatch.surf server** - Primary deployment via Docker at https://hatch.surf

## Architecture

```
site/
├── index.html          # Main homepage
├── style.css           # Stylesheet
├── main.js             # JavaScript
├── brand/              # Logo and favicon assets
│   ├── favicon/
│   └── og/
├── blog/               # Blog directory
├── Dockerfile          # nginx:alpine with static files baked in
└── docker-compose.yml  # Local development
```

### Docker Deployment Flow

```
GitHub Push → GitHub Actions → Docker Build → SCP Image → Docker Run → Caddy (reverse proxy)
```

- Static files are **baked into the Docker image** during build
- No host directory mounting (`/var/www`) required
- Image is self-contained and portable

## Server Setup

### Prerequisites

- Server: 46.250.250.48 (hatch.surf)
- Docker installed
- Caddy for reverse proxy and TLS
- SSH access for deployment

### Container Configuration

- **Container name**: `hatch-homepage`
- **Image**: `hatch-homepage:latest`
- **Internal port**: 80 (nginx)
- **Exposed port**: 127.0.0.1:3000 (proxied by Caddy)

## Deployment

### Automated (GitHub Actions)

When changes are pushed to `main` branch affecting `site/` directory:

1. **GitHub Pages** - Deployed automatically via GitHub Actions
2. **Server** - Docker image built, uploaded, and deployed

Required GitHub secrets:
- `DEPLOY_HOST` - Server IP (46.250.250.48)
- `DEPLOY_USER` - SSH username (root)
- `DEPLOY_KEY` - SSH private key (base64 encoded)

### Manual Deployment

Use the deployment script:

```bash
# Dry run
./scripts/deploy-site.sh --dry-run

# Deploy
DEPLOY_HOST=46.250.250.48 \
DEPLOY_USER=root \
DEPLOY_KEY=$(cat ~/.ssh/id_rsa | base64) \
./scripts/deploy-site.sh
```

### Docker Commands (on server)

```bash
# Build image locally
docker build -t hatch-homepage:latest ./site

# Run container
docker run -d \
  --name hatch-homepage \
  --restart unless-stopped \
  -p 127.0.0.1:3000:80 \
  hatch-homepage:latest

# Check status
docker ps | grep hatch-homepage

# View logs
docker logs hatch-homepage

# Restart
docker restart hatch-homepage
```

## Verification

### Check Site Status

```bash
# HTTP redirect
curl -s -o /dev/null -w "%{http_code}" http://hatch.surf/

# HTTPS site
curl -s -o /dev/null -w "%{http_code}" https://hatch.surf/

# SSL certificate
curl -v https://hatch.surf/ 2>&1 | grep -E "expire|issuer|subject"

# Static assets
curl -s -o /dev/null -w "%{http_code}" https://hatch.surf/style.css
curl -s -o /dev/null -w "%{http_code}" https://hatch.surf/brand/og/default.png
```

### Browser Verification

1. Visit https://hatch.surf
2. Verify no browser warnings (SSL valid)
3. Check that all assets load (CSS, images)
4. Test responsive design

## Troubleshooting

### Container Issues

```bash
# Check container status
docker ps -a | grep hatch-homepage

# View container logs
docker logs hatch-homepage

# Enter container
docker exec -it hatch-homepage sh

# Check nginx config inside container
docker exec hatch-homepage nginx -t
```

### Deployment Failures

1. Check GitHub Actions logs
2. Verify SSH access: `ssh root@46.250.250.48`
3. Check Docker status: `docker ps`
4. Check Caddy logs: `docker logs caddy`

### Missing Assets

1. Verify image contains files: `docker exec hatch-homepage ls -la /usr/share/nginx/html/`
2. Check nginx access logs: `docker logs hatch-homepage`
3. Ensure image is up to date: `docker pull hatch-homepage:latest`

## Maintenance

### Updating Content

1. Edit files in `site/` directory
2. Commit and push to `main` branch
3. GitHub Actions will auto-deploy to both locations

### SSL Renewal

Caddy handles TLS automatically via Let's Encrypt. No manual renewal needed.

### Backup

The site is version-controlled in Git. Docker images are stored on the server.

```bash
# Backup Docker image
docker save hatch-homepage:latest | gzip > hatch-homepage-backup.tar.gz

# Restore
docker load < hatch-homepage-backup.tar.gz
```

## Related Issues

- [ELF-248](/ELF/issues/ELF-248) - Properly Dockerize hatch.surf homepage deployment
- [ELF-192](/ELF/issues/ELF-192) - Build homepage and deploy it on hatch.surf
- [ELF-171](/ELF/issues/ELF-171) - Server setup (nginx, SSL)
- [ELF-196](/ELF/issues/ELF-196) - Deploy redesigned hatch.surf homepage (v2 files from ELF-195)
