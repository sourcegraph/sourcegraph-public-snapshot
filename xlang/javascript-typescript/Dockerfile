# Docker image for JavaScript/TypeScript lang/build server

FROM node:7-alpine

# Allow to mount a custom yarn config
RUN mkdir /yarn-config \
	&& touch /yarn-config/.yarnrc \
	&& ln -s /yarn-config/.yarnrc /usr/local/share/.yarnrc

# Add git, needed for yarn
RUN apk add --no-cache bash git openssh

ADD buildserver/lib /buildserver/lib
ADD buildserver/node_modules /buildserver/node_modules
ADD buildserver/package.json /buildserver/package.json

WORKDIR /buildserver
EXPOSE 2088
CMD ["node", "--max_old_space_size=4096", "./lib/language-server.js", "-p", "2088", "--strict", "--cluster", "8", "--color"]
