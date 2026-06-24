# Dockerfile for hatch.surf homepage
# Uses nginx:alpine to serve static files

FROM nginx:alpine

# Remove default nginx static assets
RUN rm -rf /usr/share/nginx/html/*

# Copy static files from site directory
COPY . /usr/share/nginx/html/

# Copy custom nginx configuration for SPA routing if needed
# COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]