# ==========================================
# Digital Papyrus Frontend - Astro Static
# ==========================================
FROM node:lts-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm ci

# Copy source code and build
COPY . .
RUN npm run build

# Serve via NGINX
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Setup custom config for SPA/Static handling
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy static assets from Astro build to nginx html folder
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port (must match docker-compose internal port)
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]