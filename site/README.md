# Hatch Static Site

This directory contains the static site for [hatch.sh](https://hatch.sh).

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

The site is automatically deployed to GitHub Pages via the `deploy-site.yml` workflow when changes are pushed to the `main` branch.

### GitHub Pages Setup

1. Go to repository Settings → Pages
2. Source: Deploy from a branch
3. Branch: `main`, folder: `/site`
4. Custom domain: `hatch.sh`

### DNS Configuration

To point `hatch.sh` to GitHub Pages:

1. Add these DNS records:
   - Type: A, Name: @, Value: 185.199.108.153
   - Type: A, Name: @, Value: 185.199.109.153
   - Type: A, Name: @, Value: 185.199.110.153
   - Type: A, Name: @, Value: 185.199.111.153
   - Type: AAAA, Name: @, Value: 2606:50c0:8000::153
   - Type: AAAA, Name: @, Value: 2606:50c0:8001::153
   - Type: AAAA, Name: @, Value: 2606:50c0:8002::153
   - Type: AAAA, Name: @, Value: 2606:50c0:8003::153

2. Enable HTTPS in GitHub Pages settings

### 301 Redirect from GitHub README

Add this to the README.md or as a GitHub Pages redirect:

```html
<!-- In site/index.html or a dedicated redirect page -->
<meta http-equiv="refresh" content="0; url=/blog/why-we-are-building-hatch/">
```

Or use a JavaScript redirect for better SEO:

```javascript
// In a script tag or separate JS file
window.location.replace('/blog/why-we-are-building-hatch/');
```

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