# This image creates a standalone web server for demoing the prototype
FROM node:16

WORKDIR /usr/src/app
RUN echo '{"type": "module"}' > package.json
RUN npm install express http-proxy-middleware
COPY  server.js ./
COPY build ./build

ENV SOURCEGRAPH_API_HOST=https://sourcegraph.com
EXPOSE 4173
CMD ["node", "server.js"]
