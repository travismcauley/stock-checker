# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./
RUN npm ci

# Copy source code
COPY . .

# Build args for environment variables at build time
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

# Build the app
RUN npm run build

# Runtime stage
FROM nginx:alpine

# Copy nginx config template
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# Railway sets PORT env var - default to 80 for local
ENV PORT=80

# Expose port (Railway will override this)
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
