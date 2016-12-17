# Docker image for JavaScript/TypeScript lang/build server

FROM node:6.9-slim

ADD buildserver/src /src
ADD buildserver/build /build
ADD buildserver/node_modules /node_modules

EXPOSE 2088
CMD ["node", "--max_old_space_size=4096", "/build/src/language-server.js", "-p", "2088", "--strict"]
