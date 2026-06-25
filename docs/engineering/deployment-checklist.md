# Deployment Checklist

Quick reference for deploying the hatch.surf homepage.

## Pre-Deployment

- [ ] Changes are in `site/` directory
- [ ] All assets load locally
- [ ] No console errors in browser
- [ ] Mobile responsive design works

## Deployment Methods

### 1. Automated (Recommended)

Push to `main` branch:
```bash
git push origin main
```

GitHub Actions will automatically:
- Deploy to GitHub Pages
- Build Docker image and deploy to hatch.surf server

### 2. Manual (From Remote Machine)

```bash
# Dry run first
./scripts/deploy-site.sh --dry-run

# Deploy
./scripts/deploy-site.sh
```

### 3. Docker Commands (On Server)

```bash
# Build image
docker build -t hatch-homepage:latest ./site

# Run container
docker run -d \
  --name hatch-homepage \
  --restart unless-stopped \
  -p 127.0.0.1:3000:80 \
  hatch-homepage:latest

# Check status
docker ps | grep hatch-homepage
```

## Post-Deployment Verification

- [ ] https://hatch.surf loads (HTTP 200)
- [ ] Anime.js magical background renders
- [ ] Scroll animations work (data-animate attributes)
- [ ] All static assets load:
  - [ ] style.css
  - [ ] main.js
  - [ ] brand/og/default.png
  - [ ] Google Fonts (Inter, JetBrains Mono)
- [ ] Mobile viewport works (390px width)
- [ ] No console errors

## Troubleshooting

### Container Not Running
```bash
# Check container status
docker ps -a | grep hatch-homepage

# View logs
docker logs hatch-homepage

# Restart container
docker restart hatch-homepage
```

### Assets Not Loading
```bash
# Check files in container
docker exec hatch-homepage ls -la /usr/share/nginx/html/

# Rebuild image
docker build -t hatch-homepage:latest ./site
docker restart hatch-homepage
```

### SSL Issues
Caddy handles TLS automatically. Check Caddy status:
```bash
docker ps | grep caddy
docker logs caddy
```

## Related Documentation

- [Full Deployment Guide](deploy-homepage.md)
- [Local Development](local-dev.md)
- [Hatch Architecture](hatch-architecture.md)
