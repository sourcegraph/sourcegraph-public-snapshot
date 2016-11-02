# Docker image for JavaScript/TypeScript lang/build server

FROM node:6.9-slim

ADD javascript-typescript-langserver/src /src
ADD javascript-typescript-langserver/build /build
ADD javascript-typescript-langserver/node_modules /node_modules

EXPOSE 2088
CMD ["node", "--max_old_space_size=4096", "/build/language-server.js", "-p", "2088", "--strict"]
