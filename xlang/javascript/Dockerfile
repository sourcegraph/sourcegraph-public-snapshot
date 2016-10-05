# Docker image for JavaScript/TypeScript lang/build server

FROM node:5.7.0-slim

RUN apt-get update -y && apt-get install -y git
RUN npm install typescript -g

ADD langserver/src /src
ADD langserver/build /build
ADD langserver/node_modules /node_modules

EXPOSE 2088
CMD ["node", "/build/language-server.js", "-p", "2088", "--strict"]
