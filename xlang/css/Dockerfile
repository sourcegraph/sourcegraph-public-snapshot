# Docker image for CSS/LESS/SCSS lang/build server

FROM node:6.9-slim

COPY css-langserver /css-langserver

EXPOSE 9001
CMD ["node", "--max_old_space_size=4096", "/css-langserver/langserver/build/language-server.js", "--lsport", "9001"]
