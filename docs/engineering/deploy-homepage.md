# Deploying the Hatch Homepage

This document describes how the Hatch static homepage is deployed to hatch.surf.

## Overview

The homepage is deployed to two locations:

1. **GitHub Pages** - Serves as a fallback and for GitHub-based discovery
2. **hatch.surf server** - Primary deployment at https://hatch.surf

## Architecture

```
site/
├── index.html          # Main homepage
├── style.css           # Stylesheet
├── brand/              # Logo and favicon assets
│   ├── favicon/
│   └── og/
└── blog/               # Blog directory (placeholder)
```

## Server Setup

### Prerequisites

- Server: 46.250.250.48 (hatch.surf)
- nginx installed and configured
- Let's Encrypt certificate for hatch.surf
- SSH access for deployment

### Nginx Configuration

The nginx server block is located at:
- `/etc/nginx/sites-available/hatch.surf.conf`
- Symlinked to `/etc/nginx/sites-enabled/hatch.surf.conf`

Configuration highlights:
- HTTP → HTTPS redirect
- SSL with Let's Encrypt certificate
- Security headers (X-Content-Type-Options, X-Frame-Options, Referrer-Policy)
- Gzip compression
- Static asset caching (30 days)

### SSL Certificate

- Certificate managed by Let's Encrypt
- Auto-renewal via certbot
- Expiry: September 22, 2026

## Deployment

### Automated (GitHub Actions)

When changes are pushed to `main` branch affecting `site/` directory:

1. **GitHub Pages** - Deployed automatically via GitHub Actions
2. **Server** - Deployed via rsync to `/var/www/hatch.surf/`

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

### Direct rsync

```bash
rsync -avz --delete site/ root@46.250.250.48:/var/www/hatch.surf/
ssh root@46.250.250.48 "nginx -t && systemctl reload nginx"
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

### SSL Certificate Issues

```bash
# Check certificate status
sudo certbot certificates

# Renew certificate manually
sudo certbot renew --cert-name hatch.surf

# Test nginx config
sudo nginx -t
```

### Deployment Failures

1. Check GitHub Actions logs
2. Verify SSH access: `ssh root@46.250.250.48`
3. Check nginx status: `sudo systemctl status nginx`
4. Check nginx logs: `sudo tail -f /var/log/nginx/error.log`

### Missing Assets

1. Verify files exist on server: `ls -la /var/www/hatch.surf/brand/`
2. Check nginx access logs: `sudo tail -f /var/log/nginx/access.log`
3. Ensure proper file permissions: `chown -R www-data:www-data /var/www/hatch.surf`

## Maintenance

### Updating Content

1. Edit files in `site/` directory
2. Commit and push to `main` branch
3. GitHub Actions will auto-deploy to both locations

### SSL Renewal

Certbot is configured to auto-renew. To check renewal status:

```bash
sudo certbot renew --dry-run
```

### Backup

The site is version-controlled in Git. For server backups:

```bash
# Backup nginx config
sudo tar -czf nginx-hatch-backup.tar.gz /etc/nginx/sites-available/hatch.surf.conf /etc/letsencrypt/live/hatch.surf/
```

## Related Issues

- [ELF-192](/ELF/issues/ELF-192) - Build homepage and deploy it on hatch.surf
- [ELF-171](/ELF/issues/ELF-171) - Server setup (nginx, SSL)
