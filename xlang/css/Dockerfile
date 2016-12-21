# Docker image for JavaScript/TypeScript lang/build server

FROM node:6.9-slim

ADD css-langserver /css-langserver

EXPOSE 1815
CMD ["node", "--max_old_space_size=4096", "/css-langserver/langserver/bin/css-langserver", "--socket", "1815", "--strict"]
