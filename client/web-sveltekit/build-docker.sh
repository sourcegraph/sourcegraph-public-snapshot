# Prepares and creates a standalone docker image for demoing the prototype
pnpm run build
docker build . -t fkling/web-sveltekit
