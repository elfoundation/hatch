# Hatch Landing Page — Deployment Guide

## Overview

The Hatch landing page (`hatch.surf`) is a **static HTML** site deployed via **Docker** on the production server.

**Key principle**: All changes MUST go through the git repository and automated deployment. Never edit files directly on the server or use `/var/www`.

---

## Repository

**Location**: `/home/nara/apps/web/`

**Git Remotes**:
- `github` → `elfoundation/hatch` (source of truth)
- `origin` → `el-foundation/hatch-surf` (Gitea push-mirror, deploys to hatch.surf)

**Structure**:
```
site/                   # Static files served by nginx
  index.html            # Main landing page
  style.css             # Styles
  main.js               # Anime.js animations
Dockerfile              # Docker build configuration (static files only)
docker-compose.yml      # Container orchestration
deploy.sh               # Deployment script
```

---

## Deployment Flow

### Source of Truth

**GitHub `main`** is the source of truth. Gitea is a push-mirror.

All development happens via PRs to `elfoundation/hatch`. The deploy script pulls from GitHub and pushes to Gitea.

### Automatic (Recommended)

1. Merge PR to GitHub `main`
2. Gitea webhook triggers `gitea-webhook.py`
3. Script runs `deploy.sh` automatically
4. `deploy.sh` pulls from GitHub, pushes to Gitea (push-mirror), rebuilds Docker, restarts

### Manual

```bash
cd /home/nara/apps/web
./deploy.sh
```

**What `deploy.sh` does**:
1. `git pull github main` — fetches latest from GitHub (source of truth)
2. `git push origin main` — syncs to Gitea (push-mirror)
3. `cd site && docker build -t hatch-web:latest .` — rebuilds image
4. Stops and removes old container
5. `docker run -d` — starts new container
6. Verifies deployment with health check

---

## Docker Setup

**Container**: `hatch-web`  
**Port**: `8080` (mapped to host)  
**Internal**: nginx serving static files on port 80  
**Network**: `hatch-network`

**Check status**:
```bash
docker ps | grep hatch-web
docker logs hatch-web
```

**Restart without rebuild**:
```bash
docker compose -f /home/nara/apps/web/docker-compose.yml restart
```

---

## Making Changes

### 1. Edit the static files

```bash
cd /home/nara/apps/web

# Edit landing page
vim out/index.html

# Edit styles
vim out/style.css

# Edit animations
vim out/main.js
```

### 2. Test locally (optional)

```bash
# Serve files locally
cd out && python3 -m http.server 3000
# Visit http://localhost:3000
```

### 3. Commit and push

```bash
git add .
git commit -m "Description of changes"
git push origin main
```

### 4. Deploy

Either wait for webhook or run manually:
```bash
./deploy.sh
```

### 5. Verify

```bash
curl -s http://localhost:8080 | head -20
# Or visit https://hatch.surf
```

---

## Common Mistakes to Avoid

❌ **DON'T**: Edit files in `/var/www/html/`  
❌ **DON'T**: Edit files directly inside the Docker container  
❌ **DON'T**: Work on static HTML/CSS files in a workspace without committing  
❌ **DON'T**: Copy files to the server without using git  

✅ **DO**: Edit files in `/home/nara/apps/web/out/`  
✅ **DO**: Commit changes to git  
✅ **DO**: Use `deploy.sh` for deployment  
✅ **DO**: Test locally before pushing  

---

## Rollback

If a deployment causes issues:

```bash
cd /home/nara/apps/web

# View recent commits
git log --oneline -10

# Revert to previous version
git revert HEAD
git push origin main

# Redeploy
./deploy.sh
```

Previous builds are backed up in `out.backup.*` directories.

---

## Cache Headers

**HTML pages**: `no-cache, must-revalidate` — browsers always revalidate, ensuring fresh content after deploys.

**Static assets** (CSS/JS/images): `public, immutable` with 1-year expiry.

**Nginx config**: Built into the Dockerfile.

If users report seeing old content after deployment:
1. Verify the container is running: `docker ps | grep hatch-web`
2. Check the HTML has new content: `curl -s http://localhost:8080 | head -5`
3. Ask user to hard refresh: Ctrl+Shift+R (Chrome/Firefox) or Cmd+Shift+R (Mac)

---

## Troubleshooting

**Container not starting**:
```bash
docker logs hatch-web
docker compose -f /home/nara/apps/web/docker-compose.yml up -d
```

**Changes not showing**:
1. Verify commit pushed: `git log --oneline -1`
2. Check container was rebuilt: `docker images | grep hatch-web`
3. Force rebuild: `cd /home/nara/apps/web && docker build -t hatch-web:latest . && docker compose up -d`

**Port conflict**:
```bash
lsof -i :8080
# Kill conflicting process or change port in docker-compose.yml
```

---

## Contact

- **DevOps**: Jing Yang (Sam Lee)
- **CTO**: Jordan Patel
- **Repository**: `/home/nara/apps/web/`
- **Live site**: https://hatch.surf
