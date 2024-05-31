FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY . /app
WORKDIR /app

FROM base AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile

FROM base AS build
RUN apt update
# This is required for the post-install check of puppeteer. If we can remove puppeteer, we can remove this
RUN apt install chromium -y
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM base
COPY --from=prod-deps /app/node_modules /app/node_modules
RUN ls -la /app/web
RUN ls -la /app/web-sveltekit
COPY --from=build /app/web/dist /app/dist/web
COPY --from=build /app/web-sveltekit/build /app/dist/web-sveltekit
RUN tar -cvzf assets.tar.gz /app/dist
