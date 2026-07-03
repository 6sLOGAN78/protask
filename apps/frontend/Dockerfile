# Multi-stage build for React Frontend
FROM oven/bun:1.2.13-alpine AS builder

WORKDIR /app

# Copy root workspace configurations
COPY package.json bun.lock turbo.json ./

# Copy packages and apps configs
COPY packages/zod/package.json ./packages/zod/
COPY packages/openapi/package.json ./packages/openapi/
COPY packages/emails/package.json ./packages/emails/
COPY apps/frontend/package.json ./apps/frontend/

# Install dependencies (workspaces will be linked)
RUN bun install --frozen-lockfile

# Copy the actual packages source code
COPY packages/zod ./packages/zod
COPY packages/openapi ./packages/openapi

# Build shared packages first
RUN bun run --filter=@protask/zod build
RUN bun run --filter=@protask/openapi build

# Copy frontend source code
COPY apps/frontend ./apps/frontend

# Build frontend
# We accept ARG for build-time config
ARG VITE_CLERK_PUBLISHABLE_KEY
ARG VITE_API_URL
ARG VITE_ENV=production

ENV VITE_CLERK_PUBLISHABLE_KEY=$VITE_CLERK_PUBLISHABLE_KEY
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_ENV=$VITE_ENV

RUN bun run --filter=frontend build

# Deploy inside Nginx
FROM nginx:1.25-alpine

COPY --from=builder /app/apps/frontend/dist /usr/share/nginx/html
# Copy custom nginx config
COPY apps/frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
