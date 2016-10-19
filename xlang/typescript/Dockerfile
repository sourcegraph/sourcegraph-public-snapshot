# Docker image for JavaScript/TypeScript lang/build server

FROM node:6.9-slim

ADD langserver-typescript/src /src
ADD langserver-typescript/build /build
ADD langserver-typescript/node_modules /node_modules

EXPOSE 2088
CMD ["node", "/build/language-server.js", "-p", "2088", "--strict"]
