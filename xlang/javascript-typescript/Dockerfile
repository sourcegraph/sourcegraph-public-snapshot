# Docker image for JavaScript/TypeScript lang/build server

FROM node:6.9-slim

ADD buildserver/src /buildserver/src
ADD buildserver/lib /buildserver/lib
ADD buildserver/node_modules /buildserver/node_modules
ADD buildserver/package.json /buildserver/package.json

WORKDIR /buildserver
EXPOSE 2088
CMD ["node", "--max_old_space_size=4096", "./lib/language-server.js", "-p", "2088", "--strict", "--cluster", "8"]
