# ---- Base stage ----
FROM node:22-alpine AS base
WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@9 --activate

# Install deps (cache layer)
COPY apps/web/package.json apps/web/pnpm-lock.yaml apps/web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source
COPY apps/web .

# ---- Dev stage ----
FROM base AS dev
EXPOSE 5173
CMD ["pnpm", "dev", "--host"]

# ---- Build stage ----
FROM base AS builder
RUN CI=true pnpm build

# ---- Run stage ----
FROM nginx:alpine AS runner
COPY --from=builder /app/dist /usr/share/nginx/html
COPY infra/nginx/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
