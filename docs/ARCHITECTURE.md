# Hatch.surf Architecture

## Overview

Hatch.surf is the landing page and brand site for El Foundation. It consists of:
- **Static frontend** (site/) — HTML/CSS/JS served via nginx
- **Go backend** (cmd/hatch) — API server for future dynamic features

## Deployment Architecture

### Current State (Deprecated)

```
GitHub Push → GitHub Actions → rsync → /var/www/hatch.surf → nginx (host)
```

**Problems:**
- Direct filesystem deployment — no isolation
- Host-dependent — can't reproduce locally
- Secrets (SSH keys) required for rsync

### New Architecture

```
GitHub Push → GitHub Actions → Build Docker Image → SCP to Server → Docker Run → nginx reverse proxy
```

**Benefits:**
- Containerized — isolated from host
- Reproducible — same image in dev/prod
- No host filesystem dependency
- Easy rollback — just change image tag

## Docker Setup

### Frontend (site/)

**Dockerfile:**
```dockerfile
FROM nginx:alpine
RUN rm -rf /usr/share/nginx/html/*
COPY . /usr/share/nginx/html/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**docker-compose.yml:**
```yaml
services:
  hatch-homepage:
    build: .
    ports:
      - "127.0.0.1:3000:80"
    volumes:
      - .:/usr/share/nginx/html:ro
    restart: unless-stopped
```

### Backend (cmd/hatch)

**Dockerfile:**
```dockerfile
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/hatch ./cmd/hatch

FROM scratch
COPY --from=build /bin/hatch /bin/hatch
EXPOSE 8080
ENTRYPOINT ["/bin/hatch"]
```

**docker-compose.yml:**
```yaml
services:
  hatch:
    build: .
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - hatch-data:/data
    environment:
      - HATCH_PORT=8080
      - HATCH_DB_PATH=/data/hatch.db
      - HATCH_BASE_URL=${HATCH_BASE_URL:-http://localhost}

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy-data:/data
      - caddy-config:/config
    environment:
      - HATCH_HOSTNAME=${HATCH_HOSTNAME:-localhost}
    profiles:
      - with-caddy

volumes:
  hatch-data:
  caddy-data:
  caddy-config:
```

## Deployment Flow

### GitHub Actions (deploy-site.yml)

1. **Build** — `docker build -t hatch-homepage:latest ./site`
2. **Save** — `docker save hatch-homepage:latest | gzip > hatch-homepage.tar.gz`
3. **SCP** — Copy tarball to server
4. **Load** — `docker load < /tmp/hatch-homepage.tar.gz`
5. **Run** — `docker run -d --name hatch-homepage ...`

### Server Configuration

**Nginx reverse proxy** (host nginx):
```nginx
server {
    listen 80;
    server_name hatch.surf;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name hatch.surf;

    ssl_certificate /etc/letsencrypt/live/hatch.surf/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/hatch.surf/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Secrets Management

- **DEPLOY_HOST** — Server IP/hostname
- **DEPLOY_USER** — SSH user (nara)
- **DEPLOY_KEY** — SSH private key (stored in GitHub Secrets)

No secrets in repo. All deployment credentials in GitHub Secrets.

## Rollback Procedure

If deployment fails:
```bash
# On server
docker stop hatch-homepage
docker rm hatch-homepage
docker run -d --name hatch-homepage --restart unless-stopped -p 127.0.0.1:3000:80 hatch-homepage:previous-tag
```

## Future Improvements

1. **Container registry** — Push images to GitHub Container Registry (ghcr.io)
2. **Health checks** — Add HEALTHCHECK to Dockerfile
3. **Monitoring** — Add Prometheus metrics endpoint
4. **CDN** — Serve static assets via Cloudflare/CloudFront
5. **SSL automation** — Certbot in container or Caddy auto-SSL
