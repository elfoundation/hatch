# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Build the application
COPY . .
RUN npm run build

# Stage 2: Production
FROM nginx:alpine AS production

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy static export (everything is in /app/out)
COPY --from=builder /app/out /usr/share/nginx/html

# Expose port (internal only, no SSL)
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:80/ || exit 1
