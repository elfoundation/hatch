# Hatch Static Site

This directory contains the static site for [hatch.surf](https://hatch.surf).

## Structure

```
site/
├── index.html              # Landing page
├── style.css               # Global styles
├── blog/
│   ├── index.html          # Blog index
│   └── why-we-are-building-hatch/
│       └── index.html      # Blog post
└── brand/
    ├── og/
    │   └── default.png     # Open Graph image
    └── favicon/
        ├── favicon.ico
        └── favicon-*.png   # Various sizes
```

## Deployment

The site is deployed via Docker to the hatch.surf server. Static files are baked into the Docker image — no host directory mounting required.

### Docker Deployment

The GitHub Actions workflow (`deploy-site.yml`) automatically:
1. Builds a Docker image with static files baked in
2. Pushes the image to the server
3. Restarts the container

### Manual Deployment

```bash
# From repo root
./scripts/deploy-site.sh

# Dry run
./scripts/deploy-site.sh --dry-run
```

### DNS Configuration

To point `hatch.surf` to your server:

1. Add A record pointing to your server IP
2. Enable HTTPS via Caddy or Let's Encrypt

### Architecture

The deployment uses:
- **Docker image**: nginx:alpine with static files baked in
- **No host mounts**: Image is self-contained, no `/var/www` dependency
- **Caddy**: Optional reverse proxy for TLS termination
- **Port 3000**: Internal port, proxied by Caddy on 80/443

## Local Development

To preview the site locally:

```bash
cd site
python3 -m http.server 8000
# Open http://localhost:8000
```

Or with Node.js:

```bash
npx serve site
```

### Docker

To run the site in Docker:

```bash
cd site
# Build and run with docker compose
docker compose up -d

# Or build and run with docker
docker build -t hatch-homepage .
docker run -d -p 3000:80 hatch-homepage
```

The site will be available at http://localhost:3000

To stop the container:

```bash
docker compose down
# Or
docker stop $(docker ps -q --filter ancestor=hatch-homepage)
```

## SEO Checklist

- [x] H1 = title
- [x] Meta description ≤ 160 chars
- [x] Primary keyword in first 200 words
- [x] OG image set
- [x] Internal link to repo
- [x] Canonical URL set
- [x] Semantic HTML
- [x] Mobile responsive
- [x] Fast loading (static HTML/CSS)

## Adding New Posts

1. Create a new directory under `blog/`
2. Create an `index.html` file with the post content
3. Update `blog/index.html` to include the new post
4. Follow the same HTML structure as existing posts