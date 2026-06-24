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
- Deploy to hatch.surf server via rsync

### 2. Manual (When on Server)

```bash
# Copy files
sudo cp site/index.html /var/www/hatch.surf/index.html
sudo cp site/style.css /var/www/hatch.surf/style.css
sudo cp site/main.js /var/www/hatch.surf/main.js

# Set permissions
sudo chown www-data:www-data /var/www/hatch.surf/{index,style,main}.{html,css,js}

# Verify
curl -s -I https://hatch.surf
```

### 3. Via SSH (From Remote Machine)

```bash
./scripts/deploy-site.sh
```

## Post-Deployment Verification

- [ ] https://hatch.surf loads (HTTP 200)
- [ ] Three.js particle animation renders (canvas element present)
- [ ] Scroll animations work (data-animate attributes)
- [ ] All static assets load:
  - [ ] style.css
  - [ ] main.js
  - [ ] brand/og/default.png
  - [ ] Google Fonts (Inter, JetBrains Mono)
- [ ] Mobile viewport works (390px width)
- [ ] No console errors

## Troubleshooting

### Site Not Loading
```bash
# Check nginx status
sudo systemctl status nginx

# Check nginx config
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

### Assets Not Loading
```bash
# Check file permissions
ls -la /var/www/hatch.surf/

# Fix permissions
sudo chown -R www-data:www-data /var/www/hatch.surf/
```

### SSL Issues
```bash
# Check certificate
sudo certbot certificates

# Renew if needed
sudo certbot renew --cert-name hatch.surf
```

## Related Documentation

- [Full Deployment Guide](deploy-homepage.md)
- [Local Development](local-dev.md)
- [Hatch Architecture](hatch-architecture.md)
